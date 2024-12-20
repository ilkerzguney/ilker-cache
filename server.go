package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// it will be http server, trade-off design

const replicationHeader = "X-Replication-Request"

type CacheServer struct {
	cache    *Cache
	peers    []string
	hashRing *HashRing
	selfID   string
	mu       sync.Mutex
}

func NewCacheServer(peers []string, selfID string) *CacheServer {
	cs := &CacheServer{
		cache:    NewCache(5),
		peers:    peers,
		selfID:   selfID,
		hashRing: NewHashRing(),
	}
	for _, peer := range peers {
		cs.hashRing.AddNode(Node{ID: peer, Addr: peer})
	}
	cs.hashRing.AddNode(Node{ID: selfID, Addr: "self"})
	return cs
}

func (cs *CacheServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		TTL   int    `json:"ttl"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ttlDefault := req.TTL
	if ttlDefault == 0 {
		ttlDefault = 1 * int(time.Hour)
	}

	targetNode := cs.hashRing.GetNode(req.Key)
	if targetNode.Addr == "self" {
		cs.cache.Set(req.Key, req.Value, time.Duration(ttlDefault))
		if r.Header.Get(replicationHeader) != "true" {
			go cs.replicateSet(req.Key, req.Value)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		cs.forwardRequest(w, targetNode, r)
	}

}

func (cs *CacheServer) forwardRequest(w http.ResponseWriter, targetNode Node, r *http.Request) {
	client := &http.Client{}
	var req *http.Request
	var err error

	if r.Method == http.MethodGet {
		url := fmt.Sprintf("%s%s?%s", targetNode.Addr, r.URL.Path, r.URL.RawQuery)
		req, err = http.NewRequest(r.Method, url, nil)
	} else if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		url := fmt.Sprintf("%s%s", targetNode.Addr, r.URL.Path)
		req, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	}

	if err != nil {
		log.Printf("Failed to create forward request %v", err)
		http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
		return
	}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to forward requests to node %s: %v", targetNode.Addr, err)
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	w.WriteHeader(response.StatusCode)
	_, err = io.Copy(w, response.Body)
	if err != nil {
		log.Printf("Failed to write response from node %s: %v", targetNode.Addr, err)
	}

}

func (cs *CacheServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Println("Received key:", key)
	targetNode := cs.hashRing.GetNode(key)
	if targetNode.Addr == "self" {
		val, found := cs.cache.Get(key)
		if !found {
			http.NotFound(w, r)
			return
		}
		err := json.NewEncoder(w).Encode(map[string]string{"value": val})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		originalSender := r.Header.Get("X-Forwarded-For")
		if originalSender == cs.selfID {
			http.Error(w, "Loop detected", http.StatusBadRequest)
			return
		}
		r.Header.Set("X-Forwarded-For", cs.selfID)
		cs.forwardRequest(w, targetNode, r)
	}
}

func (cs *CacheServer) replicateSet(key, value string) {
	req := struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{Key: key, Value: value}

	data, err := json.Marshal(req)
	if err != nil {
		fmt.Println("Data marshalling error:", err)
		fmt.Println()
		return
	}

	for _, peer := range cs.peers {
		if peer != "self" {
			go func(peer string) {
				client := http.Client{}
				req, err := http.NewRequest("POST", peer+"/set", bytes.NewReader(data))

				if err != nil {
					log.Printf("Failed to create replication request: %v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set(replicationHeader, "true")
				_, errResponse := client.Do(req)

				if errResponse != nil {
					log.Printf("Failed to replicate to peer: %s: %v", peer, errResponse)
					return
				}

				log.Printf("Replication is successful to peer: %s", peer)
			}(peer)
		}
	}
}
