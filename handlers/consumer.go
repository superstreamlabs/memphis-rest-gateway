package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
	"rest-gateway/logger"
	"rest-gateway/models"
	"strconv"
	"strings"
	"time"
)

type requestBody struct {
	ConsumerName      string `json:"consumer_name"`
	ConsumerGroupName string `json:"consumer_group_name"`
	BatchSize         int    `json:"batch_size"`
	MaxAckTime        int    `json:"max_ack_time"`
	BatchMaxWaitTime  int    `json:"batch_max_wait_time"`
	MaxMsgDeliveries  int    `json:"max_msg_deliveries"`
}

func (r requestBody) initializeDefaults() {
	if r.ConsumerGroupName == "" {
		r.ConsumerGroupName = r.ConsumerName
	}
	if r.BatchSize == 0 {
		r.BatchSize = 10
	}
	if r.BatchMaxWaitTime == 0 {
		r.BatchMaxWaitTime = 5
	}
	if r.MaxAckTime == 0 {
		r.MaxAckTime = 30
	}
	if r.MaxMsgDeliveries == 0 {
		r.MaxMsgDeliveries = 10
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
			errMsg := fmt.Sprintf("Connection does not exist")
			log.Errorf("ConsumeHandleMessage - consume: %s", errMsg)

			c.Status(fiber.StatusInternalServerError)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Server error",
			})
		}
		reqBody.initializeDefaults()
		msgs, err := conn.FetchMessages(stationName, reqBody.ConsumerName,
			memphis.FetchBatchSize(reqBody.BatchSize),
			memphis.FetchConsumerGroup(reqBody.ConsumerGroupName),
			memphis.FetchBatchMaxWaitTime(time.Duration(reqBody.BatchMaxWaitTime)),
			memphis.FetchMaxAckTime(time.Duration(reqBody.MaxAckTime)),
			memphis.FetchMaxMsgDeliveries(reqBody.MaxMsgDeliveries))

		if err != nil {
			log.Errorf("ConsumeHandleMessage - fetch messages: %s", err.Error())
			c.Status(fiber.StatusBadRequest)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Server error",
			})
		}

		type message struct {
			Data string `json:"data"`
		}
		messages := []message{}

		for _, msg := range msgs {
			err := msg.Ack()
			if err != nil {
				log.Errorf("ConsumeHandleMessage - consume: %s", err)
			}
			messages = append(messages, message{string(msg.Data())})
		}
		c.Status(fiber.StatusOK)
		return c.JSON(&messages)
	}
}
