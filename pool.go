package godis

import (
    "container/list"
    "sync"
)

var MaxConnections = 50

type connPool struct {
    free chan byte

    mu  sync.RWMutex
    buf *list.List
}

// connection pool is a thread-safe stack of potential connections
// it stores old connection. given that a caller only uses 
// synchronous calls, only a single connection is ever opened
func newConnPool() *connPool {
    p := new(connPool)
    p.free = make(chan byte, MaxConnections)
    p.buf = list.New()

    for i := 0; i < MaxConnections; i++ {
        p.free <- 'c'
    }

    return p
}

// pop will block until a connection is available
// returns *Conn, or nil if connection is yet not 
// created
func (p *connPool) pop() *Conn {
    // wait for semaphore
    <-p.free

    // aquire lock
    p.mu.Lock()
    defer p.mu.Unlock()

    // conn is not created
    if p.buf.Len() == 0 {
        return nil
    }

    e := p.buf.Front()
    t, _ := e.Value.(*Conn)
    p.buf.Remove(e)
    return t
}

func (p *connPool) push(c *Conn) {
    // aquire lock
    p.mu.Lock()
    defer p.mu.Unlock()

    if c != nil {
        p.buf.PushFront(c)
    }

    p.free <- 'c'
}
