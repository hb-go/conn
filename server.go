package conn

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/neverhook/easygo/netpoll"

	"github.com/hb-go/conn/pkg/gopool"
	"github.com/hb-go/conn/pkg/log"
)

var (
	ErrServerClosed = errors.New("server: server closed")
)

var (
	poller    netpoll.Poller
	readPool  *gopool.Pool
	writePool *gopool.Pool
)

type (
	Option func(*Options)

	ConnHandler      func(conn *Conn) error
	ConnReadHandler  func(conn net.Conn) error
	ConnWriteHandler func(conn net.Conn, p []byte) (int, error)
)

func init() {

	readPool = gopool.NewPool(512, 256, 10)
	writePool = gopool.NewPool(512, 256, 10)

	var err error
	poller, err = netpoll.New(&netpoll.Config{
		OnWaitError: func(err error) {
			log.Errorf("net poll wait error: %v", err)
		}})
	if err != nil {
		panic(err)
	}
}

type Server struct {
	opts Options

	ln net.Listener

	mu       sync.RWMutex
	doneChan chan struct{}

	ConnHandler  ConnHandler
	ReadHandler  ConnReadHandler
	WriteHandler ConnWriteHandler
}

func (srv *Server) ListenAndServe(network, address string) error {
	ln, err := net.Listen(network, address)
	if err != nil {
		log.Errorf("listen and serve error: %v", err)
		return err
	}

	defer func() {
		srv.Close()
	}()

	var tempDelay time.Duration

	srv.mu.Lock()
	srv.ln = ln
	srv.mu.Unlock()

	for {
		conn, e := ln.Accept()
		if e != nil {
			log.Infof("accept error: %v", e)

			select {
			case <-srv.getDoneChan():
				return ErrServerClosed
			default:
			}

			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				log.Errorf("accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0

		err = srv.handleConn(conn)
		if err != nil {
			log.Errorf("conn handle error: %v", err)
		}
	}

	return nil
}

func (srv *Server) Close() error {
	srv.ln.Close()

	return nil
}

func (srv *Server) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.getDoneChanLocked()
}

func (srv *Server) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *Server) closeDoneChanLocked() {
	ch := srv.getDoneChanLocked()
	select {
	case <-ch:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer, guarded
		// by s.mu.
		close(ch)
	}
}

func NewServer(opts ...Option) (*Server, error) {
	options := newOptions(opts...)
	return &Server{
		opts: options,
	}, nil
}
