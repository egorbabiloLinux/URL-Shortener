package balancer

import (
	"hash/fnv"
)

func Balance(key []byte, partitions ...int32) (partition int32) {
	hash := func(k []byte) int {
		h := fnv.New32a() 
		h.Write(k)
		return int(h.Sum32())
	}

	return partitions[hash(key)%len(partitions)]
}