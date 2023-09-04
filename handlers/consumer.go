package handlers

import (
	"fmt"
	"rest-gateway/logger"
	"rest-gateway/models"
	"strconv"
	"strings"
	"time"

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
			log.Warnf("ConsumeHandleMessage - consume: Connection does not exist")
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Server error",
			})
		}
		reqBody.initializeDefaults()
		msgs, err := conn.FetchMessages(stationName, reqBody.ConsumerName,
			memphis.FetchBatchSize(reqBody.BatchSize),
			memphis.FetchConsumerGroup(reqBody.ConsumerGroup),
			memphis.FetchBatchMaxWaitTime(time.Duration(reqBody.BatchMaxWaitTimeMs)*time.Millisecond),
			memphis.FetchMaxMsgDeliveries(1))

		if err != nil {
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
				log.Errorf("ConsumeHandleMessage - acknowledge message: %s", err)
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
