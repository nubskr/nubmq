package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	fmt.Println("Client connected")
	buffer := make([]byte, 1024)
	for {
		length, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("an error occured while reading message:", err)
			return
		}

		data := string(buffer[:length])

		stringData := strings.Fields(data)

		if stringData[0] == "SET" {
			// I DON'T LIKE THIS AT ALL, THIS POLLING IS SLOWING US DOWN, THIS IS FUCKING EXISTENTIAL
			for {
				HaltSetsMutex.RLock()
				if HaltSets == 0 {
					HaltSetsMutex.RUnlock()
					break
				}
				HaltSetsMutex.RUnlock()

				time.Sleep(1 * time.Microsecond)
			}

			SetWG.Add(1)
			_setKey(stringData[1], stringData[2])
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
