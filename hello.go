package main

import (
    "fmt"
    "log"
    "net"
    "time"
)

type Message struct {
	data string
	timestamp int64
}

func handleConnection(conn net.Conn,nubskr *[]Message) {
    fmt.Println("Client connected")
    buffer := make([]byte, 1024)
	// fmt.Println(buffer)
	// an empty buffer ?
    for {
        // Read data from the connection
        length, err := conn.Read(buffer)
        if err != nil {
            fmt.Println("Client disconnected")
            return
        }
        data := string(buffer[:length])
        fmt.Printf("Received: %s\n", data)
        message := Message {
            data: data,
            timestamp: time.Now().Unix(),
        }
        *nubskr = append(*nubskr,message)

        fmt.Println(*nubskr)
        // Send a response back to the client
        _, err = conn.Write([]byte(fmt.Sprint((*nubskr)[0].data," ", (*nubskr)[0].timestamp)))
        if err != nil {
            fmt.Println("Error sending response")
            return
        }
    }
}

func main() {
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }

	message := Message{
		data: "this is message one",
		timestamp: 1111,
	}

    // now we need to store the pointer of these messages to something

    v := []Message{} 
    v = append(v,message)

    fmt.Println(v)
	
	fmt.Println("Server listening on :8080")

    for {
        // Accept a connection
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }

        // Handle the connection
        handleConnection(conn,&v)

        fmt.Println("final stuff", v)

        conn.Close() // Close the connection after handling
    }
}