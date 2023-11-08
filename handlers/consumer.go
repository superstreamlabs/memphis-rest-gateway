package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/memphisdev/memphis-rest-gateway/logger"
	"github.com/memphisdev/memphis-rest-gateway/models"

	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

type requestBody struct {
	ConsumerName       string `json:"consumer_name"`
	ConsumerGroup      string `json:"consumer_group"`
	BatchSize          int    `json:"batch_size"`
	BatchMaxWaitTimeMs int    `json:"batch_max_wait_time_ms"`
}

func (r *requestBody) initializeDefaults() {
	if r.ConsumerGroup == "" {
		r.ConsumerGroup = "rest-gateway"
	} else {
		r.ConsumerGroup = fmt.Sprintf("%s-rest-gateway", r.ConsumerGroup)
	}
	if r.BatchSize == 0 {
		r.BatchSize = 10
	}
	if r.BatchMaxWaitTimeMs == 0 {
		r.BatchMaxWaitTimeMs = 5000
	}
}

func ConsumeHandleMessage() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger(c)
		url := c.Request().URI().String()
		urlParts := strings.Split(url, "/")
		stationName := urlParts[4]
		reqBody := requestBody{}
		err := c.BodyParser(&reqBody)
		if err != nil {
			log.Errorf("ConsumeHandleMessage - parse request body: %s", err.Error())
			c.Status(fiber.StatusBadRequest)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Invalid request body",
			})
		}
		if reqBody.ConsumerName == "" {
			c.Status(fiber.StatusBadRequest)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Consumer name is required",
			})
		}
		userData, ok := c.Locals("userData").(models.AuthSchema)
		if !ok {
			log.Errorf("ConsumeHandleMessage: failed to get the user data from the middleware")
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
		reqBody.initializeDefaults()
		msgs, err := conn.FetchMessages(stationName, reqBody.ConsumerName,
			memphis.FetchBatchSize(reqBody.BatchSize),
			memphis.FetchConsumerGroup(reqBody.ConsumerGroup),
			memphis.FetchBatchMaxWaitTime(time.Duration(reqBody.BatchMaxWaitTimeMs)*time.Millisecond),
			memphis.FetchMaxMsgDeliveries(1)) // for cases of broker crash before sending the messages to the client

		if err != nil && !strings.Contains(err.Error(), "fetch timed out") {
			log.Errorf("ConsumeHandleMessage - fetch messages: %s", err.Error())
			c.Status(fiber.StatusBadRequest)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		type message struct {
			Message string            `json:"message"`
			Headers map[string]string `json:"headers"`
		}
		messages := []message{}

		for _, msg := range msgs {
			err := msg.Ack()
			if err != nil {
				time.AfterFunc(5*time.Second, func() { // retry after 5 seconds for cases of broker crash
					err := msg.Ack()
					if err != nil {
						log.Errorf("ConsumeHandleMessage - acknowledge message: %s", err)
					}
				})
			}
			messages = append(messages, message{
				Message: string(msg.Data()),
				Headers: msg.GetHeaders(),
			})
		}
		c.Status(fiber.StatusOK)
		return c.JSON(&messages)
	}
}
