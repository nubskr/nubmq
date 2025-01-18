package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
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
		log.Print("Received from client: ", data)

		stringData := strings.Fields(data)

		/*
			SET key value EX time_in_seconds
			 0	 1    2	    3	  4
		*/

		if stringData[0] == "SET" {
			curReq := SetRequest{
				key:       stringData[1],
				value:     stringData[2],
				canExpire: false,
				TTL:       time.Now().Unix(), // just let it be for now
				status:    make(chan struct{}),
			}
			if len(stringData) == 5 {

				parsedTime, err := strconv.ParseInt(stringData[4], 10, 64)

				log.Print("========Parsed time is: ", parsedTime)
				if err != nil {
					log.Fatal("Error parsing time:", err)
				}

				curReq = SetRequest{
					key:       stringData[1],
					value:     stringData[2],
					canExpire: true,
					TTL:       parsedTime,
					status:    make(chan struct{}),
				}
			} else {
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
			_, err := conn.Write([]byte(fmt.Sprint(output + "\n")))

			if err != nil {
			} else {
			}
		}
	}
}
