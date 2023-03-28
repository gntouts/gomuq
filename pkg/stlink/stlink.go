package stlink

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var ErrStFlashNotInstalled error = errors.New("\"st-flash\" command is not available")

func checkDependencies() {
	_, err := exec.Command("which", "st-flash").CombinedOutput()
	if err != nil {
		fmt.Println("[ERR]", ErrStFlashNotInstalled.Error())
		os.Exit(1)
	}
}

func Reset() error {
	checkDependencies()
	output, err := exec.Command("st-flash", "--reset", "read", "garbage.bin", "0x8000000", "256").CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = os.Remove("garbage.bin")
	if err != nil {
		fmt.Println("Reset was successful. Failed to remove garbage.bin.")
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func Flash(filename string) error {
	checkDependencies()
	output, err := exec.Command("st-flash", "--reset", "write", filename, "0x8000000").CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
