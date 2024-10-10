package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "sync/atomic"
    "sync"
    "strings"
    "runtime"
)

/*
TODO: 

- dynamic shard initializations (double when full) ; O(1) asymptomatic time in the end
(put a mutex lock only while it is being copied, make sure no changes occur!!!)

writes to new shards are blocked while resizing is happening

how does this changes reads ?

make copy in background and just change the pointer to shardmanager when its done, nothing is darn blocked then
(the memory usage doubles for a second there, but then the garbage collector does its job)

Updated scene:

    Shard:
        []*ValueObject
        shard_size int32

    shardmanager: 
        []*Shard
        shardmanager_size int32
        current_
*/

/*
TODO:
- expiry for a key ? (just make another go routine which cleans shit ? no please, lmao)

- Key eviction (why we need this: what if shit gets full, how will we deal with it then ? increase the darn memory ffs, what else would you do)

*/

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
    Keys  map[string]int32
    mutex sync.Mutex // for adding new keys
}

type ShardManager struct {
    Shards []*Shard // pointers to shards
    mutex  sync.RWMutex 
}

// Hyperparameter
var ShardSize int32 = 1

// Global variables
var keyManager = KeyManager{
    Keys: make(map[string]int32),
}

var shardManager = ShardManager{
    Shards: make([]*Shard, 1),
}

var connectionChan = make(chan net.Conn)
var messageChan = make(chan Message)
var ShardManagerSizeLim = make(chan int32)
var curShardManagerSize = make(chan int32)

var nextIdx int32 = -1

func getNewShard(sz int32) *Shard {
    fmt.Println("making a new shard of size: ",sz)
    return &Shard{
        data: make([]*ValueData, sz, ShardSize),
    }
}

func getNewValueData(value string) *ValueData{
    return &ValueData{
        data: value,
    }
}

func resizeShardManagerWorker(addSize int32,curSize int32,curShardManagerSizeLim int32){
    fmt.Println("Starting resizing")
    shardManager.mutex.Lock()

    newShards := shardManager

    lenn := len(shardManager.Shards)
    for i := len(shardManager.Shards); i < lenn + int(addSize) ; i++ {
        // newShards.Shard[i] = getNewShard(ShardSize)
        fmt.Println("burrrrrrr")
        newShards.Shards = append(newShards.Shards,getNewShard(ShardSize))
    }

    shardManager = newShards

    shardManager.mutex.Unlock()
    fmt.Println("resizing done")
    curShardManagerSize <- curSize
    fmt.Println("curshardmanagersize updated")
    ShardManagerSizeLim <- curShardManagerSizeLim
    fmt.Println("end of resizing function")
}

func resizeShardManager(){
    for {
        fmt.Println(">-==-==-==-==-==-==-==-==-==-==-==-==-==-==-==-==-<")
        curShardManagerSizeLim := <- ShardManagerSizeLim
        
        curSize := <- curShardManagerSize

        fmt.Println("start of resizing function")

        if curSize == curShardManagerSizeLim {
            fmt.Println("triggering resizing")
            addSize := curShardManagerSizeLim
            curShardManagerSizeLim *= 2
            go resizeShardManagerWorker(addSize,curSize,curShardManagerSizeLim)
        } else {
            fmt.Println("Nothin to do here")
        }
    }
}

func _setKey(key string, value string) {

    // this thing accesses the whole darn shard manager and puts a lock on it, which is not good, very very bad
    // we are accessing the length of the number of shards and that is not what we want, we need another way to access shit man, this is bad, very very bad
    
    // Lock keyManager to ensure thread safety for adding keys

    // fmt.Println("inside the set function")
    keyManager.mutex.Lock()
    idx, exists := keyManager.Keys[key]
    if !exists {
        fmt.Println("not exists here!!")
        val := atomic.AddInt32(&nextIdx, 1)
        keyManager.Keys[key] = val
        idx = val
    }
    keyManager.mutex.Unlock()

    shardNumber := idx / ShardSize
    localShardIndex := idx%ShardSize 
    // 0,1,2,3,4,0,1,2,3,4....

    if localShardIndex == 0 {
        // at the start to each shard, just increase the current size 

        fmt.Println("start0")

        // flush the channel
        tmp := <- curShardManagerSize
        
        fmt.Println("start1")
        
        // update the channel
        curShardManagerSize <- tmp + 1

        fmt.Println("start2")
    }
    fmt.Println("end")

    fmt.Println("setting key",key,"at",idx,"at shard number",shardNumber,"at local index",localShardIndex)
    
    // Lock shardManager to ensure thread safety for adding shards
    // fmt.Println("before locking")

    shardManager.mutex.Lock()

    fmt.Println("locked now")

    fmt.Println(shardManager.Shards)

    if shardNumber >= int32(len(shardManager.Shards)) {
        // this is not good, we make it happen on its own!!
        fmt.Println("help me dadddy, I feel bad about this")
        // go resizeShardManager()
        // os.Exit(1)
        // fmt.Errorf("nooooo senpaiiiiiiiiii, this is notttttt goooood")
    }

    fmt.Println("before")
    shard := shardManager.Shards[shardNumber]
    newVal := getNewValueData(value)
    fmt.Println("after")

    if int32(len(shard.data)) > localShardIndex{
        fmt.Println("hi mom",shard.data)
        shard.data[localShardIndex] = newVal
    } else {
        fmt.Println("oh nu, hewp me daddy")
    }

    shardManager.mutex.Unlock()
    fmt.Println("Unlocked now")
}

func _getKey(key string) (string, bool) {
    // Read from keyManager without locking
    idx, exists := keyManager.Keys[key]
    if !exists {
        return "NaN", false
    }

    shardNumber := idx / ShardSize

    shardManager.mutex.Lock()
    defer shardManager.mutex.Unlock()

    if shardNumber >= int32(len(shardManager.Shards)) {
        // fmt.Println("first off")
        return "", false
    }

    shard := shardManager.Shards[shardNumber]
    localShardIndex := idx % ShardSize
    
    // fmt.Println(localShardIndex,*shard.data[1])

    if localShardIndex < int32(len(shard.data)) {
        return (shard.data[localShardIndex]).data, true
    }

    // fmt.Println("second off")
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
                    // fmt.Println("echoed message: ",msg.data)
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

        fmt.Println("we have a new message",data)
        // output := "NaN" 

        if stringData[0] == "SET"{
            // fmt.Println("Trying to set shit")
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
    }
    conn.Close()
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    ln, err := net.Listen("tcp", ":8080")

    if err != nil {
        log.Fatal(err)
    }

    go listener()
    go resizeShardManager()

    
    ShardManagerSizeLim <- 1
    curShardManagerSize <- 1

    shardManager.Shards[0] = getNewShard(ShardSize)


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