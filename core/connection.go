package core

import (
	"log"
	"sync"
	"time"
)

// ConnStore - connection store instance
type ConnStore struct {
	sync.RWMutex

	conn map[string]*connection

	// flag for auto garbage collector(close and cleanup expired connection)
	activeGC bool
	// controls gc intervals
	gcTicker *time.Ticker
	// gc stop signal
	done chan struct{}
}

type connection struct {
	// hold connection object
	resource *Resource
	// keep connect
	keepAlive bool
	// will automatically close connection
	expired time.Time
}

// NewConnectionStore - constructor connection store
func NewConnectionStore() *ConnStore {
	return &ConnStore{
		conn: make(map[string]*connection),
	}
}

// StartGC - start gc ticker
func (c *ConnStore) StartGC(interval time.Duration) {
	if interval <= 0 {
		return
	}

	c.activeGC = true

	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	c.Lock()
	c.gcTicker = ticker
	c.done = done
	c.Unlock()

	go func() {
		for {
			select {
			case <-ticker.C:
				c.Lock()
				for key := range c.conn {
					if c.isExpired(key) {
						c.delete(key)
						log.Printf("Connection %s expired and closed\n", key)
					}
				}
				c.Unlock()
			case <-done:
				return
			}
		}
	}()
}

// StopGC stops sweeping goroutine.
func (c *ConnStore) StopGC() {
	if c.activeGC {
		c.Lock()
		c.gcTicker.Stop()
		c.gcTicker = nil
		close(c.done)
		c.done = nil
		c.Unlock()
	}
}

func (c *ConnStore) extend(key string, ttl time.Duration) {
	conn := c.conn[key]
	if conn != nil {
		conn.extendConnection(ttl)
	}
}

func (c *ConnStore) isExpired(key string) bool {
	conn := c.conn[key]
	if conn == nil {
		return false
	}
	return !conn.keepAlive && conn.expired.Before(time.Now())
}

func (c *ConnStore) getAllConn() map[string]*connection {
	return c.conn
}

func (c *ConnStore) delete(key string) {
	conn := c.conn[key]
	if conn != nil {
		conn.close()
		delete(c.conn, key)
	}
}

func (c *ConnStore) addConnection(host string, res *Resource, ttl ...time.Duration) {
	conn := &connection{
		resource:  res,
		keepAlive: true,
	}

	if len(ttl) > 0 {
		conn.keepAlive = false
		conn.expired = time.Now().Add(ttl[0])
	}

	c.Lock()
	c.conn[host] = conn
	c.Unlock()
}

func (c *ConnStore) getConnection(host string) (res *Resource, found bool) {
	conn, ok := c.conn[host]
	if ok && conn.resource != nil {
		found = true
		res = conn.resource
	}
	return
}

func (conn *connection) extendConnection(ttl time.Duration) {
	conn.keepAlive = false
	conn.expired = time.Now().Add(ttl)
}

func (conn *connection) close() {
	conn.resource.Close()
}
