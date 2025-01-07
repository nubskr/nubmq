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
			// handleSets(stringData[1], stringData[2])
			curReq := SetRequest{
				key:    stringData[1],
				value:  stringData[2],
				status: make(chan struct{}),
			}

			// wait for the HaltSet mode to be over
			// HaltSetsMutex.Lock()
			// for HaltSets == 1 {
			// 	log.Print("SETS HALTED")
			// 	HaltSetcond.Wait()
			// }

			// HaltSetsMutex.Unlock()

			allowSets.Lock()
			SetWG.Add(1)
			setQueue <- curReq
			allowSets.Unlock()

			// wait for the acknowledgement for this request
			select {
			case <-curReq.status:
			// fmt.Printf("Producer %d: Task %d processed!\n")
			case <-time.After(1 * time.Second): // Timeout in case of delay
				// fmt.Printf("Producer %d: Task %d processing timeout!\n", id, i)
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
