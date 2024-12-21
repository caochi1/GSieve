package gsieve

import (
	"main/queue"
	"sync"
	"time"
)

// cronJobTime以ms为单位
func NewTTLGSieve(cap, timeInt, cronJobTime int) *GSieve {
	gs := &GSieve{
		capacity:     cap,
		timeInterval: timeInt,
		index:        make(map[string]*syncSieve, cap),
		ghostList:    NewsyncLRUCache(cap / 2),
	}
	go gs.cronJob(cronJobTime)
	return gs
}

type GSieve struct {
	ttlBuckets   [1024]*syncSieve
	index        map[string]*syncSieve //将key映射至对应队列
	timeInterval int                   //时间间隔(例如ttl 1~4，timeInterval为4)
	ghostList    *syncLRU
	lock         sync.RWMutex
	hand         *queue.Node
	capacity     int
}

type gsieveValue struct {
	k          string
	v          []byte
	visited    bool
	Expiration int64
}

func (gsieve *GSieve) Get(key string) (value []byte, isExist bool) {
	sieve, isExist := gsieve.getSieve(key)
	if isExist {
		return sieve.get(key), isExist
	} else {
		gsieve.lock.Lock()
		defer gsieve.lock.Unlock()
		if node, ok := gsieve.ghostList.get(key); ok {
			exp := node.Value.(*gsieveValue).Expiration
			gsieve.ghostList.del(key)
			if time.Now().Unix() >= exp && exp != 0 {
				return nil, false
			}
			gsieve.respawn(node)
			return node.Value.(*gsieveValue).v, true
		}
		return nil, false
	}
}

func (gsieve *GSieve) Insert(key string, value []byte, ttl int) {
	newIndex := gsieve.newIndex(ttl)
	node := &queue.Node{Value: &gsieveValue{key, value, false, expiration(ttl)}}
	gsieve.lock.Lock()
	defer gsieve.lock.Unlock()
	if _, ok := gsieve.ghostList.get(key); ok {
		gsieve.ghostList.del(key)
	}
	gsieve.lazyInit(newIndex)
	if oldSieve, isExist := gsieve.index[key]; isExist { //object update
		oldSieve.del(key)
	} else if len(gsieve.index) == gsieve.capacity {
		gsieve.evict()
	}
	//insert begin
	s := gsieve.ttlBuckets[newIndex]
	gsieve.index[key] = s
	s.set(node)
}

func (gsieve *GSieve) Delete(key string) {
	gsieve.lock.Lock()
	defer gsieve.lock.Unlock()
	if _, isExist := gsieve.index[key]; isExist {
		gsieve.del(key)
	}

}

func (gsieve *GSieve) respawn(node *queue.Node) {
	key, ttl := node.Value.(*gsieveValue).k, int(node.Value.(*gsieveValue).Expiration)
	newIndex := gsieve.newIndex(ttl)
	gsieve.lazyInit(newIndex)
	if len(gsieve.index) == gsieve.capacity {
		gsieve.evict()
	}
	//insert begin
	s := gsieve.ttlBuckets[newIndex]
	gsieve.index[key] = s
	s.set(node)
}

func (gsieve *GSieve) getSieve(key string) (*syncSieve, bool) {
	gsieve.lock.RLock()
	defer gsieve.lock.RUnlock()
	sieve, isExist := gsieve.index[key]
	return sieve, isExist
}

func (gsieve *GSieve) newIndex(ttl int) int {
	last := len(gsieve.ttlBuckets) - 1
	index := min(ttl/gsieve.timeInterval, last)
	if ttl == 0 {
		index = last
	}
	return index
}

func (gsieve *GSieve) evict() {
	if gsieve.hand == nil {
		gsieve.hand = gsieve.find(0)
	}
	for {
		//走到头部表示当前队列遍历完毕，前往下一个队列
		if gsieve.hand.Prev() == nil {
			gsieve.hand = gsieve.find(gsieve.hand.Value.(levelValue).level + 1)
		}
		gsv := gsieve.hand.Value.(*gsieveValue)
		if gsv.visited {
			gsv.visited = false
			gsieve.hand = gsieve.hand.Prev()
		} else {
			prev, ghost := gsieve.hand.Prev(), gsieve.hand
			gsieve.del(gsv.k)
			gsieve.ghostList.put(ghost)
			gsieve.hand = prev
			return
		}
	}
}

// find the next level
func (gsieve *GSieve) find(level int) *queue.Node {
	for {
		if level == len(gsieve.ttlBuckets) {
			level = 0
			continue
		}
		if gsieve.ttlBuckets[level] == nil || len(gsieve.ttlBuckets[level].cache) == 0 { //sieve is nil or list is empty
			level++
		} else {
			return gsieve.ttlBuckets[level].tail().Prev()
		}
	}
}

func (gsieve *GSieve) del(key string) {
	s := gsieve.index[key]
	s.queue.RemoveNode(s.cache[key])
	delete(s.cache, key)
	delete(gsieve.index, key)
}

func (gsieve *GSieve) cleanup() {
	var s *syncSieve
	for i := 0; i < len(gsieve.ttlBuckets)-2; i++ {
		s = gsieve.ttlBuckets[i]
		if s != nil {
			s.cleanup(gsieve)
		}
	}
}

func (gsieve *GSieve) cronJob(t int) {
	ticker := time.NewTicker(time.Duration(t) * time.Millisecond)
	for {
		<-ticker.C
		gsieve.cleanup()
	}
}

func (gsieve *GSieve) lazyInit(level int) {
	if gsieve.ttlBuckets[level] == nil {
		gsieve.ttlBuckets[level] = newSieve(level)
	}
}

func (gsieve *GSieve) ForEach() [][]byte {
	res := [][]byte{}
	for k := range gsieve.index {
		res = append(res, []byte(k))
	}
	return res
}

func expiration(ttl int) int64 {
	if ttl == 0 {
		return 0
	}
	return time.Now().Add(time.Duration(ttl) * time.Second).Unix()
}

// func expiration2ttl(expiration int64) int64 {
// 	if expiration == 0 {
// 		return 0
// 	}
// 	return time.Now().Unix() - expiration
// }
