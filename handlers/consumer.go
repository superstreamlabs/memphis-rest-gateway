package handlers

import (
	"fmt"
	"rest-gateway/cache"
	"rest-gateway/logger"
	"rest-gateway/models"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

var messageCache cache.MessageCache

func InitializeMessageCache(l *logger.Logger) {
	messageCache = cache.New(configuration, l)
}

func ConsumeHandleMessage() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger(c)
		// We do this parse to params instead of use fiber because there is a memory leak error in fiber
		// stationName := c.Params("stationName")
		url := c.Request().URI().String()
		urlParts := strings.Split(url, "/")
		stationName := urlParts[4]
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

		consumerName := getConsumerNameParameterValue(url, username)
		batchSize := getBatchSizeParameterValue(url, 1)
		messages, err := messageCache.GetMessages(stationName, consumerName, batchSize)
		if err != nil {
			log.Errorf("ConsumeHandleMessage - consume: %s", err)
			messages = []cache.Message{}
		}
		if remaining := batchSize - len(messages); remaining > 0 {
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
			consumer, err := conn.CreateConsumer(stationName, consumerName, memphis.PullInterval(15*time.Second))
			if err != nil {
				log.Errorf("ConsumeHandleMessage - consume: %s", err)
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}
			fetchedMessages, err := consumer.Fetch(remaining, true)
			if err != nil {
				log.Errorf("ConsumeHandleMessage - consume: %s", err)
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}
			for _, msg := range fetchedMessages {
				err := msg.Ack()
				if err != nil {
					log.Errorf("ConsumeHandleMessage - consume: %s", err)
				}
				cacheMessage := cache.Message{
					StationName:  stationName,
					ConsumerName: consumerName,
					Username:     username,
					Data:         string(msg.Data()),
				}
				err = messageCache.AddMessage(&cacheMessage)
				if err != nil {
					log.Errorf("ConsumeHandleMessage - consume: %s", err)
				}
				messages = append(messages, cacheMessage)
			}
		}
		c.Status(fiber.StatusOK)
		return c.JSON(&messages)
	}

}

func AcknowledgeMessage() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger(c)
		// We do this parse to params instead of use fiber because there is a memory leak error in fiber
		// stationName := c.Params("stationName")
		url := c.Request().URI().String()
		urlParts := strings.Split(url, "/")
		stationName := urlParts[4]
		type idStruct struct {
			Ids []uint `json:"ids"`
		}
		ids := idStruct{}
		err := c.BodyParser(&ids)
		if err != nil {
			log.Errorf("AcknowledgeMessage: failed to parse array of ids")
			c.Status(fiber.StatusBadRequest)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Invalid request body",
			})
		}
		userData, ok := c.Locals("userData").(models.AuthSchema)
		if !ok {
			log.Errorf("AcknowledgeMessage: failed to get the user data from the middleware")
			c.Status(fiber.StatusInternalServerError)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   "Server error",
			})
		}
		username := userData.Username

		consumerName := getConsumerNameParameterValue(url, username)

		for _, id := range ids.Ids {
			message, err := messageCache.GetMessageById(stationName, consumerName, id)
			if err != nil {
				log.Errorf("AcknowledgeMessage: failed to acknowledge message with id = %v", id)
				c.Status(fiber.StatusInternalServerError)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   "Server error",
				})
			}
			if message.Username == username {
				err := messageCache.RemoveMessage(message)
				if err != nil {
					log.Errorf("AcknowledgeMessage: failed to acknowledge message with id = %v", id)
					c.Status(fiber.StatusInternalServerError)
					return c.JSON(&fiber.Map{
						"success": false,
						"error":   "Server error",
					})
				}
			}
		}
		c.Status(fiber.StatusOK)
		return c.JSON(&fiber.Map{
			"success": true,
		})
	}

}

func getQueryParamValue(urlString, paramName string) (string, error) {
	queryStartIndex := strings.Index(urlString, "?")
	if queryStartIndex == -1 {
		return "", fmt.Errorf("no query parameters in the URL")
	}

	queryString := urlString[queryStartIndex+1:]
	queryParams := strings.Split(queryString, "&")

	for _, param := range queryParams {
		parts := strings.Split(param, "=")
		if len(parts) == 2 && parts[0] == paramName {
			return parts[1], nil
		}
	}

	return "", fmt.Errorf("query parameter '%s' not found", paramName)
}

func getConsumerNameParameterValue(urlString, defaultValue string) string {
	consumerName, err := getQueryParamValue(urlString, "consumerName")
	if err != nil {
		consumerName = defaultValue
	}
	return consumerName

}

func getBatchSizeParameterValue(urlString string, defaultValue int) int {
	batchSizeString, _ := getQueryParamValue(urlString, "batchSize")
	batchSize, err := strconv.ParseInt(batchSizeString, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(batchSize)
}
