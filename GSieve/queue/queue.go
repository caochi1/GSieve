package queue

import "fmt"

type Node struct {
	prev, next *Node
	Value      interface{}
}

func (node *Node) Prev() *Node {
	return node.prev
}

func (node *Node) Next() *Node {
	return node.next
}

type Queue struct {
	head, tail *Node
}

func NewQueue() *Queue {
	head, tail := &Node{}, &Node{}
	head.next, tail.prev = tail, head
	return &Queue{head: head, tail: tail}
}

func (q *Queue) AddToHead(node *Node) {
	node.prev = q.head
	node.next = q.head.next
	q.head.next, q.head.next.prev = node, node
}

func (q *Queue) AddToTail(node *Node) {
	node.prev = q.tail.prev
	node.next = q.tail
	q.tail.prev, q.tail.prev.next = node, node
}

func (q *Queue) RemoveNode(node *Node) {
	node.prev.next = node.next
	node.next.prev = node.prev
	node.prev, node.next = nil, nil
}

func (q *Queue) MoveToHead(node *Node) {
	if node.prev != q.head {
		q.RemoveNode(node)
		q.AddToHead(node)
	}
}

func (q *Queue) ForEach() {
	p := q.head
	for p != nil {
		fmt.Println(p.Value)
		p = p.next
	}
}

func (q *Queue) Head() *Node {
	return q.head
}

func (q *Queue) Tail() *Node {
	return q.tail
}
