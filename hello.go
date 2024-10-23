package main

import (
    "fmt"
    "log"
    "net"
    "time"
    "sync/atomic"
    "os"
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

*/

/*
TODO:
- expiry for a key ? (just make another go routine which cleans shit ? no please, lmao)

- Key eviction (why we need this: what if shit gets full, how will we deal with it then ? increase the darn memory ffs, what else would you do)
*/


/*
# Dynamic buffer changes:

keep track of last x growth values, keep x configurable

update buffer as the moving average of those x values, keep a margin of safety as well

velocity -> how much load expected in the next moment

f(velocity) = velocity * safetyFactor ; safetyFactor > 1 , convert shits to float to not get fucked 

how to store those moving averages without storing those darn last x values

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
    // mutex sync.Mutex // for adding new keys
}

type ShardManager struct {
    Shards []*Shard // pointers to shards
    mutex  sync.RWMutex 
}

// Hyperparameter
var ShardSize int32 = 5
var movingAverageXsize int32 = 4
var movingAverageArrayIdx int32 = 0 // update this shit atomically
var bufferSafetyFactor int32 = 1

var shardManagerBuffer int32 = 10

var movingAverageArray = make([]int32, movingAverageXsize)

// Global variables
var keyManager = KeyManager{
    Keys: sync.Map{},
}

var shardManager = ShardManager{
    Shards: make([]*Shard, 1),
}

var ShardManagerSizeLim int32 = 1
var curShardManagerSize int32 = 1

var nextIdx int32 = -1
var wg sync.WaitGroup
var dynamicBufferReshaperWG sync.WaitGroup

var curSetCnt int32 = 0
var lastSetCnt int32 = 0

func dynamicBufferReshaperWorker() {
    fmt.Println("Buffer reshape in progress===========================================================")
    // get the average of all the shits in movingAverageArray and then update the fucking buffer
    
    var sum int32 = 0

    for _,i := range(movingAverageArray) {
        sum += i
    }

    movingAvg := sum / movingAverageXsize

    newBuffer := movingAvg * bufferSafetyFactor

    if newBuffer == 0 {
        // can't be this low man, lmao
        newBuffer = 1
    }

    atomic.SwapInt32(&shardManagerBuffer,newBuffer)

    dynamicBufferReshaperWG.Done() 
}

/*
buffer is being updated in real time every second
*/
func dynamicBufferReshaper() {
    for {
        // time.Sleep(1 * time.Second)
        time.Sleep(50 * time.Millisecond)

        curVelocity := curSetCnt - lastSetCnt

        fmt.Println("current velocity is",curSetCnt - lastSetCnt)

        if curVelocity != 0 {
            fmt.Println(curVelocity)
        }
        
        movingAverageArray[movingAverageArrayIdx] = curVelocity

        newMovingAverageArrayIdx := (movingAverageArrayIdx + 1) % movingAverageXsize

        atomic.SwapInt32(&movingAverageArrayIdx,newMovingAverageArrayIdx)
        atomic.SwapInt32(&lastSetCnt,curSetCnt)

        dynamicBufferReshaperWG.Add(1)
        
        go dynamicBufferReshaperWorker()
        
        dynamicBufferReshaperWG.Wait()
    }
}

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

func resizeShardManagerWorker(addSize int32,curSize int32,curShardManagerSizeLim int32){
    shardManager.mutex.Lock()

    // check again if we still need to resize bro
    if int32(len(shardManager.Shards)) >= curShardManagerSizeLim {
        fmt.Println("0vo, my bad, sowwie")
        shardManager.mutex.Unlock()
        return
    }

    lenn := len(shardManager.Shards)

    for i := lenn; i < lenn + int(addSize) ; i++ {
        shardManager.Shards = append(shardManager.Shards,getNewShard(ShardSize)) // TODO: using append here might not be the best thing to do!
    }

    fmt.Println("resizing done")
    atomic.SwapInt32(&curShardManagerSize, curSize)

    fmt.Println("curshardmanagersize updated")

    atomic.SwapInt32(&ShardManagerSizeLim, curShardManagerSizeLim)

    fmt.Println("end of resizing function")
    shardManager.mutex.Unlock()
    fmt.Println("resize worker lock released")  
    wg.Done() 
}

func resizeShardManager(){
    for {        
        var curShardManagerSizeLim int32 = atomic.LoadInt32(&ShardManagerSizeLim)
        
        var curSize int32 = atomic.LoadInt32(&curShardManagerSize)

        buffer := shardManagerBuffer

        if curSize >= curShardManagerSizeLim - buffer {
            var addSize int32 = curShardManagerSizeLim
            curShardManagerSizeLim *= 2
            wg.Add(1)
            go resizeShardManagerWorker(addSize,curSize,curShardManagerSizeLim)
            wg.Wait()
        } else {
            atomic.SwapInt32(&curShardManagerSize,curSize)
            atomic.SwapInt32(&ShardManagerSizeLim,curShardManagerSizeLim)
        }
    }
}

func _setKey(key string, value string) {
    // this thing accesses the whole darn shard manager and puts a lock on it, which is not good, very very bad
    // we are accessing the length of the number of shards and that is not what we want, we need another way to access shit man, this is bad, very very bad
    
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

    shardManager.mutex.Lock()

    fmt.Println("set worker locked acquired")

    fmt.Println(shardManager.Shards)

    if shardNumber >= int32(len(shardManager.Shards)) {
        fmt.Println("help me dadddy, I feel bad about this")
        // this shit will have terrible performance lmao

        var addSize int32 = ShardManagerSizeLim
        lenn := len(shardManager.Shards)


        for i := lenn; i < lenn + int(addSize) ; i++ {
            fmt.Println("burrrrr")
            shardManager.Shards = append(shardManager.Shards,getNewShard(ShardSize)) // TODO: using append here might not be the best thing to do!
        }


        atomic.SwapInt32(&curShardManagerSize, int32(len(shardManager.Shards)))

        var newShardManagerSizeLim int32 = atomic.LoadInt32(&ShardManagerSizeLim)
        newShardManagerSizeLim *= 2

        atomic.SwapInt32(&ShardManagerSizeLim,newShardManagerSizeLim)

    }

    shard := shardManager.Shards[shardNumber]
    newVal := getNewValueData(value)

    if int32(len(shard.data)) > localShardIndex{
        fmt.Println("hi mom",shard.data)
        shard.data[localShardIndex] = newVal
    } else {
        fmt.Println("oh nu, hewp me daddy")
    }

    shardManager.mutex.Unlock()

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

    go dynamicBufferReshaper()
    go resizeShardManager()

    shardManager.Shards[0] = getNewShard(ShardSize)

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