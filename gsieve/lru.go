package gsieve

import (
	"main/queue"
)

/*
**********************************************
||				 syncLRU                    ||
**********************************************
*/
type syncLRU struct {
	cache    map[any]*queue.Node
	queue    *queue.Queue
	capacity int
}

func NewsyncLRUCache(cap int) *syncLRU {
	return &syncLRU{capacity: cap}
}

func (lru *syncLRU) get(key string) (node *queue.Node, isExist bool) {
	lru.lazyInit()
	if node, ok := lru.cache[key]; ok {
		// lru.queue.MoveToHead(node)
		return node, ok
	}
	return nil, false
}

func (lru *syncLRU) put(node *queue.Node) {
	lru.lazyInit()
	key := node.Value.(*gsieveValue).k
	if node, ok := lru.cache[key]; ok {
		lru.queue.MoveToHead(node)
		return
	}
	if len(lru.cache) == lru.capacity {
		delete(lru.cache, (lru.queue.Tail().Prev()).Value.(*gsieveValue).k)
		lru.queue.RemoveNode(lru.queue.Tail().Prev())
	}
	lru.queue.AddToHead(node)
	lru.cache[key] = node
}

func (lru *syncLRU) lazyInit() {
	if lru.queue == nil {
		lru.queue = queue.NewQueue()
		lru.cache = make(map[any]*queue.Node, lru.capacity)
	}
}

func (lru *syncLRU) del(key any) {
	if node, ok := lru.cache[key]; ok {
		delete(lru.cache, key)
		lru.queue.RemoveNode(node)
	}
}
/*
**********************************************
||				      LRU                   ||
**********************************************
*/

type LRU struct {
	cache    map[any]*queue.Node
	queue    *queue.Queue
	capacity int
}

func NewLRUCache(cap int) *LRU {
	return &LRU{cache: make(map[any]*queue.Node, cap), queue: queue.NewQueue(), capacity: cap}
}

func (lru *LRU) Get(key any) (*queue.Node, bool) {
	if node, ok := lru.cache[key]; ok {
		lru.queue.MoveToHead(node)
		return node, true
	}
	return nil, false
}

func (lru *LRU) Put(node *queue.Node) {
	key := node.Value.(*sieveValue).k
	if n, ok := lru.cache[key]; ok {
		n.Value.(*sieveValue).v = node.Value.(*sieveValue).v
		lru.queue.MoveToHead(n)
		return
	}
	if len(lru.cache) == lru.capacity {
		lru.evict()
	}
	lru.queue.AddToHead(node)
	lru.cache[key] = node
}

func (lru *LRU) evict() {
	delete(lru.cache, (lru.queue.Tail().Prev()).Value.(*sieveValue).k)
	lru.queue.RemoveNode(lru.queue.Tail().Prev())
}

func (lru *LRU) del(key any) {
	if node, ok := lru.cache[key]; ok {
		delete(lru.cache, key)
		lru.queue.RemoveNode(node)
	}
}