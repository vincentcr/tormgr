package main

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       RecordID `json:"id"`
	Email    string   `json:"email"`
	password string
}

func (user User) String() string {
	return fmt.Sprintf("User[%s, email:%s]", user.ID, user.Email)
}

const bcryptCost = 8

func UserCreate(email string, password string) (User, error) {
	hashedPassword, err := userHashPassword(password)
	if err != nil {
		return User{}, err
	}

	user := User{
		ID:       newID(),
		Email:    userNormalizeEmail(email),
		password: hashedPassword,
	}

	_, err = services.db.Exec("INSERT INTO users(id,email,password) VALUES($1,$2,$3)", user.ID, user.Email, user.password)
	if err != nil {
		if dbIsUniqueError(err) {
			return User{}, ErrUniqueViolation
		} else {
			return User{}, fmt.Errorf("unable to create user %#v: %v", user, err)
		}
	}

	return user, nil
}

func userHashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("unable to hash password of length %v with cost %v: %v", len(password), bcryptCost, err)
	}
	return string(hash), nil
}

func userNormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func UserGetByID(id RecordID) (User, error) {
	user := User{ID: id}
	err := services.db.
		QueryRow("SELECT email,password FROM users WHERE id=$1", id).
		Scan(&user.Email, &user.password)
	if err == sql.ErrNoRows {
		return User{}, ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("Error fetching user %v: %v", id, err)
	} else {
		return user, ErrNotFound
	}
}

func UserAuthenticateWithPassword(email string, password string) (User, error) {
	user := User{Email: userNormalizeEmail(email)}
	err := services.db.
		QueryRow("SELECT id,password FROM users WHERE email=$1", user.Email).
		Scan(&user.ID, &user.password)
	if err == sql.ErrNoRows {
		return User{}, ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("Error fetching user %v: %v", email, err)
	} else if userVerifyPassword(password, user) {
		return user, nil
	} else {
		return User{}, ErrNotFound
	}
}

func userVerifyPassword(password string, user User) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password)) == nil
}
