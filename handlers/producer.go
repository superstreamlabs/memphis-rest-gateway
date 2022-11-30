package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

func handleHeaders(headers map[string]string) (memphis.Headers, error) {
	hdrs := memphis.Headers{}
	hdrs.New()

	for key, value := range headers {
		err := hdrs.Add(key, value)
		if err != nil {
			return memphis.Headers{}, err
		}
	}
	return hdrs, nil
}

func handleJsonMessage(bodyReq []byte, headers map[string]string) ([]byte, memphis.Headers, error) {
	type body struct {
		Message string `json:"message"`
	}
	var bodyRequest body
	err := json.Unmarshal(bodyReq, &bodyRequest)
	if err != nil {
		return nil, memphis.Headers{}, err
	}

	hdrs, err := handleHeaders(headers)
	if err != nil {
		return nil, memphis.Headers{}, err
	}

	message, err := json.Marshal(bodyRequest.Message)
	if err != nil {
		return nil, memphis.Headers{}, err
	}
	return message, hdrs, nil
}

func CreateHandleMessage(conn *memphis.Conn) func(*fiber.Ctx) error {
	producers := make(map[string]*memphis.Producer)
	return func(c *fiber.Ctx) error {
		stationName := c.Params("stationName")
		producerName := c.Params("producerName")
		var producer *memphis.Producer
		var err error

		if len(producers) == 0 || producers[stationName].Name != producerName {
			producer, err = conn.CreateProducer(stationName, producerName)
			if err != nil {
				return err
			}
			producers[stationName] = producer
		} else {
			producer = producers[stationName]
		}

		bodyReq := c.Body()
		headers := c.GetReqHeaders()
		contentType := string(c.Request().Header.ContentType())
		var message []byte
		hdrs := memphis.Headers{}
		caseText := strings.Contains(contentType, "text")
		if caseText {
			contentType = "text/"
		}

		switch contentType {
		case "application/json":
			message, hdrs, err = handleJsonMessage(bodyReq, headers)
			if err != nil {
				return err
			}
		case "text/":
			message = bodyReq
			hdrs, err = handleHeaders(headers)
			if err != nil {
				return err
			}
		case "application/x-protobuf":
			message = bodyReq
			hdrs, err = handleHeaders(headers)
			if err != nil {
				return err
			}
		default:
			return errors.New("unsupported content type")
		}

		if err := producer.Produce(message, memphis.MsgHeaders(hdrs)); err != nil {
			c.Status(400)
			return c.JSON(&fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		c.Status(200)
		return c.JSON(&fiber.Map{
			"success": true,
			"error":   nil,
		})
	}
}
