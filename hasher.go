package main

import "os"

// just use a polynomial rolling hash for now
func getKeyHash(key string, keeper *ShardManagerKeeperTemp) int {
	if int(keeper.totalCapacity) == 0 {
		os.Exit(1)
	}
	hashValue := 0
	prime := 31 // A small prime number for mixing
	for _, char := range key {
		// TODO: totalCapacity is not a prime, which might not give a well 'distributed' distribution
		hashValue = (hashValue*prime + int(char)) % int(keeper.totalCapacity)
	}
	return hashValue
}
