package common

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type UartMessage struct {
	Incoming string
	bools    [40]bool
	ints     [10]uint8
	floats   [10]float64
}

func NewMessageFromInput(input string) *UartMessage {
	msg := &UartMessage{
		Incoming: input,
	}
	msg.parseInput()
	return msg
}

func (m *UartMessage) parseInput() error {
	parts := strings.Split(m.Incoming, ",")
	// for each of the 5 first parts BOOL
	for i := 0; i < 5; i++ {
		num, err := strconv.Atoi(parts[i])
		if err != nil {
			return err
		}
		if num > 255 {
			return errors.New("input number larger than 8-bit")
		}
		expanded := fmt.Sprintf("%b", num)
		// add padding to ensure proper mapping
		for len(expanded) < 8 {
			expanded = "0" + expanded
		}
		// now lets populate the bools array
		for k, r := range expanded {
			m.bools[i*8+k] = string(r) == "1"
		}
	}

	// for each of 5-15 parts INT
	for i := 5; i < 15; i++ {
		num, err := strconv.ParseUint(parts[i], 10, 8)
		if err != nil {
			return err
		}
		if num > 255 {
			return errors.New("input number larger than 8-bit")
		}
		m.ints[i-5] = uint8(num)
	}

	// for each of 15-25 parts float
	for i := 15; i < 25; i++ {
		num, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return err
		}
		m.floats[i-15] = num
	}
	return nil
}
func (m *UartMessage) Outgoing() [27]byte {
	var ret [27]byte
	for i := 0; i < 5; i++ {
		var temp [8]bool
		var num uint8
		copy(temp[:], m.bools[i*8:i*8+8])
		for i := 0; i < len(temp); i++ {
			if temp[i] {
				num |= 1 << uint(7-i)
			}
		}
		ret[i] = num
	}
	for ind, num := range m.ints {
		ret[5+ind] = num
	}
	for ind, num := range m.floats {
		temp := fmt.Sprintf("%.2f", num)
		parts := strings.Split(temp, ".")
		// TODO: perhaps log the error(?)
		first, _ := strconv.ParseUint(parts[0], 10, 8)
		second, _ := strconv.ParseUint(parts[1], 10, 8)
		ret[15+ind] = uint8(first)
		ret[16+ind] = uint8(second)

	}
	ret[25] = 0x0A
	ret[26] = 0x0D
	return ret
}

// TODO: Delete (deprecated)
func (m *UartMessage) Out() string {
	result := ""
	// convert booleans to integers
	for i := 0; i < 5; i++ {
		var temp [8]bool
		var num uint8
		copy(temp[:], m.bools[i*8:i*8+8])
		for i := 0; i < len(temp); i++ {
			if temp[i] {
				num |= 1 << uint(7-i)
			}
		}
		result += fmt.Sprintf("%d,", num)
	}
	// add integers as strings
	for _, num := range m.ints {
		result += fmt.Sprintf("%d,", num)
	}
	// add floats
	for _, num := range m.floats {
		temp := fmt.Sprintf("%.2f", num)
		temp = strings.ReplaceAll(temp, ".", ",")
		result += fmt.Sprintf("%s,", temp)
	}
	result = strings.TrimRight(result, ",")
	return result
}
