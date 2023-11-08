package handlers

import (
	"encoding/json"
	"errors"
	"strconv"

	"strings"

	"github.com/memphisdev/memphis-rest-gateway/logger"
	"github.com/memphisdev/memphis-rest-gateway/models"

	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

func handleHeaders(headers map[string][]string) (memphis.Headers, error) {
	hdrs := memphis.Headers{}
	hdrs.New()

	for key, value := range headers {
		err := hdrs.Add(key, value[0])
		if err != nil {
			return memphis.Headers{}, err
		}
	}
	return hdrs, nil
}

func CreateHandleMessage() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger(c)
		// We do this parse to params instead of use fiber because there is a memory leak error in fiber
		// stationName := c.Params("stationName")
		url := c.Request().URI().String()
		urlParts := strings.Split(url, "/")
		stationName := urlParts[4]
		bodyReq := c.Body()
		headers := c.GetReqHeaders()
		contentType := string(c.Request().Header.ContentType())
		caseText := strings.Contains(contentType, "text")
		caseJson := strings.Contains(contentType, "application/json")
		if caseText {
			contentType = "text/"
		} else if caseJson {
			contentType = "application/json"
		}

		switch contentType {
		case "application/json", "text/", "application/x-protobuf":
			message := bodyReq
			hdrs, err := handleHeaders(headers)
			if err != nil {
				log.Errorf("CreateHandleMessage - handleHeaders: %s", err.Error())
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}
			userData, ok := c.Locals("userData").(models.AuthSchema)
			if !ok {
				log.Errorf("CreateHandleMessage: failed to get the user data from the middleware")
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}
			username := userData.Username
			accountId := userData.AccountId
			accountIdStr := strconv.Itoa(int(accountId))
			conn := ConnectionsCache[accountIdStr][username].Connection
			if conn == nil {
				conn, err = Connect(userData.Password, username, userData.ConnectionToken, int(accountId))
				if err != nil {
					errMsg := strings.ToLower(err.Error())
					if strings.Contains(errMsg, ErrorMsgAuthorizationViolation) || strings.Contains(errMsg, "token") || strings.Contains(errMsg, ErrorMsgMissionAccountId) {
						log.Warnf("Could not establish new connection with the broker: Authentication error")
						return c.Status(401).JSON(fiber.Map{
							"message": "Unauthorized",
						})
					}

					log.Errorf("Could not establish new connection with the broker: %s", err.Error())
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"message": "Server error",
					})
				}
				if ConnectionsCache[accountIdStr] == nil {
					ConnectionsCacheLock.Lock()
					ConnectionsCache[accountIdStr] = make(map[string]Connection)
					ConnectionsCacheLock.Unlock()
				}

				ConnectionsCacheLock.Lock()
				ConnectionsCache[accountIdStr][username] = Connection{Connection: conn, ExpirationTime: userData.TokenExpiry}
				ConnectionsCacheLock.Unlock()
			}
			err = conn.Produce(stationName, "rest-gateway", message, []memphis.ProducerOpt{}, []memphis.ProduceOpt{memphis.MsgHeaders(hdrs)})
			if err != nil {
				if !strings.Contains(strings.ToLower(err.Error()), "schema validation") {
					log.Errorf("CreateHandleMessage - produce: %s", err.Error())
					c.Status(fiber.StatusInternalServerError)
				} else {
					c.Status(fiber.StatusBadRequest)
				}
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}
		default:
			return errors.New("unsupported content type")
		}

		c.Status(200)
		return c.JSON(&fiber.Map{
			"success": true,
			"error":   nil,
		})
	}
}

func CreateHandleBatch() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger(c)
		// We do this parse to params instead of use fiber because there is a memory leak error in fiber
		// stationName := c.Params("stationName")
		url := c.Request().URI().String()
		urlParts := strings.Split(url, "/")
		stationName := urlParts[4]
		bodyReq := c.Body()
		headers := c.GetReqHeaders()
		contentType := string(c.Request().Header.ContentType())

		switch contentType {
		case "application/json":
			var batchReq []map[string]any
			err := json.Unmarshal(bodyReq, &batchReq)
			if err != nil {
				log.Errorf("CreateHandleBatch - body unmarshal: %s", err.Error())
				return errors.New("unsupported request")
			}
			hdrs, err := handleHeaders(headers)
			if err != nil {
				log.Errorf("CreateHandleBatch - handleHeaders: %s", err.Error())
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}

			userData, ok := c.Locals("userData").(models.AuthSchema)
			if !ok {
				log.Errorf("CreateHandleBatch: failed to get the user data from the middleware")
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}

			username := userData.Username
			accountId := userData.AccountId
			accountIdStr := strconv.Itoa(int(accountId))
			conn := ConnectionsCache[accountIdStr][username].Connection
			if conn == nil {
				conn, err = Connect(userData.Password, username, userData.ConnectionToken, int(accountId))
				if err != nil {
					errMsg := strings.ToLower(err.Error())
					if strings.Contains(errMsg, ErrorMsgAuthorizationViolation) || strings.Contains(errMsg, "token") || strings.Contains(errMsg, ErrorMsgMissionAccountId) {
						log.Warnf("Could not establish new connection with the broker: Authentication error")
						return c.Status(401).JSON(fiber.Map{
							"message": "Unauthorized",
						})
					}

					log.Errorf("Could not establish new connection with the broker: %s", err.Error())
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"message": "Server error",
					})
				}
				if ConnectionsCache[accountIdStr] == nil {
					ConnectionsCacheLock.Lock()
					ConnectionsCache[accountIdStr] = make(map[string]Connection)
					ConnectionsCacheLock.Unlock()
				}

				ConnectionsCacheLock.Lock()
				ConnectionsCache[accountIdStr][username] = Connection{Connection: conn, ExpirationTime: userData.TokenExpiry}
				ConnectionsCacheLock.Unlock()
			}

			errCount := 0
			var allErr []string
			for _, msg := range batchReq {
				rawRes, err := json.Marshal(msg)
				if err != nil {
					errCount++
					allErr = append(allErr, err.Error())
					continue
				}
				if err := conn.Produce(stationName, "rest-gateway", rawRes, []memphis.ProducerOpt{}, []memphis.ProduceOpt{memphis.MsgHeaders(hdrs)}); err != nil {
					if !strings.Contains(strings.ToLower(err.Error()), "schema validation") {
						log.Errorf("CreateHandleBatch - produce: %s", err.Error())
						c.Status(fiber.StatusInternalServerError)
					} else {
						c.Status(fiber.StatusBadRequest)
					}
					errCount++
					allErr = append(allErr, err.Error())
					return c.JSON(&fiber.Map{
						"success": false,
						"error":   allErr,
					})
				}
			}

			if errCount > 0 {
				c.Status(400)
				return c.JSON(&fiber.Map{
					"success": false,
					"sent":    len(batchReq) - errCount,
					"fail":    errCount,
					"errors":  allErr,
				})
			}
		default:
			return errors.New("unsupported content type")
		}

		c.Status(200)
		return c.JSON(&fiber.Map{
			"success": true,
			"error":   nil,
		})
	}
}
