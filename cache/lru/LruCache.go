package lru

type LruCache struct {
	Cache      map[string]*DLinkedNode
	head, tail *DLinkedNode
}

type DLinkedNode struct {
	Key        string
	prev, next *DLinkedNode
}

func initDLinkedNode(key string) *DLinkedNode {
	return &DLinkedNode{
		Key: key,
	}
}

func Constructor() *LruCache {
	l := &LruCache{
		Cache: map[string]*DLinkedNode{},
		head:  initDLinkedNode(""),
		tail:  initDLinkedNode(""),
	}
	l.head.next = l.tail
	l.tail.prev = l.head
	return l
}

func (l *LruCache) Get(key string) string {
	if _, ok := l.Cache[key]; !ok {
		return ""
	}
	node := l.Cache[key]
	l.MoveToHead(node)
	return node.Key
}

func (l *LruCache) Put(key string) {
	if _, ok := l.Cache[key]; !ok {
		node := initDLinkedNode(key)
		l.Cache[key] = node
		l.AddToHead(node)
		//removed := l.RemoveTail()
		//delete(l.Cache, removed.Key)
	} else {
		node := l.Cache[key]
		l.MoveToHead(node)
	}
}

func (l *LruCache) AddToHead(node *DLinkedNode) {
	node.prev = l.head
	node.next = l.head.next
	l.head.next.prev = node
	l.head.next = node
}

func (l *LruCache) RemoveNode(node *DLinkedNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (l *LruCache) MoveToHead(node *DLinkedNode) {
	l.RemoveNode(node)
	l.AddToHead(node)
}

func (l *LruCache) RemoveTail() *DLinkedNode {
	node := l.tail.prev
	l.RemoveNode(node)
	return node
}
func (l *LruCache) AllKeys() (res []string) {
	k := l.head.next
	for k.next != nil {
		res = append(res, k.Key)
		k = k.next
	}
	return
}
