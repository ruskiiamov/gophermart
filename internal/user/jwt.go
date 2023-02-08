package user

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
)

const TTL = 30 * time.Minute

func createJWT(secret, userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    userID,
		"created_at": strconv.FormatInt(time.Now().Unix(), 10),
	})

	accessToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("JWT signing error: %w", err)
	}

	return accessToken, nil
}

func getUserIDFromJWT(secret, accessToken string) (string, error) {
	token, _ := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrTokenNotValid
	}

	createdAt, ok := claims["created_at"].(string)
	if !ok {
		return "", ErrTokenNotValid
	}

	intCreatedAt, err := strconv.ParseInt(createdAt, 10, 64)
	if err != nil {
		return "", ErrTokenNotValid
	}

	if time.Unix(intCreatedAt, 0).Add(TTL).Before(time.Now()) {
		return "", ErrTokenNotValid
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrTokenNotValid
	}

	return userID, nil
}
