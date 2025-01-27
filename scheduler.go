package main

import (
	"log"
	"sync"
	"time"

	"github.com/nubmq/set"
)

/*
set.Size(entry)
set.Insert(entry)
set.Erase(entry)
set.Contains(entry)
set.Clear()
set.Begin(), to get the actual entry of out it: entry := (set.Begin()).Value().(Entry)
set.RBegin()
*/

type SetStorage struct {
	queue                    chan Entry
	set                      set.Set
	mutex                    sync.Mutex
	KeyEntryKeeper           map[string]Entry
	EarliestExpiringKeyEntry Entry
}

var UpdateChan chan time.Duration = make(chan time.Duration)

func GetSet() *set.Set {
	return set.NewSet(func(a, b interface{}) int {
		entryA := a.(Entry)
		entryB := b.(Entry)
		if entryA.TTL < entryB.TTL {
			return -1
		} else if entryA.TTL > entryB.TTL {
			return 1
		}
		return 0
	})
}

var SetContainer SetStorage = SetStorage{
	queue:          make(chan Entry, 1000), // this shit can get clogged and might
	set:            *GetSet(),
	KeyEntryKeeper: make(map[string]Entry),
}

// inserts stuff in setStorage
func HandleKeyTTLInsertion(setStorage *SetStorage, updateChan *chan time.Duration) {
	for {
		entry := <-setStorage.queue

		setStorage.mutex.Lock()

		val, exists := setStorage.KeyEntryKeeper[entry.key]
		if exists {
			if setStorage.EarliestExpiringKeyEntry == val {
				if setStorage.set.Size() > 0 {
					setStorage.EarliestExpiringKeyEntry = (setStorage.set.Begin()).Value().(Entry)

				}
			}
			setStorage.set.Remove(val)
		}
		if setStorage.set.Size() == 0 || setStorage.EarliestExpiringKeyEntry.TTL > entry.TTL {
			setStorage.EarliestExpiringKeyEntry = entry
			*updateChan <- time.Duration(entry.TTL-time.Now().Unix()) * time.Second
		}
		setStorage.KeyEntryKeeper[entry.key] = entry
		setStorage.set.Insert(entry)

		setStorage.mutex.Unlock()
	}
}

func HandleKeyTTLEviction(setStorage *SetStorage, updateChan *chan time.Duration, expirationEventChannel *chan Entry) {
	defaultPollingTime := time.Duration(1 * time.Hour)
	timer := time.NewTimer(defaultPollingTime)

	for {
		select {
		case <-timer.C:
			setStorage.mutex.Lock()
			if setStorage.set.Size() > 0 {
				log.Print("in")
				for it := setStorage.set.Begin(); setStorage.set.Size() > 0; {
					log.Print("pura in")
					entry := it.Value().(Entry)
					log.Print("got some entry: ", entry.key)
					curTTL := entry.TTL
					now := time.Now().Unix()
					if curTTL <= now {
						setStorage.set.Remove(entry)
						log.Print("Key expired: ", entry.key)
						entry.isExpiryEvent = true
						*expirationEventChannel <- entry
					} else {
						log.Print("bad boi")
						duration := time.Duration(curTTL-now) * time.Second
						timer.Reset(duration)
						break
					}
				}
			} else {
				timer.Reset(defaultPollingTime)
			}
			setStorage.mutex.Unlock()
		case newDuration := <-*updateChan:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(newDuration)
		}
	}
}

// func InitScheduler() {
// 	go HandleKeyTTLInsertion(&SetContainer, &updateChan)
//  go HandleKeyTTLEviction(&SetContainer, &updateChan)

// 	// NOTE: only insert in queue if canExpire is true in entry
// }
