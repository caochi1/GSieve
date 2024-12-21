package gsieve

import "main/queue"

type gsieve struct {
	ghostList *LRU
	sieve     *Sieve
	miss      int
}

func NewGSieve(cap int) *gsieve {
	return &gsieve{
		sieve:     NewSieveCache(cap),
		ghostList: NewLRUCache(cap / 2),
	}
}

func (gsieve *gsieve) Miss() int {
	return gsieve.miss
}

func (gsieve *gsieve) Get(key any) (any, bool) {
	if value, isExist := gsieve.sieve.Get(key); isExist {
		return value, true
	} else if node, ok := gsieve.ghostList.Get(key); ok {
		gsieve.ghostList.del(key)
		s := gsieve.sieve
		if len(s.cache) == s.capacity {
			gsieve.evict()
		}
		s.queue.AddToHead(node)
		s.cache[key] = node
		return node.Value.(*sieveValue).v, true
	}
	gsieve.miss++
	return nil, false
}

func (gsieve *gsieve) Insert(key, value any) {
	s := gsieve.sieve
	if node, exist := s.cache[key]; exist {
		node.Value.(*sieveValue).v = value
		node.Value.(*sieveValue).visited = true
		return
	}
	node := &queue.Node{Value: &sieveValue{key, value, false}}
	if len(s.cache) == s.capacity {
		gsieve.evict()
	}
	if _, ok := gsieve.ghostList.Get(key); ok {
		gsieve.ghostList.del(key)
	}
	s.queue.AddToHead(node)
	s.cache[key] = node
}

func (gsieve *gsieve) evict() {
	s := gsieve.sieve
	ghost := s.evict()
	gsieve.ghostList.Put(ghost)
}
