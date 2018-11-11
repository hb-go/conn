// +build !netpoll

package conn

import (
	"io"
	"net"

	"github.com/hb-go/conn/pkg/log"
)

func (srv *Server) handleConn(conn net.Conn) (err error) {

	c := getConnection(conn)
	go func() {
		for {
			err := srv.ReadHandler(c.conn)
			if err != nil {
				log.Errorf("read process handle error: %v", err)
				if err == io.EOF {
					c.close()
					break
				}
			}
		}
	}()

	return
}
