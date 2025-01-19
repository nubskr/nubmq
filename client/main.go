package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func reader(conn net.Conn, msgQueue chan string) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Fatal("failed to read data from conn:", err)
		}
		msgData := string(buffer[:n])
		msgQueue <- msgData
	}
}

func ShowMessages(msgQueue chan string) {
	for {
		data := <-msgQueue
		log.Print("-> ", data)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	msgQueue := make(chan string)
	go reader(conn, msgQueue)
	go ShowMessages(msgQueue)

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading input:", err)
		}

		rawMsg := string([]byte(input))
		stringData := strings.Fields(rawMsg)

		stuff := string([]byte(input))

		if len(stringData) == 5 {
			intValue, err := strconv.Atoi(stringData[4])
			if err != nil {
				log.Fatal("Error converting string to integer:", err)
			} else {
			}

			fmeBro := strconv.FormatInt(int64(time.Now().Unix()+int64(intValue)), 10)

			finalString := ""

			for i := 0; i < 4; i++ {
				finalString += stringData[i] + " "
			}

			finalString += fmeBro
			stuff = finalString
		}

		log.Print("sending this shit: ", stuff)

		_, err = conn.Write([]byte(stuff))
		if err != nil {
			log.Fatal("Error writing to connection:", err)
		}
	}
}
