package main

import (
	"crypto/sha1"
	"sort"
	"sync"
)

type Node struct {
	ID   string //unique identifier
	Addr string // network address
}

type HashRing struct {
	nodes  []Node   //list of nodes
	hashes []uint32 // hashes for nodes
	lock   sync.RWMutex
}

func NewHashRing() *HashRing {
	return &HashRing{}
}

func (h *HashRing) AddNode(node Node) {
	h.lock.Lock()
	defer h.lock.Unlock()
	hash := h.hash(node.ID)
	h.nodes = append(h.nodes, node)
	h.hashes = append(h.hashes, hash)
	sort.Slice(h.hashes, func(i, j int) bool {
		return h.hashes[i] < h.hashes[j]
	})

}

func (h *HashRing) RemoveNode(nodeID string) {
	// TO BE IMPLEMENTED
}

func (h *HashRing) GetNode(key string) Node {
	// performs binary search on the sorted hashes
	if len(h.nodes) == 0 {
		return Node{}
	}

	h.lock.RLock()
	defer h.lock.RUnlock()
	hash := h.hash(key)
	index := sort.Search(len(h.hashes), func(i int) bool {
		return h.hashes[i] >= hash
	})
	if index == len(h.hashes) {
		index = 0
	}
	return h.nodes[index]
}

func (h *HashRing) hash(key string) uint32 {
	// uses SHA-1
	hsh := sha1.New()
	hsh.Write([]byte(key))
	return h.bytesToUint32(hsh.Sum(nil))
}

func (h *HashRing) bytesToUint32(b []byte) uint32 {
	return (uint32(b[0]) << 24) | ((uint32(b[1]) << 16)) | (uint32(b[2]) << 8) | (uint32(b[3]))
}
