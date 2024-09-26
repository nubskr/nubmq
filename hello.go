package main

/*

we use a freakin map

last = 0
put a mutex on last


Map[key string] -> last + 1

last++

then just insert the value at that 
*/

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

/*
the keys are just a straight off map, we can randomly access them, from the index of a key, we can find the shard number and position in which we need to find the data in that shard



*/

type Shard struct {
    size int32
    data []string
}

func setKey(sz int32,data []string){
    // the keys and values are just a string bro

    /*
    - Check if the key exists in `keys`, if yes then just update it
    - If it doesnt, then just insert a new one



    */


    // insert a new one

    /*
    - Append to Keys

    - check if the latest shard has some space or not
    
    - 
    */



}

func getKey(sz int32,data []string){
    // the keys and values are just a string bro


}



var connectionChan = make(chan net.Conn)
var messageChan = make(chan Message)
var []string Keys

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