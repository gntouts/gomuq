package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidStringMsg = errors.New("not enough parts for valid message")
var MessageKeys []string = []string{
	"b0",
	"b1",
	"b2",
	"b3",
	"b4",
	"b5",
	"b6",
	"b7",
	"b8",
	"b9",
	"b10",
	"b11",
	"b12",
	"b13",
	"b14",
	"b15",
	"b16",
	"b17",
	"b18",
	"b19",
	"b20",
	"b21",
	"b22",
	"b23",
	"f0",
	"f1",
	"f2",
	"i0",
	"i1",
	"i2",
}

type Message struct {
	data map[string]string
}

func MessageFromString(input string) (Message, error) {
	// 0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00,0.00

	Log.Info(input)
	input = strings.TrimSpace(input)
	// parts2 := strings.Split(input, ",")
	// 5 boolean , 10integers ,10floats

	var newMessage Message
	m := make(map[string]string)
	newMessage.data = m
	input = strings.TrimSpace(input)
	parts := strings.Split(input, ",")
	if len(parts) != 9 {
		return Message{}, ErrInvalidStringMsg
	}

	// Parse first 8-bit int
	bool_temp, err := stringToBoolArr(parts[0])
	if err != nil {
		return Message{}, err
	}
	for i := 0; i < 8; i++ {
		key := "b" + strconv.Itoa(i)
		newMessage.data[key] = strconv.FormatBool(bool_temp[i])
	}

	// Parse second 8-bit int
	bool_temp, err = stringToBoolArr(parts[1])
	if err != nil {
		return Message{}, err
	}
	for i := 8; i < 16; i++ {
		key := "b" + strconv.Itoa(i)
		newMessage.data[key] = strconv.FormatBool(bool_temp[i-8])
	}

	// Parse third 8-bit int
	bool_temp, err = stringToBoolArr(parts[2])
	if err != nil {
		return Message{}, err
	}
	for i := 16; i < 24; i++ {
		key := "b" + strconv.Itoa(i)
		newMessage.data[key] = strconv.FormatBool(bool_temp[i-16])
	}

	// Since we store the floats as strings, no need to parse. But we need to verify they are valid floats
	if err := isValidFloat(parts[3]); err != nil {
		return Message{}, err
	}
	newMessage.data["f0"] = parts[3]

	if err := isValidFloat(parts[4]); err != nil {
		return Message{}, err
	}
	newMessage.data["f1"] = parts[4]

	if err := isValidFloat(parts[5]); err != nil {
		return Message{}, err
	}
	newMessage.data["f2"] = parts[5]

	// Same applies for integers
	if err := isValidInt(parts[6]); err != nil {
		return Message{}, err
	}
	newMessage.data["i0"] = parts[6]

	if err := isValidInt(parts[7]); err != nil {
		return Message{}, err
	}
	newMessage.data["i1"] = parts[7]

	if err := isValidInt(parts[8]); err != nil {
		return Message{}, err
	}
	newMessage.data["i2"] = parts[8]

	return newMessage, nil
}

func isValidFloat(input string) error {
	_, err := strconv.ParseFloat(input, 64)
	return err
}

func isValidInt(input string) error {
	_, err := strconv.Atoi(input)
	return err
}

func (m Message) ToString() (string, error) {
	msgString := ""

	// convert 0-8 booleans to 8-bit int
	temp := ""
	for i := 0; i < 8; i++ {
		key := "b" + strconv.Itoa(i)
		if m.data[key] == "true" {
			temp += "1"
		} else {
			temp += "0"
		}
	}
	converted, err := strconv.ParseInt(temp, 2, 0)
	if err != nil {
		return "", err
	}
	msgString += strconv.Itoa(int(converted)) + ","

	// convert 8-16 first booleans to 8-bit int
	temp = ""
	for i := 8; i < 16; i++ {
		key := "b" + strconv.Itoa(i)
		if m.data[key] == "true" {
			temp += "1"
		} else {
			temp += "0"
		}
	}
	converted, err = strconv.ParseInt(temp, 2, 0)
	if err != nil {
		return "", err
	}
	msgString += strconv.Itoa(int(converted)) + ","

	// convert 16-24 first booleans to 8-bit int
	temp = ""
	for i := 16; i < 24; i++ {
		key := "b" + strconv.Itoa(i)
		if m.data[key] == "true" {
			temp += "1"
		} else {
			temp += "0"
		}
	}
	converted, err = strconv.ParseInt(temp, 2, 0)
	if err != nil {
		return "", err
	}
	msgString += strconv.Itoa(int(converted)) + ","

	// The floats are stored as strings, so no need to convert
	msgString += m.data["f0"]
	msgString += m.data["f1"]
	msgString += m.data["f2"]

	// Same goes for ints
	msgString += m.data["i0"]
	msgString += m.data["i1"]
	msgString += m.data["i2"]

	return msgString, nil
}

func MessageFromDB() Message {
	var newMessage Message
	m := make(map[string]string)
	newMessage.data = m
	for _, t := range MessageKeys {
		ctx := context.Background()
		temp, err := get(ctx, t)
		if err != nil {
			return Message{}
		}
		newMessage.data[t] = temp
	}
	return newMessage
}

// Converts an 8-bit integer string to an array of 8 bool values
func stringToBoolArr(input string) ([8]bool, error) {
	ret := [8]bool{}
	ints, err := strconv.Atoi(input)
	if err != nil {
		return ret, err
	}
	if ints > 255 {
		return ret, errors.New("input number larger than 8-bit")
	}
	strs := fmt.Sprintf("%b", ints)
	padding := ""
	for i := 0; i < 8-len(strs); i++ {
		padding += "0"
	}
	strs = padding + strs

	for i, c := range strs {
		if string(c) == "0" {
			ret[i] = false
		} else {
			ret[i] = true
		}
	}
	return ret, nil
}
