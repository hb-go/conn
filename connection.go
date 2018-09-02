package conn

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hb-go/conn/pkg/log"
)

var (
	connPool  sync.Pool
	poolCount int64
)

type Conn struct {
	id   int64
	Conn net.Conn

	reading int32
	closed  int32

	timestamp time.Time
}

func init() {
	connPool = sync.Pool{
		New: func() interface{} {
			poolCount++
			c := Conn{
				id: poolCount,
			}

			return &c
		},
	}
}

func getConnection(conn net.Conn) *Conn {
	c := connPool.Get().(*Conn)
	c.Conn = conn
	c.closed = 0
	return c
}

func (c *Conn) closeAndPut() {
	c.close()
	c.put()
}

func (c *Conn) put() {
	c.reset()
	connPool.Put(c)
}

func (c *Conn) reset() {
	c.Conn = nil
	c.closed = 0
	c.reading = 0
}

// set closed flag，conn.Close()
func (c *Conn) close() {
	doit := atomic.CompareAndSwapInt32(&c.closed, 0, 1)

	if !doit {
		// 关闭一个已关闭的Connection
		return
	}

	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			log.Errorf("conn close error:%v", err)
		}
	}

	return
}

func (c *Conn) isClosed() bool {
	return c.closed > 0
}

func (c *Conn) setReading() bool {
	doit := atomic.CompareAndSwapInt32(&c.reading, 0, 1)

	return doit
}

func (c *Conn) setReaded() bool {
	return atomic.CompareAndSwapInt32(&c.reading, 1, 0)
}

func (c *Conn) isReading() bool {
	return c.reading > 0
}
