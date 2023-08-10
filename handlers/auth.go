package handlers

import (
	"rest-gateway/conf"
	"rest-gateway/logger"
	"rest-gateway/models"
	"rest-gateway/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/memphisdev/memphis.go"
)

var configuration = conf.GetConfig()
var connectionsCacheLock sync.Mutex

type AuthHandler struct{}

type Connection struct {
	Connection     *memphis.Conn `json:"connection"`
	ExpirationTime int64         `json:"expiration_time"`
}

type refreshTokenExpiration struct {
	TokenExpiration        int64 `json:"token_expiration"`
	RefreshTokenExpiration int64 `json:"refresh_token_expiration"`
}

var connectionsCache = map[string]map[string]Connection{}

func connect(password, username, connectionToken string, accountId int) (*memphis.Conn, error) {
	if configuration.USER_PASS_BASED_AUTH {
		if accountId == 0 {
			accountId = 1
		}
	}
	var err error
	opts := []memphis.Option{memphis.Reconnect(true), memphis.MaxReconnect(10), memphis.ReconnectInterval(3 * time.Second)}
	if configuration.USER_PASS_BASED_AUTH {
		opts = append(opts, memphis.Password(password), memphis.AccountId(accountId))
	} else {
		opts = append(opts, memphis.ConnectionToken(connectionToken))
	}
	if configuration.CLIENT_CERT_PATH != "" && configuration.CLIENT_KEY_PATH != "" && configuration.ROOT_CA_PATH != "" {
		opts = append(opts, memphis.Tls(configuration.CLIENT_CERT_PATH, configuration.CLIENT_KEY_PATH, configuration.ROOT_CA_PATH))
	}
	conn, err := memphis.Connect(configuration.MEMPHIS_HOST, username, opts...)
	if err != nil {
		return conn, err
	}
	return conn, nil
}

func (ah AuthHandler) Authenticate(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.AuthSchema
	if err := c.BodyParser(&body); err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}

	conn, err := connect(body.Password, body.Username, body.ConnectionToken, int(body.AccountId))
	if err != nil {
		errorMsgAuthorizationViolation := "authorization violation"
		errorMsgMissionAccountId := "memphis: wrong / missing account id"
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, errorMsgAuthorizationViolation) || strings.Contains(errMsg, "token") || strings.Contains(errMsg, errorMsgMissionAccountId) {
			log.Warnf("Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	if body.AccountId == 0 {
		body.AccountId = 1
	}
	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, body.Username, int(body.AccountId), body.Password, body.ConnectionToken)
	if err != nil {
		log.Errorf("Authenticate: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	username := strings.ToLower(body.Username)
	accountId := strconv.Itoa(int(body.AccountId))
	if connectionsCache[accountId] == nil {
		connectionsCacheLock.Lock()
		connectionsCache[accountId] = make(map[string]Connection)
		connectionsCacheLock.Unlock()
	}

	connectionsCacheLock.Lock()
	connectionsCache[accountId][username] = Connection{Connection: conn, ExpirationTime: tokenExpiry}
	connectionsCacheLock.Unlock()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func createTokens(tokenExpiryMins, refreshTokenExpiryMins int, username string, accountId int, password, connectionToken string) (string, string, int64, int64, error) {
	if tokenExpiryMins <= 0 {
		tokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	if refreshTokenExpiryMins <= 0 {
		refreshTokenExpiryMins = configuration.JWT_EXPIRES_IN_MINUTES
	}

	atClaims := jwt.MapClaims{}
	atClaims["username"] = username
	atClaims["password"] = password
	atClaims["account_id"] = accountId
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	atClaims["connection_token"] = connectionToken
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(configuration.JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}

	atClaims["token_exp"] = time.Now().Add(time.Minute * time.Duration(tokenExpiryMins)).Unix()
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(refreshTokenExpiryMins)).Unix()
	at = jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken, err := at.SignedString([]byte(configuration.REFRESH_JWT_SECRET))
	if err != nil {
		return "", "", 0, 0, err
	}
	tokenExpiry := atClaims["exp"].(int64)
	refreshTokenExpiry := refreshTokenExpiration{
		RefreshTokenExpiration: atClaims["exp"].(int64),
		TokenExpiration:        atClaims["token_exp"].(int64),
	}

	return token, refreshToken, tokenExpiry, refreshTokenExpiry.TokenExpiration, nil
}

func (ah AuthHandler) RefreshToken(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	var body models.RefreshTokenSchema
	if err := c.BodyParser(&body); err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}
	if err := utils.Validate(body); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": err,
		})
	}
	userData, ok := c.Locals("userData").(models.AuthSchema)
	if !ok {
		log.Errorf("RefreshToken: failed to get the user data from the middleware")
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(&fiber.Map{
			"success": false,
			"error":   "Server error",
		})
	}

	username := userData.Username
	accountId := int(userData.AccountId)
	password := userData.Password
	connectionToken := userData.ConnectionToken

	conn, err := connect(password, username, connectionToken, accountId)
	if err != nil {
		errorMsgAuthorizationViolation := "authorization violation"
		errorMsgMissionAccountId := "memphis: wrong / missing account id"
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, errorMsgAuthorizationViolation) || strings.Contains(errMsg, "token") || strings.Contains(errMsg, errorMsgMissionAccountId) {
			log.Warnf("RefreshToken: Authentication error")
			return c.Status(401).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		}

		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	token, refreshToken, tokenExpiry, refreshTokenExpiry, err := createTokens(body.TokenExpiryMins, body.RefreshTokenExpiryMins, username, accountId, password, connectionToken)
	if err != nil {
		log.Errorf("RefreshToken: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Server error",
		})
	}

	accountId = int(accountId)
	if connectionsCache[strconv.Itoa(int(accountId))] == nil {
		connectionsCacheLock.Lock()
		connectionsCache[strconv.Itoa(accountId)] = make(map[string]Connection)
		connectionsCacheLock.Unlock()
	}

	connectionsCacheLock.Lock()
	connectionsCache[strconv.Itoa(accountId)][username] = Connection{Connection: conn, ExpirationTime: refreshTokenExpiry}
	connectionsCacheLock.Unlock()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"jwt":                      token,
		"expires_in":               tokenExpiry * 60 * 1000,
		"jwt_refresh_token":        refreshToken,
		"refresh_token_expires_in": refreshTokenExpiry * 60 * 1000,
	})
}

func CleanConnectionsCache() {
	for range time.Tick(time.Second * 30) {
		for t, tenant := range connectionsCache {
			for u, user := range tenant {
				currentTime := time.Now()
				unixTimeNow := currentTime.Unix()
				conn := connectionsCache[t][u].Connection
				if unixTimeNow > int64(user.ExpirationTime) {
					conn.Close()
					connectionsCacheLock.Lock()
					delete(connectionsCache[t], u)
					connectionsCacheLock.Unlock()
				}
			}
			if len(connectionsCache[t]) == 0 {
				connectionsCacheLock.Lock()
				delete(connectionsCache, t)
				connectionsCacheLock.Unlock()
			}
		}
	}
}
