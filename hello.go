package main

import (
    "fmt"
    "log"
    "net"
)

type Nubskr struct {
	name string
	age int32
	message string
	arr []int
}

func(p Nubskr) one() string {
	return "one"
}

func handleConnection(conn net.Conn,nubskr []Nubskr) {
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
        fmt.Printf("Received: %s\n", string(buffer[:length]))

        // Send a response back to the client
        _, err = conn.Write([]byte(fmt.Sprint(nubskr[0].arr)))
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
 
	v := []Nubskr{} 

	oneNubskr := Nubskr{
		name: "lmao",
		age: 69,
		message: "test - message",
		arr: []int{1,2,3},
	}

	v = append(v,oneNubskr)
	
	// fmt.Println(v)
	
	v[0].arr = append(v[0].arr,23)

	fmt.Println("Server listening on :8080")

    for {
        // Accept a connection
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }

        // Handle the connection
        handleConnection(conn,v)
        conn.Close() // Close the connection after handling
    }
}