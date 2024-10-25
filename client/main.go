package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

func main() {
	CONNECT := "localhost:8080"

	s, err := net.ResolveUDPAddr("udp4", CONNECT)
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
	defer c.Close()
	c.Write([]byte("first init connection \n"))

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			data := []byte(input + "\n")
			_, err = c.Write(data)
			if strings.TrimSpace(string(data)) == "STOP" {
				fmt.Println("Exiting UDP client!")
				os.Exit(0)
			}

			if err != nil {
				fmt.Println(err)
				return
			}
		}
		wg.Done()
	}()

	go func() {
		for {
			buffer := make([]byte, 1024)
			n, _, err := c.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println(err)
				return
			}
			msg := string(buffer[0:n])

			fmt.Printf("%s", Transform(msg, c.LocalAddr().String()))
		}
		wg.Done()
	}()

	wg.Wait()
}

func Transform(input string, ip string) string {
	// Regular expression to match the IP address and port
	pattern := `^(\d{1,3}(?:\.\d{1,3}){3}:\d{1,5})`
	re := regexp.MustCompile(pattern)

	// Find the matches
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		// Replace the extracted string with "you"
		if match[1] == ip {
			replaced := re.ReplaceAllString(input, "you")
			return replaced
		}
		return input
	} else {
		return input
	}
}
