package main

import (
    "fmt"
    "log"
    "net"
    // "time"
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
    Keys  sync.Map
    mutex sync.Mutex // for adding new keys
}

type ShardManager struct {
    Shards []*Shard // pointers to shards
    mutex  sync.RWMutex 
}

// Hyperparameter
var ShardSize int32 = 5

// Global variables
var keyManager = KeyManager{
    Keys: sync.Map{},
}

var shardManager = ShardManager{
    Shards: make([]*Shard, 1),
}

var ShardManagerSizeLim = make(chan int32,1)
var curShardManagerSize = make(chan int32,1)

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
    fmt.Println("resize worker lock acquired")

    // newShards := shardManager

    lenn := len(shardManager.Shards)
    for i := len(shardManager.Shards); i < lenn + int(addSize) ; i++ {
        // newShards.Shard[i] = getNewShard(ShardSize)
        fmt.Println("burrrrrrr")
        shardManager.Shards = append(shardManager.Shards,getNewShard(ShardSize))
    }

    // shardManager = newShards

    fmt.Println("resizing done")
    curShardManagerSize <- curSize
    fmt.Println("curshardmanagersize updated")
    ShardManagerSizeLim <- curShardManagerSizeLim
    fmt.Println("end of resizing function")
    shardManager.mutex.Unlock()
    fmt.Println("resize worker lock released")   
}

func resizeShardManager(){
    for {
        // fmt.Println(">-==-==-==-==-==-==-==-==-==-==-==-==-==-==-==-==-<")
        
        curShardManagerSizeLim := <- ShardManagerSizeLim
        
        curSize := <- curShardManagerSize

        // fmt.Println("start of resizing function")
        // fmt.Println("current size is ",curSize)

        buffer := int32(50) // TODO: this is bullshit, don't rely on this

        if curSize >= curShardManagerSizeLim - buffer {
            fmt.Println("triggering resizing")
            addSize := curShardManagerSizeLim
            curShardManagerSizeLim *= 2
            go resizeShardManagerWorker(addSize,curSize,curShardManagerSizeLim)
            
            // block here untill the above go routine completes
            // fmt.Println("resizing triggered")
        } else {
            // fmt.Println("Nothin to do here")
            curShardManagerSize <- curSize // blocked untill someone listens to this shit
            ShardManagerSizeLim <- curShardManagerSizeLim
        }
    }
}

func _setKey(key string, value string) {

    // this thing accesses the whole darn shard manager and puts a lock on it, which is not good, very very bad
    // we are accessing the length of the number of shards and that is not what we want, we need another way to access shit man, this is bad, very very bad
    
    // Lock keyManager to ensure thread safety for adding keys

    // fmt.Println("inside the set function")
    // keyManager.mutex.Lock()
    // idx, exists := keyManager.Keys[key]
    // if !exists {
    //     fmt.Println("not exists here!!")
        // val := atomic.AddInt32(&nextIdx, 1)
        // // keyManager.Keys[key] = val
        // keyManager.Keys.Store(key, val)
        // idx = val
    // }
    // keyManager.mutex.Unlock()

    idx := int32(6969)

    if value, ok := keyManager.Keys.Load(key); ok {
        if intValue, ok := value.(int32); ok { // Type assertion to int
            idx = int32(intValue)
            // fmt.Println("Loaded:", intValue) // Uncomment to print the loaded value
        } else {
            fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO -x-x-x-x-x-x-x-x-x-x-xx-x-x-x-x-x-x--x",value,"-->")
        }
        // idx = value
        // fmt.Println("Loaded:", value) // Will print: Loaded: 42
    } else {
        // fmt.Println("Key does not exist.")
        val := atomic.AddInt32(&nextIdx, 1)
        // keyManager.Keys[key] = val
        // keyManager.Keys.Store(key, val)
        keyManager.Keys.Store(key, val)
        idx = val
    }

    shardNumber := idx / ShardSize
    localShardIndex := idx%ShardSize 
    // 0,1,2,3,4,0,1,2,3,4....

    if localShardIndex == 0 {
        // at the start to each shard, just increase the current size 

        fmt.Println("start0")

        
        tmp := <- curShardManagerSize
        
        fmt.Println("start1")
        
        // update the channel

        curShardManagerSize <- tmp + 1
        
        // } we want this encapsulated shit to happen exactly at once!!
        fmt.Println("start2")
    }
    fmt.Println("end")

    fmt.Println("setting key",key,"at",idx,"at shard number",shardNumber,"at local index",localShardIndex)
    
    // Lock shardManager to ensure thread safety for adding shards
    // fmt.Println("before locking")

    // TODO: we can't acquite lock here at some times for some reason, why ?

    fmt.Println("trying to acquire lock to set key")


    shardManager.mutex.Lock()

    fmt.Println("set worker locked acquired")

    fmt.Println(shardManager.Shards)

    if shardNumber >= int32(len(shardManager.Shards)) {
        //TODO: we have accidentaly stumbled upon this thing first before the resize worker could trigger, don't worry, just manually trigger the resize worker or maybe skip the iteration so that the worker can get triggered (bad idea) 

        // this is not good, we make it happen on its own!!
        // fmt.Println("help me dadddy, I feel bad about this")
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
    fmt.Println("set worker locked released")

    // fmt.Println("Unlocked now")
}

func _getKey(key string) (string, bool) {
    // Read from keyManager without locking

    // idx, exists := keyManager.Keys[key]
    // if !exists {
    //     return "NaN", false
    // }

    idx := int32(6969)

    if value, ok := keyManager.Keys.Load(key); ok {
        if intValue, ok := value.(int32); ok { // Type assertion to int
            idx = int32(intValue)
            // fmt.Println("Loaded:", intValue) // Uncomment to print the loaded value
        } else {

            fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO -x-x-x-x-x-x-x-x-x-x-xx-x-x-x-x-x-x--x",value,"-->")
            // fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO -x-x-x-x-x-x-x-x-x-x-xx-x-x-x-x-x-x--x")
        }
    } else {
        // fmt.Println("Key does not exist.")
        // keyManager.Keys.Store(key, val)
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

func handleConnection(conn net.Conn) {
    fmt.Println("Client connected")
    buffer := make([]byte, 1024)
    for {
        fmt.Println("START")
        // Read data from the connection
        length, err := conn.Read(buffer)
        
        if err != nil {
            log.Fatal(err)
            return
        }

        fmt.Println("WE HAVE SOMETHING",length)

        data := string(buffer[:length])

        stringData := strings.Fields(data)

        fmt.Println("we have a new message",data)

        if stringData[0] == "SET"{
            _setKey(stringData[1],stringData[2])
            _, err := conn.Write([]byte(fmt.Sprint("SET done\n"))) // Send message over the connection

            if err != nil {
                log.Println("failed to reply message:" ,err)
                // Handle error (e.g., log it, remove the connection, etc.)
            } else{
                fmt.Println("replied message: ","output")
            }
        } else{
            // fmt.Println("Trying to get shit")
            output , exists := _getKey(stringData[1])

            fmt.Println(exists)

            _, err := conn.Write([]byte(fmt.Sprint(output+"\n"))) // Send message over the connection

            if err != nil {
                log.Println("failed to reply message:" ,err)
                // Handle error (e.g., log it, remove the connection, etc.)
            } else{
                fmt.Println("replied message: ",output)
            }
        }
        fmt.Println("END")
    }
    // conn.Close()
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    ln, err := net.Listen("tcp", ":8080")

    if err != nil {
        log.Fatal(err)
    }

    // go listener()
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