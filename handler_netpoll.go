// +build netpoll

package conn

import (
	"io"
	"net"
	"time"

	"github.com/neverhook/easygo/netpoll"

	"github.com/hb-go/conn/pkg/log"
)

func (srv *Server) handleConn(conn net.Conn) (err error) {

	c := getConnection(conn)
	if tc, ok := conn.(*net.TCPConn); ok {
		c.file, err = tc.File()
		if err != nil {
			return
		}
	}

	desc, e := netpoll.HandleFile(c.file, netpoll.EventRead|netpoll.EventOneShot)
	if e != nil {
		err = e
		c.closeAndPut()
		return
	}

	c.desc = desc
	defer func() {
		if err != nil {
			c.close()
			c.put()

			if e := poller.Stop(desc); e != nil {
				log.Errorf("poller stop error: %v", e)
			}
			if e := desc.Close(); e != nil {
				log.Errorf("poller desc close error: %v", e)
			}
		}
	}()

	rf := func() {
		if c.desc == desc {
			defer func() {
				if !c.isClosed() {
					if err := poller.Resume(desc); err != nil {
						log.Errorf("poller.Resume() error: %v", err)
					}
				}
			}()

			if !c.isClosed() {
				err := srv.ReadHandler(c.conn)
				c.setReaded()
				log.Debugf("(connId=%d) reading unlock duration: %v", c.id, time.Since(c.timestamp).String())
				if err != nil {
					if e, ok := err.(net.Error); ok && (e.Timeout() || e.Temporary()) {
						log.Warnf("read process handle error: %v", err)
					} else {
						log.Errorf("read process handle error: %v", err)
						c.close()
					}
				}
			}
		}
	}

	err = poller.Start(desc, func(event netpoll.Event) {

		log.Debugf("net poll event: %v", event)
		if event&netpoll.EventRead == 0 {
			return
		}

		if c.desc != desc {
			poller.Stop(desc)
			desc.Close()

			return
		}

		defer func() {
			// recover from panic
			if r := recover(); r != nil {
				if !c.isClosed() {
					c.close()
				}
				c.put()

				poller.Stop(desc)
				desc.Close()

				log.Errorf("poller event panic: %v", r)
			}
		}()

		if c.isClosed() || event&netpoll.EventReadHup == netpoll.EventReadHup {
			if !c.isClosed() {
				c.close()
			}
			c.put()

			poller.Stop(desc)
			desc.Close()

			return
		}

		// read process
		usePool := true
		if usePool {
			log.Debugf("(connId=%d) reading pre lock: %d", c.id, c.reading)

			// TODO 处于Reading状态的Conn不添加新任务
			if c.setReading() {
				c.timestamp = time.Now()
				log.Debugf("(connId=%d) reading lock: %d", c.id, c.reading)

				if err := readPool.ScheduleTimeout(time.Microsecond*100, rf); err != nil {
					// TODO 任务队列timeout处理
					c.setReaded()
					log.Debugf("(connId=%d) reading unlock", c.id)
					log.Errorf("read pool schedule error: %v", err)

					if err := poller.Resume(desc); err != nil {
						log.Errorf("poller.Resume() error: %v", err)
					}
				}
			} else {
				log.Warnf("(connId=%d) receive read event while reading locked", c.id)
			}
		} else {
			go func() {
				if c.desc == desc {
					err := srv.ReadHandler(c.conn)
					if err != nil {
						log.Errorf("read process handle error: %v", err)
						if err == io.EOF {
							c.close()
						}
					}
				}

				if err := poller.Resume(desc); err != nil {
					log.Panicf("poller.Resume() error: %v", err)
				}
			}()
		}
	})

	return
}
