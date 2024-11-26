package main

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", NewErr(ErrInternal, err, "failed to create password hash")
	}
	return string(hashed), err
}

func VerifyPassword(password, hashed string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		return NewErr(ErrInput, err, "password does not match")
	}
	return nil
}
