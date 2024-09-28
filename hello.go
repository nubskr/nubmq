package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "sync/atomic"
    "sync"
    "strings"
    // "runtime"
)

type Message struct {
	data string
	timestamp int64
}

type Shard struct {
    // TODO: make the data a []customobject, so that you can just update that object instead of locking down the whole shard
    data []string
    size int32
}

type KeyManager struct {
    Keys  map[string]int32
    mutex sync.Mutex // for adding new keys
}

type ShardManager struct {
    Shards []*Shard // pointers to shards
    mutex  sync.Mutex 
}

// Hyperparameter
var ShardSize int32 = 2

// Global variables
var keyManager = KeyManager{
    Keys: make(map[string]int32),
}

var shardManager = ShardManager{
    Shards: make([]*Shard, 0),
}

var connectionChan = make(chan net.Conn)
var messageChan = make(chan Message)
var nextIdx int32 = 0 

func getNewShard() *Shard {
    return &Shard{
        data: make([]string, 0, ShardSize),
        size: 0,
    }
}

func _setKey(key string, value string) {
    // Lock keyManager to ensure thread safety for adding keys
    keyManager.mutex.Lock()
    idx, exists := keyManager.Keys[key]
    if !exists {
        fmt.Println("not exists here!!")
        keyManager.Keys[key] = val
        idx = val
        val := atomic.AddInt32(&nextIdx, 1)
    }
    keyManager.mutex.Unlock()

    shardNumber := (idx + ShardSize - 1) / ShardSize
    localShardIndex := idx%ShardSize 

    fmt.Println("setting key",key,"at",idx,"at shard number",shardNumber,"at local index",localShardIndex)
    
    // Lock shardManager to ensure thread safety for adding shards
    shardManager.mutex.Lock()
    defer shardManager.mutex.Unlock()

    for shardNumber >= int32(len(shardManager.Shards)) {
        shardManager.Shards = append(shardManager.Shards, getNewShard())
    }

    shard := shardManager.Shards[shardNumber]

    if int32(len(shard.data)) > localShardIndex{
        shard.data[localShardIndex] = value
    } else {
        shard.data = append(shard.data, value)
    }
}

func _getKey(key string) (string, bool) {
    // Read from keyManager without locking
    idx, exists := keyManager.Keys[key]
    if !exists {
        return "NaN", false
    }

    shardNumber := (idx + ShardSize - 1 ) / ShardSize

    shardManager.mutex.Lock()
    defer shardManager.mutex.Unlock()

    if shardNumber >= int32(len(shardManager.Shards)) {
        return "", false
    }

    shard := shardManager.Shards[shardNumber]
    localShardIndex := idx % ShardSize
    
    if localShardIndex < int32(len(shard.data)) {
        return shard.data[localShardIndex], true
    }

    // shard does not exist, should never reach here!!!!
    return "NaN", false
}

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

        stringData := strings.Fields(data)

        // output := "NaN" 

        if stringData[0] == "SET"{
            _setKey(stringData[1],stringData[2])
        } else{
            output , exists := _getKey(stringData[1])

            fmt.Println(exists)

            _, err := conn.Write([]byte(fmt.Sprint(output))) // Send message over the connection

            if err != nil {
                log.Fatal("failed to reply message:" ,err)
                // Handle error (e.g., log it, remove the connection, etc.)
            } else{
                fmt.Println("replied message: ",output)
            }
        }

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