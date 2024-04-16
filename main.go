package main

import (
	"arduino-serial/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jacobsa/go-serial/serial"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	redisClient, err := InitRedis(cfg.Redis.DSN, 3)
	if err != nil {
		panic(err)
	}

	mappingLabelFile, err := os.ReadFile(cfg.Mapping.LabelToCategory)
	if err != nil {
		panic(err)
	}

	mappingLabel := map[string]string{}
	err = json.Unmarshal(mappingLabelFile, &mappingLabel)
	if err != nil {
		panic(err)
	}

	mappingServoFile, err := os.ReadFile(cfg.Mapping.CategoryToServo)
	if err != nil {
		panic(err)
	}

	mappingServo := map[string][]int{}
	err = json.Unmarshal(mappingServoFile, &mappingServo)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Set up options.
	options := serial.OpenOptions{
		PortName:        cfg.Serial.PortName,
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1000,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	// signalChan := make(chan os.Signal, 1)
	// signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		isAck := false

		for {
			serialData := make([]byte, 1)
			n, err := port.Read(serialData)
			if err != nil {
				panic(err)
			}

			if n <= 0 {
				continue
			}

			if string(serialData) != "1" {
				continue
			}

			if !isAck {
				_, err = port.Write([]byte("1\n"))
				if err != nil {
					panic(err)
				}
				isAck = true
			}

			fmt.Println("Arduino is ready")
			break
		}

		for {
			value, err := redisClient.Get(ctx, cfg.Redis.Key).Result()
			if err == redis.Nil {
				fmt.Println("key does not exist")
			} else if err != nil {
				panic(err)
			} else {
				fmt.Println("key", value)
			}

			payload := value
			fmt.Println("Payload: ", payload)
			// payload := "bottle"

			labelToCategory, ok := mappingLabel[payload]
			if !ok {
				labelToCategory = "fiveFinger"
			}

			servoToMove, ok := mappingServo[labelToCategory]
			if !ok {
				servoToMove = mappingServo["fiveFinger"]
			}

			var servoValue string
			for idx, servo := range servoToMove {
				if idx == len(servoToMove)-1 {
					servoValue += fmt.Sprintf("%d", servo)
				} else {
					servoValue += fmt.Sprintf("%d ", servo)
				}
			}

			_, err = port.Write([]byte(servoValue))
			if err != nil {
				log.Fatalf("failed on writing to serial: %v", err)
				continue
			}

			_, err = port.Write([]byte("\n"))
			if err != nil {
				log.Fatalf("failed on writing to serial: %v", err)
				continue
			}

			break
		}
	}
}
