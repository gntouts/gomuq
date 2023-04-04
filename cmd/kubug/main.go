package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/albenik/go-serial/v2"
	common "github.com/gntouts/gomuq/pkg/common"
	"github.com/gntouts/gomuq/pkg/usbtool"
	"github.com/sirupsen/logrus"
)

var uartOut = make(chan string, 1)
var mqttOut = make(chan string, 1)

func findDevice(name string) (string, error) {
	out, err := usbtool.SearchDevice(name)
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
		logrus.Error(err.Error())
		return conn, err
	}
	logrus.Info("Connected to " + usbPath)
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
				logrus.WithError(err).Error("Failed to send UART message:" + msg)
				return ""
			}
			logrus.Info("Sent UART message:" + msg)
			fmt.Println(msg)
		}
		n, err := conn.Read(buff)
		if err != nil {
			logrus.WithError(err).Debug("Failed to read from UART")
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
	// defer conn.Close()
	disconnected := true
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("Program killed!")
		conn.Close()
		os.Exit(1)
	}()
	for {
		if disconnected {
			usbPath, err := findDevice(usbName)
			if err != nil {
				logrus.WithError(err).Error("Failed to find devices with name " + usbName)
				time.Sleep(2 * time.Second)
				continue
			}
			conn, err = connectUart(usbPath, baudRate)
			if err != nil {
				logrus.WithError(err).Error("Failed to connect to " + usbPath)
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
			// message := msg.NewMessageFromInput("3,2,1,0,1,2,3,0,0,0,0,0,0,0,0,0.00,0.00,0.00,0.00,0.00,0.00,15.54,0.00,0.00,0.00")
			// ret := message.Outgoing()
			// fmt.Println(ret)
			// time.Sleep(5 * time.Second)
			// _, err := conn.Write([]byte(ret + "\n\r"))
			// if err != nil {
			// 	fmt.Println(err.Error())
			// }

			for {
				if true {
					msg := waitForSingleMessage(*conn)
					if msg == "" {
						disconnected = true
						logrus.Error("Disconnected from UART device")
						conn.Close()
						break
					}
					parts := strings.Split(msg, ",")
					fmt.Println("Message: ", msg, "Parts: ", len(parts))

					if len(parts) == 25 {
						message := common.NewMessageFromInput(msg)
						fmt.Println(message.Outgoing())
					}
				}
				if false {
					message := common.NewMessageFromInput("3,2,1,0,1,2,3,0,0,0,0,0,0,0,0,0.00,0.00,0.00,0.00,0.00,0.00,15.54,0.00,0.00,0.00")
					ret := message.Outgoing()

					time.Sleep(5 * time.Second)
					_, err := conn.Write(ret[:])
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println("done")
					conn.Close()
					os.Exit(0)

				}

			}
		}
	}
}

func main() {
	uartHandler("ST-LINK", 115200)
}
