package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

var myMap sync.Map

func __setKey(key string, value interface{}) {
	myMap.Store(key, value)
}

func __getKey(key string) (interface{}, bool) {
	return myMap.Load(key)
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
		} else {
			output, exists := __getKey(stringData[1])
			if !exists {
				output = "Key not found"
			}
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
