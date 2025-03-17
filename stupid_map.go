package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var myMap sync.Map

func __setKey(key string, value interface{}) {
	myMap.Store(key, value)
}

func __getKey(key string) (interface{}, bool) {
	return myMap.Load(key)
}

var _appendQueue chan int64 = make(chan int64, 50000000) // just don't block ffs

func _appendToLog() {
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
		appendData := <-_appendQueue
		_, err := f.WriteString(fmt.Sprintf("%d\n", appendData))
		if err != nil {
			log.Printf("Failed to append to log: %v", err)
		}
		// Ensure data is written to disk
		// f.Sync()
	}
}

func __handleConnection(conn net.Conn) {
	fmt.Println("Client connected")
	buffer := make([]byte, 1024)
	for {
		length, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("An error occurred while reading message:", err)
			return
		}

		data := string(buffer[:length])
		stringData := strings.Fields(data)

		if stringData[0] == "SET" {
			__setKey(stringData[1], stringData[2])
			_, err := conn.Write([]byte("SET done\n"))
			if err != nil {
				log.Println("Failed to reply message:", err)
			}
			_appendQueue <- time.Now().UnixMilli()
		} else {
			output, exists := __getKey(stringData[1])
			if !exists {
				output = "Key not found"
			}
			_appendQueue <- time.Now().UnixMilli()
			_, err := conn.Write([]byte(fmt.Sprint(output, "\n")))
			if err != nil {
				log.Println("Failed to send message:", err)
			}
		}
	}
}

// func main() {
// 	fasttttt := true
// 	// fasttttt = false

// 	if fasttttt {
// 		runtime.GOMAXPROCS(runtime.NumCPU())
// 	}

// 	go _appendToLog()

// 	ln, err := net.Listen("tcp", ":8080")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("Server listening on :8080")

// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 		go __handleConnection(conn)
// 	}
// }
