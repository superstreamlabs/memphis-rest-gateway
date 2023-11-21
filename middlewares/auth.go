package middlewares

import (
	"errors"
	"fmt"
	"rest-gateway/conf"
	"rest-gateway/handlers"
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

func verifyToken(tokenString string, secret string) (models.AuthSchema, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return models.AuthSchema{}, errors.New("f")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return models.AuthSchema{}, errors.New("f")
	}

	var user models.AuthSchema
	if _, ok := claims["username"].(string); ok {
		if !configuration.USER_PASS_BASED_AUTH {
			user = models.AuthSchema{
				Username:        claims["username"].(string),
				ConnectionToken: claims["connection_token"].(string),
				AccountId:       1,
				TokenExpiry:     claims["exp"].(int64),
			}
		} else {
			user = models.AuthSchema{
				Username:    claims["username"].(string),
				Password:    claims["password"].(string),
				AccountId:   claims["account_id"].(float64),
				TokenExpiry: int64(claims["exp"].(float64)),
			}
		}
	} else {
		// for backward compatability
		user = models.AuthSchema{
			TokenExpiryMins: int(claims["exp"].(float64)),
			TokenExpiry:     int64(claims["exp"].(float64)),
		}
	}
	return user, nil
}

func Authenticate(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	path := strings.ToLower(string(c.Context().URI().RequestURI()))
	var user models.AuthSchema
	var err error
	path = strings.Split(path, "?")[0]
	if isAuthNeeded(path) {
		headers := c.GetReqHeaders()
		tokenString := ""
		if len(headers["Authorization"]) == 0 {
			tokenString = c.Query("authorization")
			if tokenString == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Unauthorized",
				})
			}
		} else {
			tokenString, err = extractToken(headers["Authorization"][0])
			if err != nil || tokenString == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"message": "Unauthorized",
				})
			}
		}
		
		user, err = verifyToken(tokenString, configuration.JWT_SECRET)
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
				"message": "Server error",
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

		user, err = verifyToken(body.JwtRefreshToken, configuration.REFRESH_JWT_SECRET)
		if err != nil {
			log.Warnf("Authentication error - refresh token validation has failed")
			if configuration.DEBUG {
				fmt.Printf("Method: %s, Path: %s, IP: %s\nBody: %s\n", c.Method(), c.Path(), c.IP(), string(c.Body()))
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}
	} else if !configuration.USER_PASS_BASED_AUTH && !isAuthNeeded(path) {
		user.AccountId = 1
	}

	// for backward compatability
	if strings.HasSuffix(path, "/produce/single") || strings.HasSuffix(path, "/produce/batch") || path == "/auth/refreshtoken" {
		if user.Username == "" {
			accountId := 1
			conn, err := handlers.Connect(configuration.ROOT_PASSWORD, configuration.ROOT_USER, configuration.CONNECTION_TOKEN, accountId)
			if err != nil {
				errMsg := strings.ToLower(err.Error())
				if strings.Contains(errMsg, handlers.ErrorMsgAuthorizationViolation) || strings.Contains(errMsg, "token") || strings.Contains(errMsg, handlers.ErrorMsgMissionAccountId) {
					log.Warnf("Authentication error")
					return c.Status(401).JSON(fiber.Map{
						"message": "Unauthorized",
					})
				}
			}
			if handlers.ConnectionsCache["1"] == nil {
				handlers.ConnectionsCacheLock.Lock()
				handlers.ConnectionsCache["1"] = make(map[string]handlers.Connection)
				handlers.ConnectionsCacheLock.Unlock()
			}

			handlers.ConnectionsCache["1"][configuration.ROOT_USER] = handlers.Connection{Connection: conn, ExpirationTime: int64(user.TokenExpiryMins)}

			if !configuration.USER_PASS_BASED_AUTH {
				user = models.AuthSchema{
					Username:        configuration.ROOT_USER,
					ConnectionToken: configuration.CONNECTION_TOKEN,
					AccountId:       1,
				}
			} else {
				user = models.AuthSchema{
					Username:  configuration.ROOT_USER,
					Password:  configuration.ROOT_PASSWORD,
					AccountId: 1,
				}
			}
		}
	}

	c.Locals("userData", user)
	return c.Next()
}
