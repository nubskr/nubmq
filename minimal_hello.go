package main

import (
    "fmt"
    "log"
    "net"
    "sync/atomic"
    "os"
	"unsafe"
    "sync"
    "strings"
    "runtime"
)

type Message struct {
	data string
	timestamp int64
}

type ValueData struct {
	data string
    mutex sync.RWMutex
}

type Shard struct {
    data []*ValueData
    size int32
}

type KeyManager struct {
    Keys  sync.Map
    // mutex sync.Mutex // for adding new keys
}

type ShardManager struct {
    Shards []*Shard // pointers to shards
    mutex  sync.RWMutex 
}

var ShardSize int32 = 500

// Global variables
var keyManager = KeyManager{
    Keys: sync.Map{},
}

// init a thousand shards
var shardManager = ShardManager{
    Shards: make([]*Shard, 1000),
}

var ShardManagerSizeLim int32 = 1
var curShardManagerSize int32 = 1

var nextIdx int32 = -1
var wg sync.WaitGroup

var curSetCnt int32 = 0
var lastSetCnt int32 = 0

func getNewShard(sz int32) *Shard {
    return &Shard{
        data: make([]*ValueData, sz, ShardSize),
    }
}

func getNewValueData(value string) *ValueData{
    return &ValueData{
        data: value,
    }
}

func _setKey(key string, value string) {    
    idx := int32(696969696)

    if value, ok := keyManager.Keys.Load(key); ok {
        if intValue, ok := value.(int32); ok {
            idx = int32(intValue)
        } else {
            fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO set-x-x-x-x-x-x-x-x-x-x-xx-x-x-x-x-x-x--x",value,"-->")
            os.Exit(1)
        }
    } else {
        val := atomic.AddInt32(&nextIdx, 1)
        keyManager.Keys.Store(key, val)
        idx = val
    }

    if idx == 696969696 {
        fmt.Println("trying to set non existing shit")
        os.Exit(1)
    }

    shardNumber := idx / ShardSize
    localShardIndex := idx%ShardSize 

    if localShardIndex == 0 {
        atomic.AddInt32(&curShardManagerSize, 1)
    }

    fmt.Println("setting key",key,"at",idx,"at shard number",shardNumber,"at local index",localShardIndex)
    
    fmt.Println("trying to acquire lock to set key")


	newVal := getNewValueData(value)


    shardManager.mutex.Lock()


	// TODO: fix the below shit, it should not be this way
    // fmt.Println("set worker locked acquired")

    // if shardNumber >= int32(len(shardManager.Shards)) {
    //     os.Exit(1)
    // }

    shard := shardManager.Shards[shardNumber]

	shardManager.mutex.Unlock()

	fmt.Println("set worker locked released")

	// value is a darn string
	atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.data[localShardIndex])), unsafe.Pointer(newVal))
}

func _getKey(key string) (string, bool) {

    idx := int32(696969696) // TODO: remove this shit

    if value, ok := keyManager.Keys.Load(key); ok {
        if intValue, ok := value.(int32); ok {
            idx = int32(intValue)
        } else {

            fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO get-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x",value,"-->")
            os.Exit(1)
        }
    } else {
        return "NaN", false
    }

    if idx == 696969696 {
        fmt.Println("trying to get non existing shit")
        os.Exit(1)
    }

    shardNumber := idx / ShardSize

    shardManager.mutex.Lock()

    if shardNumber >= int32(len(shardManager.Shards)) {
        shardManager.mutex.Unlock()
        return "", false
    }

    shard := shardManager.Shards[shardNumber]
    localShardIndex := idx % ShardSize
    
    if localShardIndex < int32(len(shard.data)) {
        shardManager.mutex.Unlock()
        return (shard.data[localShardIndex]).data, true
    }

    shardManager.mutex.Unlock()
    return "NaN", false
}

func handleConnection(conn net.Conn) {
    fmt.Println("Client connected")
    buffer := make([]byte, 1024)
    for {
        length, err := conn.Read(buffer)
        
        if err != nil {
            fmt.Println("an error occured while reading message:",err)
            return
        }

        data := string(buffer[:length])

        stringData := strings.Fields(data)

        if stringData[0] == "SET"{
            atomic.AddInt32(&curSetCnt, 1)
            _setKey(stringData[1],stringData[2])
            _, err := conn.Write([]byte(fmt.Sprint("SET done\n")))

            if err != nil {
                log.Println("failed to reply message:" ,err)
            } else{
            }
        } else{
            output ,exists := _getKey(stringData[1])

            if exists {

            }
            _, err := conn.Write([]byte(fmt.Sprint(output+"\n"))) // Send message over connection

            if err != nil {
            } else{
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

	// TODO: remove this shit
    for i := 0 ; i < 1000 ; i++ {
        shardManager.Shards[i] = getNewShard(ShardSize)
    }


	fmt.Println("Server listening on :8080")

    for {
        // Accept connection
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }

        go handleConnection(conn)   
    }
}