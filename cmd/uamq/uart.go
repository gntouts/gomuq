package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	serial "github.com/albenik/go-serial/v2"
	"github.com/gntouts/nt-cli/pkg/usbdev"
	"github.com/go-redis/redis/v8"
)

func findDevice(name string) (string, error) {
	out, err := usbdev.SearchDevice(name)
	if err != nil {
		return "", err
	}
	device_path := strings.TrimSpace(string(out.Path))
	if device_path == "" {
		return "", errors.New("device not found")
	}
	return "/dev/" + device_path, nil
}

func connectUart(usbPath string, baudRate int) (*serial.Port, error) {
	conn, err := serial.Open(usbPath, serial.WithBaudrate(baudRate))
	if err != nil {
		Log.Error(err.Error())
		return conn, err
	}
	Log.Info("Connected to " + usbPath)
	return conn, nil
}

func waitForSingleMessage(conn serial.Port) string {
	var b bytes.Buffer
	buff := make([]byte, 100)
	for {
		if len(uartOut) > 0 {
			msg := <-uartOut
			_, err := conn.Write([]byte(msg + "\n\r"))
			if err != nil {
				Log.WithError(err).Error("Failed to send UART message:" + msg)
				return ""
			}
			Log.Info("Sent UART message:" + msg)
			fmt.Println(msg)
		}
		n, err := conn.Read(buff)
		if err != nil {
			Log.WithError(err).Debug("Failed to read from UART")
			if err.Error() == "other error" {
				return ""
			}
			fmt.Println(err.Error())
			break
		}
		if n == 0 {
			buff = make([]byte, 100)
		}
		if n != 0 {
			ret := string(buff[:n])
			for _, c := range ret {
				letter := string(c)
				if letter != "\n" {
					b.WriteString(letter)
				} else {
					temp := b.String()
					if len(temp) != 0 {
						b.Reset()
						return temp
					}
				}
			}
		}
	}

	ret := string(buff)
	return ret
}

func uartHandler(usbName string, baudRate int) {
	var conn *serial.Port
	defer conn.Close()
	disconnected := true
	for {
		if len(uartKill) > 0 {
			<-uartKill
			Log.Info("Received UART kill signal")
			return
		}
		if disconnected {
			usbPath, err := findDevice(usbName)
			if err != nil {
				Log.WithError(err).Error("Failed to find devices with name " + usbName)
				time.Sleep(2 * time.Second)
				continue
			}
			conn, err = connectUart(usbPath, baudRate)
			if err != nil {
				Log.WithError(err).Error("Failed to connect to " + usbPath)
				time.Sleep(2 * time.Second)
				continue
			}
			disconnected = false
		}

		if !disconnected {
			// This helps eliminate incomplete messages at first run
			buff := make([]byte, 100)
			conn.ResetInputBuffer()
			_, _ = conn.Read(buff)

			for {
				if len(uartKill) > 0 {
					<-uartKill
					Log.Info("Received UART kill signal")
					return
				}
				msg := waitForSingleMessage(*conn)
				if msg == "" {
					disconnected = true
					Log.Error("Disconnected fro UART device")
					break
				}
				message, err := MessageFromString(string(msg))
				if err != nil {
					Log.WithError(err).Error("Message invalid. Disconnecting from UART")
					conn.Close()
					disconnected = true
					break
				}
				Log.Info("Received UART message:" + msg)
				// check if db is populated
				ctx := context.Background()
				_, err = get(ctx, "b0")
				if err == nil {
					// then db is populated
					dbMessage := MessageFromDB()
					for _, t := range MessageKeys {
						if message.data[t] != dbMessage.data[t] {
							ctx := context.Background()
							set(ctx, t, message.data[t])
							topic := "hass/listen_hass/" + t
							payload := message.data[t]
							mqtt_msg := topic + ":" + payload
							mqttOut <- mqtt_msg
						}
					}
				} else if err == redis.Nil {
					Log.Info("Redis is empty. Populating with received values")
					// then db is empty, let's populate and send everything via mqtt
					for _, t := range MessageKeys {
						ctx := context.Background()
						set(ctx, t, message.data[t])
						topic := "hass/listen_hass/" + t
						payload := message.data[t]
						mqtt_msg := topic + ":" + payload
						mqttOut <- mqtt_msg
					}
				} else {
					Log.WithError(err).Error("Failed to get value from Redis")
				}
			}
		}
	}
}
