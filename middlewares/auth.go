package middlewares

import (
	"errors"
	"fmt"
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/models"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
)

var configuration = conf.GetConfig()
var noNeedAuthRoutes = []string{
	"/",
	"/monitoring/status",
	"/auth/authenticate",
	"/auth/refreshtoken",
	"/monitoring/getresourcesutilization",
}

func isAuthNeeded(path string) bool {
	for _, route := range noNeedAuthRoutes {
		if route == path {
			return false
		}
	}

	return true
}

func extractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("unsupported auth header")
	}

	splited := strings.Split(authHeader, " ")
	if len(splited) != 2 {
		return "", errors.New("unsupported auth header")
	}

	tokenString := splited[1]
	return tokenString, nil
}

func verifyToken(tokenString string, secret string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return errors.New("f")
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return errors.New("f")
	}

	return nil
}

func Authenticate(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	path := strings.ToLower(string(c.Context().URI().RequestURI()))
	if isAuthNeeded(path) {
		headers := c.GetReqHeaders()
		tokenString, err := extractToken(headers["Authorization"])
		if err != nil || tokenString == "" {
			tokenString = c.Query("authorization")
			if tokenString == "" { // fallback - get the token from the query params
				log.Warnf("Authentication error - jwt token is missing")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Unauthorized",
				})
			}
		}
		err = verifyToken(tokenString, configuration.JWT_SECRET)
		if err != nil {
			log.Warnf("Authentication error - jwt token validation has failed")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
	} else if path == "/auth/refreshtoken" {
		var body models.RefreshTokenSchema
		if err := c.BodyParser(&body); err != nil {
			log.Errorf("Authenticate: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if body.JwtRefreshToken == "" {
			log.Warnf("Authentication error - refresh token is missing")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		err := verifyToken(body.JwtRefreshToken, configuration.REFRESH_JWT_SECRET)
		if err != nil {
			log.Warnf("Authentication error - refresh token validation has failed")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
	}

	return c.Next()
}
