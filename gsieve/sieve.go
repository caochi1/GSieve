package gsieve

import (
	"main/queue"
	"sync"
	"time"
)

/*
**********************************************
||				syncSieve                   ||
**********************************************
*/

type levelValue struct {
	level int
}

func newSieve(level int) *syncSieve {
	s := &syncSieve{
		queue: queue.NewQueue(),
		cache: make(map[string]*queue.Node),
		level: level,
	}
	s.queue.Head().Value = levelValue{level}
	return s
}

type syncSieve struct {
	queue *queue.Queue
	lock  sync.RWMutex
	cache map[string]*queue.Node
	level int
}

func (s *syncSieve) get(key string) []byte {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if node, exist := s.cache[key]; exist {
		node.Value.(*gsieveValue).visited = true
		return node.Value.(*gsieveValue).v
	}
	return nil
}

func (s *syncSieve) set(node *queue.Node) {
	key := node.Value.(*gsieveValue).k
	s.lock.Lock()
	defer s.lock.Unlock()
	if n, exist := s.cache[key]; exist {
		n.Value.(*gsieveValue).v = node.Value.(*gsieveValue).v
		n.Value.(*gsieveValue).visited = true
	} else {
		s.cache[key] = node
		s.queue.AddToHead(node)
	}
}

func (s *syncSieve) del(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.queue.RemoveNode(s.cache[key])
	delete(s.cache, key)
}

func (s *syncSieve) cleanup(gs *GSieve) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.cache) == 0 {
		return
	}
	node := s.tail().Prev()
	for node != s.head() {
		if !isExpired(node) {
			break
		}
		prev := node.Prev()
		gs.lock.Lock()
		gs.del(node.Value.(*gsieveValue).k)
		gs.lock.Unlock()
		node = prev
	}
}

func (s *syncSieve) head() *queue.Node {
	return s.queue.Head()
}

func (s *syncSieve) tail() *queue.Node {
	return s.queue.Tail()
}

func isExpired(node *queue.Node) bool {
	return time.Now().Unix() > node.Value.(*gsieveValue).Expiration
}

/*
**********************************************
||				    Sieve                   ||
**********************************************
*/

type Sieve struct {
	cache    map[any]*queue.Node
	queue    *queue.Queue
	hand     *queue.Node
	capacity int
	miss     int
}

type sieveValue struct {
	k, v    interface{}
	visited bool
}

func NewSieveCache(cap int) *Sieve {
	q := queue.NewQueue()
	q.Head().Value = &sieveValue{}
	return &Sieve{
		cache:    make(map[interface{}]*queue.Node, cap),
		queue:    q,
		hand:     q.Head(),
		capacity: cap,
	}
}

func (s *Sieve) Miss() int {
	return s.miss
}

func (s *Sieve) Get(key interface{}) (interface{}, bool) {
	if node, exist := s.cache[key]; exist {
		node.Value.(*sieveValue).visited = true
		return node.Value.(*sieveValue).v, true
	}
	s.miss++
	return nil, false
}

func (s *Sieve) Insert(key, value any) {
	if node, exist := s.cache[key]; exist {
		node.Value.(*sieveValue).v = value
		node.Value.(*sieveValue).visited = true
		return
	}
	if len(s.cache) == s.capacity {
		s.evict()
	}
	node := &queue.Node{Value: &sieveValue{key, value, false}}
	s.cache[key] = node
	s.queue.AddToHead(node)
}

func (s *Sieve) evict() *queue.Node {
	for {
		if s.hand == s.queue.Head() {
			s.hand = s.queue.Tail().Prev()
		}
		sv := s.hand.Value.(*sieveValue)
		if sv.visited {
			sv.visited = false
			s.hand = s.hand.Prev()
		} else {
			prev, ghost := s.hand.Prev(), s.hand
			delete(s.cache, sv.k)
			s.queue.RemoveNode(s.hand)
			s.hand = prev
			return ghost
		}
	}
}
