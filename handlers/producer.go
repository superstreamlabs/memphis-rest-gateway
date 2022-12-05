package handlers

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
)

var producers = make(map[string]*memphis.Producer)

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

func createProducer(conn *memphis.Conn, producers map[string]*memphis.Producer, stationName string) (*memphis.Producer, error) {
	producerName := "http_proxy"
	var producer *memphis.Producer
	var err error
	if _, ok := producers[stationName]; !ok {
		producer, err = conn.CreateProducer(stationName, producerName, memphis.ProducerGenUniqueSuffix())
		if err != nil {
			return nil, err
		}
		producers[stationName] = producer
	} else {
		producer = producers[stationName]
	}

	return producer, nil
}

func CreateHandleMessage(conn *memphis.Conn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		stationName := c.Params("stationName")
		var producer *memphis.Producer

		producer, err := createProducer(conn, producers, stationName)
		if err != nil {
			return err
		}

		bodyReq := c.Body()
		headers := c.GetReqHeaders()
		contentType := string(c.Request().Header.ContentType())
		caseText := strings.Contains(contentType, "text")
		if caseText {
			contentType = "text/"
		}

		switch contentType {
		case "application/json", "text/", "application/x-protobuf":
			message := bodyReq
			hdrs, err := handleHeaders(headers)
			if err != nil {
				return err
			}
			if err := producer.Produce(message, memphis.MsgHeaders(hdrs)); err != nil {
				if strings.Contains(err.Error(), "memphis: no responders available for request") {
					delete(producers, stationName)
					producer, err = createProducer(conn, producers, stationName)
					if err != nil {
						c.Status(400)
						return c.JSON(&fiber.Map{
							"success": false,
							"error":   err.Error(),
						})
					}
					err = producer.Produce(message, memphis.MsgHeaders(hdrs))
					if err != nil {
						c.Status(400)
						return c.JSON(&fiber.Map{
							"success": false,
							"error":   err.Error(),
						})
					}
				} else {
					c.Status(400)
					return c.JSON(&fiber.Map{
						"success": false,
						"error":   err.Error(),
					})
				}
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

func CreateHandleBatch(conn *memphis.Conn) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		stationName := c.Params("stationName")
		var producer *memphis.Producer

		producer, err := createProducer(conn, producers, stationName)
		if err != nil {
			return err
		}

		bodyReq := c.Body()
		headers := c.GetReqHeaders()
		contentType := string(c.Request().Header.ContentType())

		switch contentType {
		case "application/json":
			var batchReq []map[string]any
			err := json.Unmarshal(bodyReq, &batchReq)
			if err != nil {
				return errors.New("unsupported request")
			}
			hdrs, err := handleHeaders(headers)
			if err != nil {
				return err
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
				if err := producer.Produce(rawRes, memphis.MsgHeaders(hdrs)); err != nil {
					if strings.Contains(err.Error(), "memphis: no responders available for request") {
						delete(producers, stationName)
						producer, err = createProducer(conn, producers, stationName)
						if err != nil {
							errCount++
							allErr = append(allErr, err.Error())
							c.Status(400)
							return c.JSON(&fiber.Map{
								"success": false,
								"error":   allErr,
							})
						}
						err = producer.Produce(rawRes, memphis.MsgHeaders(hdrs))
						if err != nil {
							errCount++
							allErr = append(allErr, err.Error())
							c.Status(400)
							return c.JSON(&fiber.Map{
								"success": false,
								"error":   allErr,
							})
						}
					} else {
						errCount++
						allErr = append(allErr, err.Error())
						c.Status(400)
						return c.JSON(&fiber.Map{
							"success": false,
							"error":   allErr,
						})
					}
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
