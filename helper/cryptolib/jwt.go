package cryptolib

import (
	stderrors "errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yerobalg/wealthpulse-service/helper/errors"
)

type jwtLib struct {
	expiredTimeSec int64
	secretKey      string
}

type JWTInterface interface {
	Encode(data any) (string, error)
	Decode(string) (map[string]any, error)
	Expiry() time.Duration
}

func InitJWT(expiredTimeSec int64, secretKey string) JWTInterface {
	return &jwtLib{
		expiredTimeSec: expiredTimeSec,
		secretKey:      secretKey,
	}
}

func (j *jwtLib) Encode(data any) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  j.expiredTimeSec + time.Now().Unix(),
	})

	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtLib) Decode(token string) (map[string]any, error) {
	decoded, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(j.secretKey), nil
	})
	if err != nil {
		if !stderrors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.Unauthorized("Invalid token")
		}

		claims, ok := decoded.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.Unauthorized("Invalid token")
		}

		return map[string]any(claims), errors.Unauthorized("Token expired")
	}

	claims, ok := decoded.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.InternalServerError("Failed to decode token")
	}
	if !decoded.Valid {
		return nil, errors.Unauthorized("Invalid token")
	}

	return claims, nil
}

func (j *jwtLib) Expiry() time.Duration {
	return time.Duration(j.expiredTimeSec) * time.Second
}
