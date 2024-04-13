package main

import (
	"arduino-serial/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacobsa/go-serial/serial"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	redis, err := InitRedis(cfg.Redis.DSN, 3)
	if err != nil {
		panic(err)
	}

	mappingLabelFile, err := os.ReadFile("mapping_label.json")
	if err != nil {
		panic(err)
	}

	mappingLabel := map[string]string{}
	err = json.Unmarshal(mappingLabelFile, &mappingLabel)
	if err != nil {
		panic(err)
	}

	mappingServoFile, err := os.ReadFile("mapping_servo.json")
	if err != nil {
		panic(err)
	}

	mappingServo := map[string][]int{}
	err = json.Unmarshal(mappingServoFile, &mappingServo)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	subscriber := redis.Subscribe(ctx, "object-detection")

	defer subscriber.Close()

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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		serialData := make([]byte, 1000)
		n, err := port.Read(serialData)
		if err != nil {
			fmt.Println(n, err)
			continue
		}

		if n <= 0 {
			continue
		}

		select {
		case <-signalChan:
			return
		default:
			fmt.Println("Waiting for message")
		}

		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			log.Fatalf("failed on receiving message from redis: %v", err)
			continue
		}

		payload := msg.Payload
		labelToCategory, ok := mappingLabel[payload]
		if !ok {
			labelToCategory = "fiveFinger"
		}

		servoToMove, ok := mappingServo[labelToCategory]
		if !ok {
			servoToMove = mappingServo["fiveFinger"]
		}

		var servoValue string
		for _, servo := range servoToMove {
			servoValue += fmt.Sprintf("%d ", servo)
		}

		reply := fmt.Sprintf("%s\n", servoValue)
		_, err = port.Write([]byte(reply))
		if err != nil {
			log.Fatalf("failed on writing to serial: %v", err)
			continue
		}
	}
}
