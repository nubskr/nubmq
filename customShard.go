package main

import (
	"sync"
	"time"
)

/*
- customHome.Store(key, entry)->set
- customHome.Load(key) ->get
*/
var CUSTOM_MAP_BUCKET_SIZE = 3

type CustomMapEntry struct {
	Data  []Entry
	mutex sync.RWMutex
}

type CustomMap struct {
	Buckets []CustomMapEntry
}

func GetNewCustomMapEntry() *CustomMapEntry {
	return &CustomMapEntry{
		Data: make([]Entry, 0),
	}
}

func GetNewCustomMap() *CustomMap {
	data := CustomMap{
		Buckets: make([]CustomMapEntry, CUSTOM_MAP_BUCKET_SIZE),
	}
	for i := 0; i < CUSTOM_MAP_BUCKET_SIZE; i++ {
		data.Buckets[i] = *GetNewCustomMapEntry()
	}

	return &data
}

func (m *CustomMap) getHash(key string) int {
	// get the first character, last character, and the middle character, get a XOR hash out of it % CUSTOM_MAP_BUCKET_SIZE
	length := len(key)
	if length == 0 {
		return 0
	}

	hash := int(key[0]) ^ int(key[length/2]) ^ int(key[length-1])

	return hash % CUSTOM_MAP_BUCKET_SIZE
}

func (m *CustomMap) Load(key string) (Entry, bool) {
	// value -> Entry object btw
	entry := Entry{}
	exists := false
	// get the bucket by hashing the key, then traverse the bucket
	bucket := m.getHash(key)
	m.Buckets[bucket].mutex.RLock()
	defer m.Buckets[bucket].mutex.RUnlock()

	for _, e := range m.Buckets[bucket].Data {
		if e.key == key {
			entry = e
			exists = true
			break
		}
	}

	return entry, exists
}

func (m *CustomMap) Store(key string, value Entry) bool {
	alreadyExists := false

	bucket := m.getHash(key)
	m.Buckets[bucket].mutex.Lock()
	defer m.Buckets[bucket].mutex.Unlock()

	for i, e := range m.Buckets[bucket].Data {
		if e.key == key {
			m.Buckets[bucket].Data[i] = value
			alreadyExists = true
			break
		}
	}

	if !alreadyExists {
		m.Buckets[bucket].Data = append(m.Buckets[bucket].Data, value)
	}

	return alreadyExists
}

func (m *CustomMap) GetAll() []Entry {
	result := make([]Entry, 0)
	now := time.Now().Unix()

	for i := 0; i < CUSTOM_MAP_BUCKET_SIZE; i++ {
		m.Buckets[i].mutex.RLock()
		for _, entryVal := range m.Buckets[i].Data {
			if entryVal.canExpire && entryVal.TTL < now {

			} else {
				result = append(result, entryVal)
			}
		}
		m.Buckets[i].mutex.RUnlock()
	}

	return result
}
