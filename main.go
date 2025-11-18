package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sashko/go-uinput"
)

func findDevice(deviceName string) (string, error) {
	data, _ := os.ReadFile("/proc/bus/input/devices")
	for _, block := range strings.Split(string(data), "\n\n") {
		var name, handlers string
		for _, line := range strings.Split(block, "\n") {
			if strings.HasPrefix(line, "N: Name=") {
				start := strings.Index(line, `"`)
				end := strings.LastIndex(line, `"`)
				if start != -1 && end != -1 && end > start {
					name = line[start+1 : end]
				}
			}
			if strings.HasPrefix(line, "H: Handlers=") {
				handlers = strings.TrimPrefix(line, "H: Handlers=")
			}

		}
		//fmt.Printf("DEBUG BLOCK: name=%q handlers=%q\n", name, handlers)
		if strings.TrimSpace(name) == deviceName && handlers != "" {
			for _, h := range strings.Fields(handlers) {
				if strings.HasPrefix(h, "event") {
					return "/dev/input/" + h, nil
				}
			}
		}
	}
	return "", fmt.Errorf("device '%s' not found", deviceName)
}

func main() {
	if os.Geteuid() != 0 {
		cmd := exec.Command("sudo", append([]string{os.Args[0]}, os.Args[1:]...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			fmt.Println("failed to elevate:", err)
		}
		return
	}

	device, err := findDevice("Logitech USB Receiver")
	if err != nil {
		fmt.Println("Error finding device:", err)
		return
	}
	//fmt.Printf("Using input device: %s\n", device)

	file, err := os.Open(device)
	if err != nil {
		fmt.Println("Error opening input device:", err)
		return
	}
	defer file.Close()

	var toggle bool = false
	buf := make([]byte, 24)
	keyboard, err := uinput.CreateKeyboard()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()
	mice, err := uinput.CreateMice(0, 1919, 0, 1079)
	if err != nil {
		return
	}
	defer mice.Close()

	go func() { // deathray loop here
		for {
			if toggle == true {
				keyboard.KeyPress(uinput.Key1)
				time.Sleep(17 * time.Millisecond)
				mice.LeftClick()
				time.Sleep(6 * time.Millisecond)
				keyboard.KeyPress(uinput.Key2)
				time.Sleep(17 * time.Millisecond)
				mice.LeftClick()
				time.Sleep(6 * time.Millisecond)
				keyboard.KeyPress(uinput.Key3)
				time.Sleep(17 * time.Millisecond)
				mice.LeftClick()
				time.Sleep(6 * time.Millisecond)
				keyboard.KeyPress(uinput.Key4)
				time.Sleep(17 * time.Millisecond)
				mice.LeftClick()
				time.Sleep(6 * time.Millisecond)
				keyboard.KeyPress(uinput.Key5)
				time.Sleep(17 * time.Millisecond)
				mice.LeftClick()
				time.Sleep(6 * time.Millisecond)
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	for { // input loop
		n, err := file.Read(buf)
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		if n != 24 {
			continue
		}

		evType := binary.LittleEndian.Uint16(buf[16:18])
		code := binary.LittleEndian.Uint16(buf[18:20])
		value := binary.LittleEndian.Uint32(buf[20:24])

		if evType == 1 && code == 275 {
			switch value {
			case 1:
				toggle = true
				fmt.Printf("on \r")
			case 0:
				toggle = false
				fmt.Printf("off\r")
			}
		}
	}
}
