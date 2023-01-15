package handlers

import (
	"http-proxy/logger"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type DevInfoHandler struct{}

func (ih DevInfoHandler) GetSystemInfo(c *fiber.Ctx) error {
	log := logger.GetLogger(c)
	memoryUsage := float64(0)
	cpuUsage := float64(0)
	if runtime.GOOS != "windows" {
		pid := os.Getpid()
		strPid := strconv.Itoa(pid)
		out, err := exec.Command("ps", "-p", strPid, "-o", "%cpu").Output()
		if err != nil {
			log.Errorf("GetInfo: exec command: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		output := string(out[:])
		splitted_output := strings.Split(output, "\n")
		stringCpu := strings.ReplaceAll(splitted_output[1], " ", "")
		if stringCpu != "0.0" {
			cpuUsage, err = strconv.ParseFloat(stringCpu, 64)
			if err != nil {
				log.Errorf("GetInfo: ParseFloat1: %s", err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
		}
		out2, err := exec.Command("ps", "-p", strPid, "-o", "%mem").Output()
		if err != nil {
			log.Errorf("GetInfo: exec command: %s", err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		output2 := string(out2[:])
		splitted_output2 := strings.Split(output2, "\n")
		stringMem := strings.ReplaceAll(splitted_output2[1], " ", "")
		if stringMem != "0.0" {
			memoryUsage, err = strconv.ParseFloat(stringMem, 64)
			if err != nil {
				log.Errorf("GetInfo: ParseFloat2: %s", err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"cpu":    cpuUsage,
		"memory": memoryUsage,
	})
}
