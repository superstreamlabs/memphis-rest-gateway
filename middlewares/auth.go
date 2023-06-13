package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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
	switch configuration.AUTH_METHOD {
	case "jwt":
		return AuthenticateJWT(c)
	case "api_token":
		return AuthenticateAPIToken(c)
	case "hmac_token":
		return AuthenticateHmacToken(c)
	case "none":
		return AuthenticateNone(c)
	default:
		/* default authentication method for backward compatibility with older configuration files. */
		return AuthenticateJWT(c)
	}
}

func AuthenticateNone(c *fiber.Ctx) error {
	return c.Next()
}

func AuthenticateAPIToken(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	headers := c.GetReqHeaders()

	api_token, ok := headers[configuration.API_TOKEN_HEADER]
	if !ok || api_token == "" {
		log.Warnf("Authentication error - API token header is either empty or missing")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	if api_token != configuration.API_TOKEN {
		log.Warnf("Authentication error - API token mismatch")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.Next()
}

func AuthenticateHmacToken(c *fiber.Ctx) error {
	var hash hash.Hash

	log := logger.GetLogger(c)
	headers := c.GetReqHeaders()

	signature, ok := headers[configuration.HMAC_TOKEN_HEADER]
	if !ok || signature == "" {
		log.Warnf("Authentication error - token header is either empty or missing")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	secret := []byte(configuration.HMAC_TOKEN_SECRET)

	switch configuration.HMAC_TOKEN_HASH {
	case "sha512":
		hash = hmac.New(sha512.New, secret)
	case "sha256":
		hash = hmac.New(sha256.New, secret)
	default:
		log.Warnf("Authentication error - hmac hash is missing")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})

	}

	body := c.Body()
	hash.Write(body)
	calculated_signature := hex.EncodeToString(hash.Sum(nil))

	if calculated_signature != signature {
		log.Warnf("Authentication error - signature mismatch")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.Next()
}

func AuthenticateJWT(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	path := strings.ToLower(string(c.Context().URI().RequestURI()))
	if isAuthNeeded(path) {
		headers := c.GetReqHeaders()
		tokenString, err := extractToken(headers["Authorization"])
		if err != nil || tokenString == "" {
			tokenString = c.Query("authorization")
			if tokenString == "" { // fallback - get the token from the query params
				log.Warnf("Authentication error - jwt token is missing")
				if configuration.DEBUG {
					fmt.Printf("Method: %s, Path: %s, IP: %s\nBody: %s\n", c.Method(), c.Path(), c.IP(), string(c.Body()))
				}
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Unauthorized",
				})
			}
		}
		err = verifyToken(tokenString, configuration.JWT_SECRET)
		if err != nil {
			log.Warnf("Authentication error - jwt token validation has failed")
			if configuration.DEBUG {
				fmt.Printf("Method: %s, Path: %s, IP: %s\nBody: %s\n", c.Method(), c.Path(), c.IP(), string(c.Body()))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
	} else if path == "/auth/refreshtoken" {
		var body models.RefreshTokenSchema
		if err := c.BodyParser(&body); err != nil {
			log.Errorf("Authenticate: %s", err.Error())
			if configuration.DEBUG {
				fmt.Printf("Method: %s, Path: %s, IP: %s\nBody: %s\n", c.Method(), c.Path(), c.IP(), string(c.Body()))
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if body.JwtRefreshToken == "" {
			log.Warnf("Authentication error - refresh token is missing")
			if configuration.DEBUG {
				fmt.Printf("Method: %s, Path: %s, IP: %s\nBody: %s\n", c.Method(), c.Path(), c.IP(), string(c.Body()))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		err := verifyToken(body.JwtRefreshToken, configuration.REFRESH_JWT_SECRET)
		if err != nil {
			log.Warnf("Authentication error - refresh token validation has failed")
			if configuration.DEBUG {
				fmt.Printf("Method: %s, Path: %s, IP: %s\nBody: %s\n", c.Method(), c.Path(), c.IP(), string(c.Body()))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
	}

	return c.Next()
}
