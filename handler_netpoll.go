// +build netpoll

package conn

import (
	"io"
	"net"
	"time"

	"github.com/mailru/easygo/netpoll"

	"github.com/hb-go/conn/pkg/log"
)

func (srv *Server) handleConn(conn net.Conn) (err error) {

	c := getConnection(conn)
	desc, e := netpoll.HandleRead(conn)
	if e != nil {
		err = e
		c.closeAndPut()
		return
	}

	srv.conns.Store(desc, c)
	defer func() {
		if err != nil {
			c.close()
			srv.conns.Delete(desc)
			c.put()

			if e := poller.Stop(desc); e != nil {
				log.Errorf("poller stop error: %v", e)
			}
			if e := desc.Close(); e != nil {
				log.Errorf("poller desc close error: %v", e)
			}
		}
	}()

	err = poller.Start(desc, func(event netpoll.Event) {

		log.Debugf("net poll event: %v", event)
		if event&netpoll.EventRead == 0 {
			return
		}

		// Map中取Connection
		v, ok := srv.conns.Load(desc)
		if !ok {
			return
		}
		c, ok := v.(*Conn)
		if !ok {
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
				srv.conns.Delete(desc)
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
			srv.conns.Delete(desc)
			c.put()

			poller.Stop(desc)
			desc.Close()

			return
		}

		// read process
		usePool := true
		if usePool {
			log.Infof("(connId=%d) reading pre lock: %d", c.id, c.reading)
			if c.setReading() {
				c.timestamp = time.Now()
				log.Infof("(connId=%d) reading lock: %d", c.id, c.reading)
				err := readPool.ScheduleTimeout(time.Microsecond*10, func() {
					if v, ok := srv.conns.Load(desc); ok {
						c := v.(*Conn)
						if !c.isClosed() {
							err := srv.ReadHandler(c)
							c.setReaded()
							log.Infof("(connId=%d) reading unlock duration: %v", c.id, time.Since(c.timestamp).String())
							if err != nil {
								if e, ok := err.(net.Error); ok && (e.Timeout() || e.Temporary()) {
									log.Warnf("read process handle error: %v", err)
								} else {
									log.Errorf("read process handle error: %v", err)
								}

								if err == io.EOF {
									c.close()
								}
							}
						}
					}
				})
				if err != nil {
					c.setReaded()
					log.Infof("(connId=%d) reading unlock", c.id)
					log.Errorf("read pool schedule error: %v", err)
				}
			}

		} else {
			go func() {
				if v, ok := srv.conns.Load(desc); ok {
					c := v.(*Conn)
					err := srv.ReadHandler(c)
					if err != nil {
						log.Errorf("read process handle error: %v", err)
						if err == io.EOF {
							c.close()
						}
					}
				}
			}()
		}

		if err := poller.Resume(desc); err != nil {
			log.Panicf("poller.Resume() error: %v", err)
		}
	})

	return
}
