package main

import (
    "fmt"
    "log"
    "net"
    "time"
    // "sync"
    // "runtime"
)

type Message struct {
	data string
	timestamp int64
}

var connectionChan = make(chan net.Conn)
var messageChan = make(chan Message)

func listener(){
    // always listening
    fmt.Println("listener started")

    var connections []net.Conn

    for {
        select{
        case conn := <- connectionChan:
            connections = append(connections,conn)

        case msg := <- messageChan:
            for _,conn:= range connections{
                _, err := conn.Write([]byte(fmt.Sprint(msg.data))) // Send message over the connection

                if err != nil {
                    log.Fatal("failed to echo message:" ,err)
                    // Handle error (e.g., log it, remove the connection, etc.)
                } else{
                    fmt.Println("echoed message: ",msg.data)
                }
            }
        }
    }
}

func handleConnection(conn net.Conn) {
    fmt.Println("Client connected")
    buffer := make([]byte, 1024)
    connectionChan <- conn
    for {
        // Read data from the connection
        length, err := conn.Read(buffer)
        if err != nil {
            log.Fatal(err)
            return
        }

        data := string(buffer[:length])
        
        message := Message {
            data: data,
            timestamp: time.Now().Unix(),
        }

        messageChan <- message
        
        fmt.Println("received message: ",message)
        // _, err = conn.Write([]byte(fmt.Sprint("helo wurld")))
        
        // if err != nil {
        //     log.Fatal(err)
        //     return
        // }
    }
    conn.Close()
}

func main() {
    // runtime.GOMAXPROCS(runtime.NumCPU())

    ln, err := net.Listen("tcp", ":8080")

    if err != nil {
        log.Fatal(err)
    }

    go listener()

	fmt.Println("Server listening on :8080")

    for {
        // Accept a connection
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }

        go handleConnection(conn)   
    }
}