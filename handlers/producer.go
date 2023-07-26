package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"rest-gateway/logger"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/memphisdev/memphis.go"
)

func getTenantNameAndUserFromToken(bearerToken string) (string, string, error) {
	re := regexp.MustCompile(`Bearer\s+(.*?)$`)
	matches := re.FindStringSubmatch(bearerToken)

	if len(matches) < 2 {
		return "", "", fmt.Errorf("Bearer token not found in the bearerToken")
	}

	token := matches[1]
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(configuration.JWT_SECRET), nil
	})
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		// Handle the case where the claims are not of the expected type
	}
	if err != nil {
		return "", "", err
	}

	tenantName := strconv.Itoa(int(claims["account_id"].(float64)))
	username := claims["username"].(string)

	return tenantName, username, nil
}

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

func CreateHandleMessage() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		log := logger.GetLogger(c)
		stationName := c.Params("stationName")
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
				return err
			}
			tenantName, username, _ := getTenantNameAndUserFromToken(headers["Authorization"])
			conn := connectionsCache[tenantName][username].Connection
			if conn == nil {
				errMsg := fmt.Sprintf("Connection does not exists")
				log.Errorf("CreateHandleMessage - produce: %s", errMsg)

				c.Status(500)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   errMsg,
				})
			}

			err = conn.Produce(stationName, "rest_gateway", message, []memphis.ProducerOpt{}, []memphis.ProduceOpt{memphis.MsgHeaders(hdrs)})
			if err != nil {
				log.Errorf("CreateHandleMessage - produce: %s", err.Error())
				c.Status(500)
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
		stationName := c.Params("stationName")
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
				return err
			}

			tenantName, username, _ := getTenantNameAndUserFromToken(headers["Authorization"])
			conn := connectionsCache[tenantName][username].Connection
			if conn == nil {
				errMsg := fmt.Sprintf("Connection does not exists")
				log.Errorf("CreateHandleBatch - produce: %s", errMsg)

				c.Status(500)
				return c.JSON(&fiber.Map{
					"success": false,
					"error":   errMsg,
				})
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
				if err := conn.Produce(stationName, "rest_gateway", rawRes, []memphis.ProducerOpt{}, []memphis.ProduceOpt{memphis.MsgHeaders(hdrs)}); err != nil {
					log.Errorf("CreateHandleBatch - produce: %s", err.Error())
					errCount++
					allErr = append(allErr, err.Error())
					c.Status(400)
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
