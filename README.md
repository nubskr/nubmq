Goal: Be a beast outperforming everything that comes in its way

# TODO:

## Core engine changes:
- Optimize for speed

## New feats:
- key expiry event notifs(tricky)

Problem:

- we are relying on usedCapacity of SM to trigger upgrades and downgrades, but it is not reliable at all

solution: make the totalCapacity and usedCapacity a *int from int

- we need a non blocking EventQueue so core engine connection reads don't get blocked, such a massively buffered channel is being created for a global queue, takes a SHIT LOAD of memory on initialization, also in some case like a connection disconnecting, their secondary write channel just keeps getting bigger and bigger without getting drained, could be handled separately when they disconnect, but still stays a big issue for the core engine

### commands supported rn:
```
SET <key> <value>
SET <key> <value> EX <expiry_time_in_seconds>

GET <key>

SUBSCRIBE <key_name>
```




log this from server's side:


for each request irrespective of set or get:
- time in epochs of when it was DONE processing from the engine and the amount of time it took to.... wait.. not again, okay, just the Timestamp_ms, nothing else

var appendQueue chan int64 = make(chan int64, 50000000) // just don't block ffs

func appendToLog(logPath string,epochTimestamp int){
    // init log on logPath by emptying or creating it and appending `Timestamp_ms` to it

    // each append under 4kb ,so should directly hit the storage instead of page cache
    for{
        appendData <- appendQueue
        // append to the log on logPath

    }
}