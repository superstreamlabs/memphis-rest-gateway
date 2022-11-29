package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/memphisdev/memphis.go"
	"google.golang.org/protobuf/proto"
)

func main() {
	app := fiber.New()
	app.Use(cors.New())

	conn, err := memphis.Connect("localhost", "root", "memphis")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	producer, err := conn.CreateProducer("test-fiber-go", "simple_go_producer")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	app.Post("/", func(c *fiber.Ctx) error {
		type body struct {
			Message string `json:"message"`
			Headers string `json:"headers"`
		}
		var b body
		bodyReq := c.Body()

		contentType := string(c.Request().Header.ContentType())
		var message []byte
		hdrs := memphis.Headers{}
		hdrs.New()
		caseText := strings.Contains(contentType, "text")

		if caseText {
			contentType = "text/"
		}

		switch contentType {
		case "application/json":
			err := json.Unmarshal(bodyReq, &b)
			if err != nil {
				return err
			}

			var headers map[string]string
			err = json.Unmarshal([]byte(b.Headers), &headers)
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

			message, err = json.Marshal(b.Message)
			if err != nil {
				return err
			}
		case "text/":
			err := json.Unmarshal(bodyReq, &b)
			if err != nil {
				return err
			}
			var headers map[string]string
			err = json.Unmarshal([]byte(b.Headers), &headers)
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

			message, err = json.Marshal(b.Message)
			if err != nil {
				return err
			}
		case "application/x-protobuf":
			msg := &Msg{}
			err = proto.Unmarshal(bodyReq, msg)
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
				fmt.Print(err)

				return err
			}

			var k, v string
			for key, value := range headers {
				k = key
				v = value

				err = hdrs.Add(k, v)
				if err != nil {
					fmt.Print(err)
					return err
				}
			}
		default:
			return errors.New("unsupported content type")
		}

		if err := producer.Produce(message, memphis.MsgHeaders(hdrs)); err != nil {
			fmt.Println(err.Error())
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
	})

	app.Listen(":3000")
}
