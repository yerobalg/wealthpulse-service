package cryptolib

import (
	"golang.org/x/crypto/bcrypt"
)

type passwordLib struct {
	saltRound int
}

type PasswordInterface interface {
	Hash(string) (string, error)
	Compare(string, string) bool
}

func InitPassword(saltRound int) PasswordInterface {
	return &passwordLib{
		saltRound: saltRound,
	}
}

func (p *passwordLib) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.saltRound)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *passwordLib) Compare(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
