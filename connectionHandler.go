package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func WriteStuffToConn(conn net.Conn, stuff string) {
	_, err := conn.Write([]byte(fmt.Sprint(stuff + "\n")))

	if err != nil {
		log.Print("failed to reply message:", err)
	}
}

var appendQueue chan int64 = make(chan int64, 50000000) // just don't block ffs

func appendToLog() {

	logPath := "./analysing-stuff/top_requests.csv"

	// Create or truncate the file and write header

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	if err != nil {

		log.Fatal("Failed to open log file:", err)

	}

	_, err = f.WriteString("Timestamp_ms\n")

	if err != nil {

		log.Fatal("Failed to write header:", err)

	}

	f.Close()

	// Open file in append mode for subsequent writes

	f, err = os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {

		log.Fatal("Failed to open log file for appending:", err)

	}

	defer f.Close()

	// Write timestamps as they come in

	for {

		appendData := <-appendQueue

		_, err := f.WriteString(fmt.Sprintf("%d\n", appendData))

		if err != nil {

			log.Printf("Failed to append to log: %v", err)

		}

		// Ensure data is written to disk

		// f.Sync()

	}

}

func handleConnection(conn net.Conn) {
	fmt.Println("Client connected")
	buffer := make([]byte, 1024)

	writeChanPrimary := make(chan string, 10)
	writeChanSecondary := make(chan string, 100)

	// Writer
	go func(conn net.Conn) {
		// we have a hierarchy, like, SET and GET replies get higher priority than say Event Notifications
		for {
			select {
			case val := <-writeChanPrimary:
				WriteStuffToConn(conn, val)
			case val := <-writeChanSecondary:
				WriteStuffToConn(conn, val)
			}
		}
	}(conn)

	for {
		length, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("client disconnected")
			return
		}

		data := string(buffer[:length])
		log.Print("Received from client: ", data)

		stringData := strings.Fields(data)
		if len(stringData) == 0 {
			continue
		}

		/*
			SET key value EX time_in_seconds
			 0	 1    2	    3	  4
		*/

		if stringData[0] == "SET" && len(stringData) > 2 {

			curReq := SetRequest{
				key:       stringData[1],
				value:     stringData[2],
				canExpire: false,
				TTL:       time.Now().Unix(), // just let it be for now
				status:    make(chan struct{}),
			}

			entry := Entry{
				key:   stringData[1],
				value: stringData[2],
			}

			EventQueue <- entry

			if len(stringData) == 5 {

				parsedTime, err := strconv.ParseInt(stringData[4], 10, 64)

				log.Print("========Parsed time is: ", parsedTime)
				if err != nil {
					log.Fatal("Error parsing time:", err)
				}

				// canExpire     bool
				// TTL           int64
				// isExpiryEvent bool
				entry.canExpire = true
				entry.TTL = parsedTime

				SetContainer.queue <- entry // handles TTL notifications

				curReq = SetRequest{
					key:       stringData[1],
					value:     stringData[2],
					canExpire: true,
					TTL:       parsedTime,
					status:    make(chan struct{}),
				}
			}

			allowSets.Lock()
			SetWG.Add(1)
			allowSets.Unlock()
			setQueue <- curReq

			select {
			case <-curReq.status:
				appendQueue <- time.Now().UnixMilli()
			case <-time.After(100 * time.Second): // Timeout in case of delay
				// log.Fatal("BAD WORKER, SET REQUEST TIMED OUT FOR KEY: ", curReq.key)
			}

			writeChanPrimary <- "SET done"

		} else if stringData[0] == "GET" {
			res, exists := _getKey(stringData[1])
			appendQueue <- time.Now().UnixMilli()
			if exists {
				writeChanPrimary <- res
			} else {
				writeChanPrimary <- "(nil)"
				// log.Fatal("tf just happened here")
			}
		} else if stringData[0] == "SUBSCRIBE" {
			key := stringData[1]

			SubscribersMutex.Lock()
			Subscribers[key] = append(Subscribers[key], &writeChanSecondary)
			SubscribersMutex.Unlock()
			writeChanSecondary <- "SUBSCRIBED TO CHANNEL"
		} else {
			writeChanPrimary <- "invalid command"
		}
	}
}
