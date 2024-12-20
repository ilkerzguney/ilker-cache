package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

//A simple cache but to prevent concurrent access we need to use thread safety approaches, some options to consider.
// sync package -> sync.Mutex, reduced performance under heavy load , leads to contention
// sync package -> sync.RWMutex better than above, multiple reads same time access cache but only one writer is allowed to write
// sync.map built-in concurrent map implementation for frequent reads and infrequent writes, and many other third party libraries.

//lock-free data structures for some specific use cases like atomic operations without using explicit locking, more complex to design
//channel-based synchronization
// * serialized access: a dedicated goroutine that handles all cache operations, other gors communicate with this goroutine through channels.
// * sharded cache: divide the cache into multiple shards, each protected by its own mutex or concurrent map
// choosing the right approach depends on specific reqs.
// read/write ratio: reads are significantly  more frequent than writes, sync.RWMutex or sync.Map
// performance: if max perf is critical, consider lock-free ds or sharded cache
// simplicity: if ease of implementation is a priority, sync.Mutex or a channel-based approach might be simpler
// for now, let make things simpler. sync.RWMutex

type CacheItem struct {
	Value string
	TTL   time.Time
}

type Cache struct {
	mu       sync.RWMutex
	items    map[string]*list.Element
	eviction *list.List
	capacity int
}

type entry struct {
	key   string
	value CacheItem
}

func NewCache(capacity int) *Cache {
	return &Cache{
		items:    make(map[string]*list.Element),
		eviction: list.New(),
		capacity: capacity,
	}
}

func (c *Cache) Set(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, found := c.items[key]; found {
		c.eviction.Remove(element)
		delete(c.items, key)
	}

	if c.eviction.Len() >= c.capacity {
		c.evictLRU()
	}

	newItem := CacheItem{
		Value: value,
		TTL:   time.Now().Add(ttl),
	}
	elem := c.eviction.PushFront(&entry{key: key, value: newItem})
	c.items[key] = elem
	fmt.Printf("set for key:%s completed\n", key)
}

func (c *Cache) evictLRU() {
	elem := c.eviction.Back()
	if elem != nil {
		c.eviction.Remove(elem)
		delete(c.items, elem.Value.(*entry).key)
	}
	panic("unimplemented")
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[key]
	if !found {
		return "", false
	}

	if time.Now().After(item.Value.(*entry).value.TTL) {
		c.eviction.Remove(item)
		delete(c.items, key)
		return "", false
	}

	c.eviction.MoveToFront(item)
	return item.Value.(*entry).value.Value, true
}

func (c *Cache) starEvictionTicker(d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			c.evictExpiredItems()
		}
	}()
}

func (c *Cache) evictExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Value.(*entry).value.TTL) {
			delete(c.items, key)
		}
	}
}
