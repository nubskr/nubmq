package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	defer log.Print("Connection terminated due to inactivity")
	defer activeConns.Store(conn.RemoteAddr(), false)
	fmt.Println("routine launched")
	buffer := make([]byte, 1024)

	for {
		err := conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			// return 0, fmt.Errorf("failed to set read deadline: %w", err)
			log.Fatal("Failed to set read deadline")
		}

		length, err := conn.Read(buffer)

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// exit the goroutine, connection has been dead for a while
			// log.Fatal("go routine terminated")
			return
		}

		if err != nil {
			fmt.Println("an error occured while reading message:", err)
			return
		}

		data := string(buffer[:length])

		stringData := strings.Fields(data)

		if stringData[0] == "SET" {
			curReq := SetRequest{
				key:    stringData[1],
				value:  stringData[2],
				status: make(chan struct{}),
			}

			allowSets.Lock()
			SetWG.Add(1)
			allowSets.Unlock()
			setQueue <- curReq

			select {
			case <-curReq.status:
			case <-time.After(2 * time.Second): // Timeout in case of delay
				log.Fatal("BAD WORKER, SET REQUEST TIMED OUT FOR KEY: ", curReq.key)
			}

			_, err := conn.Write([]byte(fmt.Sprint("SET done\n")))

			if err != nil {
				log.Println("failed to reply message:", err)
			} else {
			}
		} else {
			output, exists := _getKey(stringData[1])

			if exists {

			}
			_, err := conn.Write([]byte(fmt.Sprint(output + "\n"))) // Send message over connection

			if err != nil {
			} else {
			}
		}
	}
}
