package main

import (
	"testing"
)

func TestHashRing_AddNode(t *testing.T) {
	hashRing := NewHashRing()
	node := Node{ID: "123", Addr: ":5000"}
	hashRing.AddNode(node)

	if len(hashRing.nodes) != 1 {
		t.Errorf("len(hashRing.nodes) = %v, want %v", len(hashRing.nodes), 1)
	}
}

func TestHashRing_GetNode(t *testing.T) {
	hashRing := NewHashRing()
	node := Node{ID: "123", Addr: ":5000"}
	key := "123"
	hashRing.AddNode(node)

	if len(hashRing.nodes) != 1 {
		t.Errorf("len(hashRing.nodes) = %v, want %v", len(hashRing.nodes), 1)
	}

	targetNode := hashRing.GetNode(key)
	if targetNode.ID == "" {
		t.Errorf("targetNode is nil, want not nil")
	}
}
