package auth

import (
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"
    "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var ErrNoAuthHeaderIncluded = errors.New("no authorization header included")

// in your auth package
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// MakeJWT -
func MakeJWT(userID int, tokenSecret string, expiresIn time.Duration) (string, error) {
    signingKey := []byte(tokenSecret)

    claims := jwt.StandardClaims{
        Issuer:    "chirpy",
        IssuedAt:  time.Now().UTC().Unix(),
        ExpiresAt: time.Now().UTC().Add(expiresIn).Unix(),
        Subject:   fmt.Sprintf("%d", userID),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(signingKey)
}

// ValidateJWT -
func ValidateJWT(tokenString, tokenSecret string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(tokenSecret), nil
    })
    if err != nil {
        return "", err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return "", errors.New("invalid token")
    }

    userIDString, ok := claims["sub"].(string)
    if !ok {
        return "", errors.New("invalid subject")
    }

    return userIDString, nil
}

// GetBearerToken -
func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header included")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
