package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"
)

const tokenKeyFormat = "token.%s"
const tokenListKeyFormat = "tokenlist.%s"
const tokenSecretSize = 32
const tokenDeleteExpiredInterval = 1 * time.Hour
const maxDuration time.Duration = 1<<63 - 1

type Access int

const (
	AccessRead      Access = 1 << iota
	AccessWrite            = 1 << iota
	AccessReadWrite        = AccessRead | AccessWrite
	AccessNone             = 0
)

type AccessToken struct {
	Secret  string
	UserID  RecordID
	Access  Access
	Expires *time.Time
}

type tokenOptions struct {
	duration   time.Duration
	access     Access
	secretSize int
}

func (access *Access) Parse(str string) error {
	switch str {
	case "Read":
		*access = AccessRead
	case "Write":
		*access = AccessWrite
	case "":
		fallthrough
	case "ReadWrite":
		*access = AccessReadWrite
	default:
		return fmt.Errorf("Invalid access string %s", str)
	}
	return nil
}

func AccessTokenCreateFull(user User) (string, error) {
	return AccessTokenCreate(user, AccessReadWrite)
}

func AccessTokenCreate(user User, access Access) (string, error) {
	duration := maxDuration

	token, err := tokenDbCreate(user, access, duration)
	if err != nil {
		return "", err
	}

	err = tokenAddToCache(user, token)

	return token.Secret, nil
}

func tokenDbCreate(user User, access Access, duration time.Duration) (AccessToken, error) {
	secret, err := tokenGenerateSecret(user.ID, tokenSecretSize)
	if err != nil {
		return AccessToken{}, err
	}

	token := AccessToken{
		Secret:  secret,
		UserID:  user.ID,
		Access:  access,
		Expires: tokenMkExpires(duration),
	}

	_, err = services.db.Exec("INSERT INTO access_tokens(secret, user_id, access, expires) VALUES($1,$2,$3,$4)",
		token.Secret, token.UserID, token.Access, token.Expires)
	if err != nil {
		return AccessToken{}, fmt.Errorf("Failed to insert token %v into db: %v", token, err)
	}
	return token, nil
}

func tokenMkExpires(duration time.Duration) *time.Time {
	if duration == maxDuration {
		return nil
	} else {
		expires := time.Now().Add(duration)
		return &expires
	}
}

func tokenGenerateSecret(userID RecordID, size int) (string, error) {
	secretOffset := len(userID) + 1
	buf := make([]byte, secretOffset+size)
	copy(buf, userID+":")
	nRandBytes, err := rand.Read(buf[secretOffset:])
	if err != nil {
		return "", fmt.Errorf("unable to generate %v bytes of randomness: %v", size, err)
	} else if nRandBytes < size {
		return "", fmt.Errorf("got %d bytes from rand instead of requested %v", nRandBytes, size)
	}

	encoded := base64.URLEncoding.EncodeToString(buf)

	//URLEncode still might contain the '=', which is not very URL-friendly.
	encodedForURL := strings.Replace(encoded, "=", "", -1)

	return encodedForURL, nil
}

func tokenAddToCache(user User, token AccessToken) error {
	userJson, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to json-encode user %v: %v", user, err)
	}

	key := fmt.Sprintf(tokenKeyFormat, token.Secret)
	accessStr := strconv.Itoa(int(token.Access))
	err = services.redis.HMSet(key, "user", string(userJson), "access", accessStr).Err()
	if err != nil {
		return fmt.Errorf("redis.HMSet(%v, 'user', %s, 'access', %v) failed: %v", key, userJson, accessStr, err)
	}
	if token.Expires != nil {
		err = services.redis.ExpireAt(key, *token.Expires).Err()
		if err != nil {
			return fmt.Errorf("redis.ExpireAt(%v, %v) failed: %v", key, token.Expires, err)
		}

	}

	tokenListKey := fmt.Sprintf(tokenListKeyFormat, user.ID)
	err = services.redis.SAdd(tokenListKey, key).Err()
	if err != nil {
		return fmt.Errorf("redis.Sadd(%v, %s) failed: %v", tokenListKey, key)
	}
	return nil
}

func AccessTokenAuthenticateUser(secret string) (User, Access, error) {
	user, access, err := tokenGetFromCache(secret)
	if err != ErrNotFound {
		return user, access, err
	}

	user, token, err := tokenGetFromDB(secret)
	if err == nil {
		tokenAddToCache(user, token)
	}

	return user, token.Access, err
}

func tokenGetFromDB(secret string) (User, AccessToken, error) {
	user := User{}
	token := AccessToken{Secret: secret}

	query := `select users.id, users.email, access_tokens.access, access_tokens.expires
		FROM access_tokens INNER JOIN users ON users.id = access_tokens.user_id
		WHERE access_token_is_valid($1, access_tokens.*)`
	err := services.db.
		QueryRow(query, secret).
		Scan(&user.ID, &user.Email, &token.Access, &token.Expires)
	if err == sql.ErrNoRows {
		return User{}, AccessToken{}, ErrNotFound
	} else if err != nil {
		return User{}, AccessToken{}, fmt.Errorf("Error fetching user from token %v: %v", token, err)
	} else {
		token.UserID = user.ID
		return user, token, nil
	}

}

func tokenGetFromCache(token string) (User, Access, error) {
	key := fmt.Sprintf(tokenKeyFormat, token)
	data, err := services.redis.HMGet(key, "user", "access").Result()
	if err == redis.Nil || data[0] == nil {
		return User{}, AccessNone, ErrNotFound
	} else if err != nil {
		return User{}, AccessNone, fmt.Errorf("unable to get key %v: %v", token, err)
	}

	user := User{}
	err = json.Unmarshal([]byte(data[0].(string)), &user)
	if err != nil {
		return User{}, AccessNone, fmt.Errorf("unable to unmarshall user from json %v: %v", data[0], err)
	}

	access, err := strconv.Atoi(data[1].(string))
	if err != nil {
		return User{}, AccessNone, fmt.Errorf("unable to convert access string %v to an int: %v", data[1], err)
	}

	return user, Access(access), nil
}

func AccessTokenDelete(user User, token string) error {
	err := tokenDeleteFromCache(user, token)
	if err != nil {
		return err
	}

	return tokenDeleteFromDB(user, token)
}

func tokenDeleteFromCache(user User, token string) error {
	key := fmt.Sprintf(tokenKeyFormat, token)
	err := services.redis.Del(key).Err()
	if err != nil {
		return fmt.Errorf("unable to delete key %v: %v", key, err)
	}

	tokenListKey := fmt.Sprintf(tokenListKeyFormat, user.ID)
	err = services.redis.SRem(tokenListKey, key).Err()
	if err != nil {
		return fmt.Errorf("unable to remove entry %v from set %v: %v", key, tokenListKey, err)
	}

	return nil
}

func tokenDeleteFromDB(user User, secret string) error {
	_, err := services.db.Exec("DELETE FROM access_tokens WHERE secret = $1 AND user_id = $2", secret, user.ID)
	if err != nil {
		return fmt.Errorf("failed to delete token %v from db: %v", secret, err)
	}
	return nil
}

func AccessTokenDeleteAll(userID RecordID) error {
	err := tokenDeleteAllFromCache(userID)
	if err != nil {
		return err
	}

	return tokenDeleteAllFromDB(userID)
}

func tokenDeleteAllFromCache(userID RecordID) error {
	tokenListKey := fmt.Sprintf(tokenListKeyFormat, userID)
	keys, err := services.redis.SMembers(tokenListKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("unable to get members of set %v: %v", tokenListKey, err)
	}

	err = services.redis.Del(keys...).Err()
	if err != nil {
		return fmt.Errorf("unable to delete keys %v: %v", keys, err)
	}

	return nil
}

func tokenDeleteAllFromDB(userID RecordID) error {
	_, err := services.db.Exec("DELETE FROM access_tokens WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete all tokens of user %v from db: %v", userID, err)
	}
	return nil
}

func tokensStartDeleteExpiredLoop() {
	go func() {
		tick := time.Tick(tokenDeleteExpiredInterval)
		for range tick {
			if err := tokensDeleteExpired(); err != nil {
				log.Println("[tokensStartDeleteExpiredLoop]", err)
			}
		}
	}()
}

func tokensDeleteExpired() error {
	_, err := services.db.Exec("DELETE FROM access_tokens WHERE expires IS NOT NULL AND expires < NOW();")
	if err != nil {
		return err
	}
	return nil
}
