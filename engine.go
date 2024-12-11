package main

// import (
//     "fmt"
//     "log"
//     "net"
//     "sync/atomic"
//     "os"
// 	"unsafe"
//     "sync"
// 	// "time"
//     "strings"
//     "reflect"
//     "runtime"
// )

// /*
// resizing_going_on -> bool

// nextShardManagerSize -> int

// nextShardManager.Shards -> *[]Shards

// _set request comes:
//     if current sm size is not big enough:
//         if nextshardmanagersize is not big enough
//             is the resizing going on ? if yes, then wait for it to be over
//                     is it

// for atomic.LoadInt32(nextShardManagerSize) <= curShard
//     for atomic.LoadInt32(resizing_going_on) == 1
//         // wait for resizing to be done

//     // is the size big enough now ? should we wait for something to happen ?

// */

// type Message struct {
// 	data string
// 	timestamp int64
// }

// type ValueData struct {
// 	data string
//     mutex sync.RWMutex
// }

// type Shard struct {
//     data []*ValueData
//     size int32
// }

// type KeyManager struct {
//     Keys  sync.Map
//     // mutex sync.Mutex // for adding new keys
// }

// type ShardManager struct {
//     Shards []*Shard // pointers to shards
//     mutex  sync.RWMutex
// }

// var ShardSize int32 = 50
// var movingAverageXsize int32 = 4
// var movingAverageArrayIdx int32 = 0 // update this shit atomically
// var resizing_going_on int32 = 0 // true or false , using int just so we can manipulate it atomically
// var bufferSafetyFactor int32 = 10

// var shardManagerBuffer int32 = 10

// var movingAverageArray = make([]int32, movingAverageXsize)

// var keyManager = KeyManager{
//     Keys: sync.Map{},
// }

// // init shards
// var shardManager = ShardManager{
//     Shards: make([]*Shard, 1),
// }

// var nextShardManager = ShardManager{
//     Shards: make([]*Shard, 1),
// }

// var ShardManagerSizeLim int32 = 1
// var curShardManagerSize int32 = 1

// var nextIdx int32 = -1
// var wg sync.WaitGroup
// var dynamicBufferReshaperWG sync.WaitGroup

// var curSetCnt int32 = 0
// var lastSetCnt int32 = 0

// // func dynamicBufferReshaperWorker() {
// //     // get the average of all the shits in movingAverageArray and then update the darn  buffer

// //     var sum int32 = 0

// //     for _,i := range(movingAverageArray) {
// //         sum += i
// //     }

// //     movingAvg := sum / movingAverageXsize

// //     newBuffer := movingAvg * bufferSafetyFactor

// //     if newBuffer == 0 {
// //         newBuffer = 1
// //     }

// //     atomic.SwapInt32(&shardManagerBuffer,newBuffer)

// //     dynamicBufferReshaperWG.Done()
// // }

// /*
// buffer is being updated in real time every second
// */
// // func dynamicBufferReshaper() {
// //     for {
// //         time.Sleep(100 * time.Millisecond)

// //         curVelocity := curSetCnt - lastSetCnt

// //         if curVelocity != 0 {
// //             fmt.Println("current velocity is",curSetCnt - lastSetCnt)
// //         }

// //         movingAverageArray[movingAverageArrayIdx] = curVelocity

// //         newMovingAverageArrayIdx := (movingAverageArrayIdx + 1) % movingAverageXsize

// //         atomic.SwapInt32(&movingAverageArrayIdx,newMovingAverageArrayIdx)
// //         atomic.SwapInt32(&lastSetCnt,curSetCnt)

// //         dynamicBufferReshaperWG.Add(1)

// //         go dynamicBufferReshaperWorker()

// //         dynamicBufferReshaperWG.Wait()
// //     }
// // }

// func getNewShard(sz int32) *Shard {
//     return &Shard{
//         data: make([]*ValueData, sz, ShardSize),
//     }
// }

// func getNewValueData(value string) *ValueData{
//     return &ValueData{
//         data: value,
//     }
// }

// // func resizeShardManagerWorker(addSize int32,curSize int32,curShardManagerSizeLim int32){
// //     fmt.Println("<-------------resize worker started------------->")
// //     shardManager.mutex.Lock()
// // 	fmt.Println("resize manager acquired shardmanager lock")

// //     // check again if we still need to resize
// //     if int32(len(shardManager.Shards)) >= curShardManagerSizeLim {
// //         fmt.Println("0vo, my bad, sowwie")
// //         shardManager.mutex.Unlock()

// //         os.Exit(1)
// //         return
// //     }

// //     lenn := len(shardManager.Shards)

// //     // we make a deep copy
// //     tempShardManager := make([]*Shard, lenn+int(addSize))
// //     copy(tempShardManager, shardManager.Shards)

// //     shardManager.mutex.Unlock()

// //     atomic.SwapInt32(&resizing_going_on, 1)

// //     for i := lenn; i < lenn + int(addSize) ; i++ {
// //         tempShardManager[i] = getNewShard(ShardSize)
// //     }

// //     nextShardManager.mutex.Lock()

// //     nextShardManager.Shards = tempShardManager

// //     nextShardManager.mutex.Unlock()

// //     atomic.SwapInt32(&resizing_going_on, 0)

// // 	fmt.Println("resize manager released shardmanager lock")

// //     atomic.SwapInt32(&curShardManagerSize, curSize)

// //     fmt.Println("curshardmanagersize updated")

// //     atomic.SwapInt32(&ShardManagerSizeLim, curShardManagerSizeLim)

// //     wg.Done()
// // }

// // func resizeShardManagerWorker(addSize int32,curSize int32,curShardManagerSizeLim int32){
// //     fmt.Println("<-------------resize worker started------------->")
// //     shardManager.mutex.Lock()
// // 	fmt.Println("resize manager acquired shardmanager lock")

// //     // check again if we still need to resize
// //     if int32(len(shardManager.Shards)) >= curShardManagerSizeLim {
// //         fmt.Println("0vo, my bad, sowwie")
// //         shardManager.mutex.Unlock()

// //         os.Exit(1)
// //         return
// //     }

// //     lenn := len(shardManager.Shards)

// //     // we make a deep copy
// //     tempShardManager := make([]*Shard, lenn+int(addSize))
// //     copy(tempShardManager, shardManager.Shards)

// //     shardManager.mutex.Unlock()

// //     atomic.SwapInt32(&resizing_going_on, 1)

// //     for i := lenn; i < lenn + int(addSize) ; i++ {
// //         tempShardManager[i] = getNewShard(ShardSize)
// //     }

// //     nextShardManager.mutex.Lock()

// //     nextShardManager.Shards = tempShardManager

// //     nextShardManager.mutex.Unlock()

// //     atomic.SwapInt32(&resizing_going_on, 0)

// // 	fmt.Println("resize manager released shardmanager lock")

// //     atomic.SwapInt32(&curShardManagerSize, curSize)

// //     fmt.Println("curshardmanagersize updated")

// //     atomic.SwapInt32(&ShardManagerSizeLim, curShardManagerSizeLim)

// //     wg.Done()
// // }

// // func resizeShardManager(){
// //     for {
// //         var curShardManagerSizeLim int32 = atomic.LoadInt32(&ShardManagerSizeLim)

// //         var curSize int32 = atomic.LoadInt32(&curShardManagerSize)

// //         buffer := shardManagerBuffer

// //         if curSize >= curShardManagerSizeLim - buffer {
// //             var addSize int32 = curShardManagerSizeLim
// //             curShardManagerSizeLim *= 2
// //             wg.Add(1)
// //             go resizeShardManagerWorker(addSize,curSize,curShardManagerSizeLim)
// //             wg.Wait()
// //         } else {
// //             atomic.SwapInt32(&curShardManagerSize,curSize)
// //             atomic.SwapInt32(&ShardManagerSizeLim,curShardManagerSizeLim)
// //         }
// //     }
// // }

// func waitTillBigEnough(shardNumber) {
//     // waits untill we are big enough to keep this shit

//     for shardNumber >= int32(len(shardManager.Shards)) {
//         for atomic.LoadInt32(&) == 1 {
//             fmt.Println("waiting on resizing")

//         }

//         // fuck it, remove the crap, make it very simple, probably a darn function
//     }
// }

// func doubleNextShardManager() {
//     // we only want atmax one of this thing running at a time
//     fmt.Println("we tryna do somethin here")
//     // os.Exit(1)
//     if atomic.LoadInt32(&resizing_going_on) == 1 {
//         fmt.Println("More than one doubling called bruh")
//         os.Exit(1)
//     }

//     // shardManager.mutex.Lock()

//     lenn := len(shardManager.Shards)

//     // we make a deep copy
//     tempShardManager := make([]*Shard, lenn*2)
//     copy(tempShardManager, shardManager.Shards)

//     // shardManager.mutex.Unlock()

//     atomic.SwapInt32(&resizing_going_on, 1)

//     for i := lenn; i < 2 * lenn; i++ {
//         tempShardManager[i] = getNewShard(ShardSize)
//     }

//     fmt.Println(1)
//     nextShardManager.mutex.Lock()
//     fmt.Println(2)
//     nextShardManager.Shards = tempShardManager

//     nextShardManager.mutex.Unlock()
//     fmt.Println(3)
//     fmt.Println("doubling finished")
//     atomic.SwapInt32(&resizing_going_on, 0)
// }

// func _setKey(key string, value string) {
//     idx := int32(696969696)

//     if value, ok := keyManager.Keys.Load(key); ok {
//         if intValue, ok := value.(int32); ok {
//             idx = int32(intValue)
//         } else {
//             fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO set-x-x-x-x-x-x-x-x-x-x-xx-x-x-x-x-x-x--x",value,"-->")
//             os.Exit(1)
//         }
//     } else {
//         val := atomic.AddInt32(&nextIdx, 1)
//         keyManager.Keys.Store(key, val)
//         idx = val
//     }

//     if idx == 696969696 {
//         fmt.Println("trying to set non existing shit")
//         os.Exit(1)
//     }

//     shardNumber := idx / ShardSize
//     localShardIndex := idx%ShardSize

//     fmt.Println("setting key",key,"at",idx,"at shard number",shardNumber,"at local index",localShardIndex)

//     fmt.Println("trying to acquire lock to set key")

// 	newVal := getNewValueData(value)

//     shardManager.mutex.Lock()

// 	for shardNumber >= int32(len(shardManager.Shards)) {
//         for atomic.LoadInt32(&) == 1 {
//             fmt.Println("waiting on resizing")

//         }

//         // fuck it, remove the crap, make it very simple, probably a darn function
//     }

//     if shardNumber >= int32(len(shardManager.Shards)) {
//         fmt.Println("We're fucked sire, *salutes*")
//         fmt.Println(shardNumber,int32(len(shardManager.Shards)))
//         os.Exit(1)
//     }

//     shard := shardManager.Shards[shardNumber]
// 	fmt.Println("set worker lock released")

// 	shardManager.mutex.Unlock()

// 	// value is a darn string
// 	atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&shard.data[localShardIndex])), unsafe.Pointer(newVal))
// }

// func _getKey(key string) (string, bool) {

//     idx := int32(696969696) // TODO: remove this shit

//     if value, ok := keyManager.Keys.Load(key); ok {
//         if intValue, ok := value.(int32); ok {
//             idx = int32(intValue)
//         } else {

//             fmt.Println("NOOOOOOOOOOOOOOOOOOOOOO get-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x-x",value,"-->")
//             os.Exit(1)
//         }
//     } else {
//         return "NaN", false
//     }

//     if idx == 696969696 {
//         fmt.Println("trying to get non existing shit")
//         os.Exit(1)
//     }

//     shardNumber := idx / ShardSize

//     shardManager.mutex.Lock()

//     if shardNumber >= int32(len(shardManager.Shards)) {
//         shardManager.mutex.Unlock()
//         return "", false
//     }

//     shard := shardManager.Shards[shardNumber]
//     localShardIndex := idx % ShardSize

//     if localShardIndex < int32(len(shard.data)) {
//         shardManager.mutex.Unlock()
//         return (shard.data[localShardIndex]).data, true
//     }

//     shardManager.mutex.Unlock()
//     return "NaN", false
// }

// func handleConnection(conn net.Conn) {
//     fmt.Println("Client connected")
//     buffer := make([]byte, 1024)
//     for {
//         length, err := conn.Read(buffer)

//         if err != nil {
//             fmt.Println("an error occured while reading message:",err)
//             return
//         }

//         data := string(buffer[:length])

//         stringData := strings.Fields(data)

//         if stringData[0] == "SET"{
//             atomic.AddInt32(&curSetCnt, 1)
//             _setKey(stringData[1],stringData[2])
//             _, err := conn.Write([]byte(fmt.Sprint("SET done\n")))

//             if err != nil {
//                 log.Println("failed to reply message:" ,err)
//             } else{
//             }
//         } else{
//             output ,exists := _getKey(stringData[1])

//             if exists {

//             }
//             _, err := conn.Write([]byte(fmt.Sprint(output+"\n"))) // Send message over connection

//             if err != nil {
//             } else{
//             }
//         }
//     }
// }

// func main() {
//     fasttttt := true

//     // fasttttt = false

//     if fasttttt {
//         runtime.GOMAXPROCS(runtime.NumCPU())
//     }

//     ln, err := net.Listen("tcp", ":8080")

//     if err != nil {
//         log.Fatal(err)
//     }

// 	// go dynamicBufferReshaper()
//     // go resizeShardManager()

// 	// TODO: remove this shit
//     for i := 0 ; i < 1 ; i++ {
//         shardManager.Shards[i] = getNewShard(ShardSize)
//     }

// 	fmt.Println("Server listening on :8080")

//     for {
//         // Accept connection
//         conn, err := ln.Accept()
//         if err != nil {
//             log.Println(err)
//             continue
//         }

//         go handleConnection(conn)
//     }
// }
