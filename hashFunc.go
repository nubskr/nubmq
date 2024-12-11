package main

var tableSize int = 10

// just use a polynomial rolling hash for now
func getKeyHash(key string) int {
	hashValue := 0
	prime := 31 // A small prime number for mixing
	for _, char := range key {
		// TODO: totalCapacity is not a prime, which might not give a well 'distributed' distribution
		hashValue = (hashValue*prime + int(char)) % int(ShardManagerKeeper.totalCapacity)
	}
	return hashValue
}
