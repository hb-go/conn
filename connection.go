package conn

import (
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hb-go/conn/pkg/log"
	"github.com/neverhook/easygo/netpoll"
)

var (
	connPool  sync.Pool
	poolCount int64
)

type Conn struct {
	id   int64
	conn net.Conn
	file *os.File //copy of origin connection fd

	desc *netpoll.Desc

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
	c.conn = conn
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
	c.conn = nil
	c.file = nil
	c.desc = nil
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

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Errorf("conn close error:%v", err)
		}
	}

	// close copied fd
	if c.file != nil {
		if err := c.file.Close(); err != nil {
			log.Errorf("conn file close error:%v", err)
		}
	}

	return
}

func (c *Conn) isClosed() bool {
	return c.closed > 0
}

func (c *Conn) setReading() bool {
	return atomic.CompareAndSwapInt32(&c.reading, 0, 1)
}

func (c *Conn) setReaded() bool {
	return atomic.CompareAndSwapInt32(&c.reading, 1, 0)
}

func (c *Conn) isReading() bool {
	return c.reading > 0
}
