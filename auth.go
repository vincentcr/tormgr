package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func (c *TMContext) GetUserAccess() Access {
	access, ok := c.Env["userAccess"]
	if !ok {
		panic("no access present in context env")
	}
	return access.(Access)
}

func mustAuthenticateRW(h handler) handler {
	return mustAuthenticate(AccessReadWrite, h)
}

func mustAuthenticateR(h handler) handler {
	return mustAuthenticate(AccessRead, h)
}

func mustAuthenticate(access Access, h handler) handler {
	return func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		_, ok := c.GetUser()
		if !ok {
			w.Header().Set("WWW-Authenticate", "Basic realm=\"Please enter your username and password\"")
			return NewHttpError(http.StatusUnauthorized)
		}
		h(c, w, r)
		return nil
	}
}

type AuthMethod string
type AuthCreds []string

const (
	AuthMethodNone   AuthMethod = ""
	AuthMethodBasic  AuthMethod = "Basic"
	AuthMethodBearer AuthMethod = "Bearer"
	AuthMethodToken  AuthMethod = "Token"
)

func authenticate(c *TMContext, w http.ResponseWriter, r *http.Request) error {
	method, creds, err := parseAuthorizationFromRequest(r)
	if err != nil {
		return err
	}

	if method != AuthMethodNone {
		user, access, err := verifyCredentials(c, method, creds)
		if err == ErrNotFound {
			return NewHttpErrorWithText(http.StatusUnauthorized, "Invalid Credentials")
		} else if err != nil {
			return err
		}

		log.Printf("authenticated as %v", user)
		c.Env["user"] = user
		c.Env["userAccess"] = access

	}

	return nil
}

type credentialParser func(r *http.Request) (AuthMethod, AuthCreds, error)

var credentialParsers = []credentialParser{parseAuthorizationFromHeader, parseAuthorizationFromForm}

func parseAuthorizationFromRequest(r *http.Request) (AuthMethod, AuthCreds, error) {

	for _, parser := range credentialParsers {
		method, creds, err := parser(r)
		if method != AuthMethodNone || err != nil {
			return method, creds, err
		}
	}
	return AuthMethodNone, nil, nil
}

func verifyCredentials(c *TMContext, method AuthMethod, creds AuthCreds) (User, Access, error) {
	switch method {
	case AuthMethodBasic:
		username := creds[0]
		password := creds[1]
		user, err := UserAuthenticateWithPassword(username, password)
		return user, AccessReadWrite, err
	case AuthMethodBearer:
		fallthrough
	case AuthMethodToken:
		token := creds[0]
		return AccessTokenAuthenticateUser(token)
	default:
		return User{}, AccessNone, NewHttpErrorWithText(http.StatusBadRequest, fmt.Sprintf("Unknown auth method %s", method))
	}

}

func parseAuthorizationFromHeader(r *http.Request) (AuthMethod, AuthCreds, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return AuthMethodNone, nil, nil
	}

	match := regexp.MustCompile("^(.+?)\\s+(.+)$").FindStringSubmatch(header)
	if len(match) == 0 {
		return "", nil, NewHttpErrorWithText(http.StatusBadRequest, "Invalid auth header")
	}

	method := AuthMethod(match[1])
	encodedCreds := match[2]
	var creds AuthCreds

	if method == AuthMethodBasic {
		userPasswordStr, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			return "", nil, NewHttpErrorWithText(http.StatusBadRequest, "Invalid basic auth header: not base64")
		}

		creds = strings.Split(string(userPasswordStr), ":")

	} else if method == AuthMethodToken {
		tokenSecretMatch := regexp.MustCompile("token=\"(.+?)\".*").FindStringSubmatch(encodedCreds)
		if len(tokenSecretMatch) == 0 {
			return "", nil, NewHttpErrorWithText(http.StatusBadRequest, "Invalid auth token header")
		}

		creds = tokenSecretMatch[1:2]
	} else if method == AuthMethodBearer {
		creds = []string{encodedCreds}
	}

	return method, creds, nil
}

func parseAuthorizationFromForm(r *http.Request) (AuthMethod, AuthCreds, error) {
	token := r.FormValue("_tok")
	if token != "" {
		return AuthMethodToken, []string{token}, nil
	}
	return AuthMethodNone, nil, nil
}
