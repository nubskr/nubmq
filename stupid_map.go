package main

import (
    "fmt"
    "log"
    "net"
    "runtime"
    "strings"
    "sync"
)

var myMap sync.Map

func _setKey(key string, value interface{}) {
    myMap.Store(key, value)
}

func _getKey(key string) (interface{}, bool) {
    return myMap.Load(key)
}

func handleConnection(conn net.Conn) {
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
            _setKey(stringData[1], stringData[2])
            _, err := conn.Write([]byte("SET done\n"))
            if err != nil {
                log.Println("Failed to reply message:", err)
            }
        } else {
            output, exists := _getKey(stringData[1])
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

func main() {
    fasttttt := true
    // fasttttt = false

    if fasttttt {
        runtime.GOMAXPROCS(runtime.NumCPU())
    }

    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Server listening on :8080")

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        go handleConnection(conn)
    }
}
