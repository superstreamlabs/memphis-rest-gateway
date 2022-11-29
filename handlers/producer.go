package handler

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/memphisdev/memphis.go"
	"google.golang.org/protobuf/proto"
)

type Handler struct{ P *memphis.Producer }

func (p Handler) produce(message []byte, hdrs memphis.Headers) error {
	if err := p.P.Produce(message, memphis.MsgHeaders(hdrs)); err != nil {
		return err
	}
	return nil

}

func (p Handler) HandleMessageHdrs(bodyReq []byte, hdrs memphis.Headers) ([]byte, memphis.Headers, error) {
	type body struct {
		Message string `json:"message"`
		Headers string `json:"headers"`
	}
	var b body
	err := json.Unmarshal(bodyReq, &b)
	if err != nil {
		return nil, memphis.Headers{}, err
	}

	var headers map[string]string
	err = json.Unmarshal([]byte(b.Headers), &headers)
	if err != nil {
		return nil, memphis.Headers{}, err
	}

	var k, v string
	for key, value := range headers {
		k = key
		v = value

		err = hdrs.Add(k, v)
		if err != nil {
			return nil, memphis.Headers{}, err
		}
	}

	message, err := json.Marshal(b.Message)
	if err != nil {
		return nil, memphis.Headers{}, err
	}
	return message, hdrs, nil
}

func (p Handler) HandleMessage(c *fiber.Ctx) error {
	bodyReq := c.Body()
	contentType := string(c.Request().Header.ContentType())
	var message []byte
	var err error
	hdrs := memphis.Headers{}
	hdrs.New()
	caseText := strings.Contains(contentType, "text")

	if caseText {
		contentType = "text/"
	}

	switch contentType {
	case "application/json":
		message, hdrs, err = p.HandleMessageHdrs(bodyReq, hdrs)
		if err != nil {
			return err
		}
	case "text/":
		message, hdrs, err = p.HandleMessageHdrs(bodyReq, hdrs)
		if err != nil {
			return err
		}
	case "application/x-protobuf":
		msg := &Msg{}
		err := proto.Unmarshal(bodyReq, msg)
		if err != nil {
			log.Fatal("unmarshaling error: ", err)
		}

		message, err = json.Marshal(msg.Message)
		if err != nil {
			return err
		}

		var headers map[string]string
		err = json.Unmarshal([]byte(msg.Headers), &headers)
		if err != nil {
			return err
		}

		var k, v string
		for key, value := range headers {
			k = key
			v = value
			err = hdrs.Add(k, v)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported content type")
	}

	if err := p.produce(message, hdrs); err != nil {
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
