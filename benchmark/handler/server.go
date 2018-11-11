package handler

import (
	"net"
	"time"

	"github.com/surgemq/message"

	"github.com/hb-go/conn"
	"github.com/hb-go/conn/pkg/log"
)

func OnConn(c *conn.Conn) error {

	return nil
}

func Reader(c net.Conn) (err error) {
	log.Debugf("conn read begin")
	begin := time.Now()
	c.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	bts, err := getMessageBuffer(c)
	c.SetReadDeadline(time.Time{})
	if err != nil {
		// TODO temporary错误重试
		log.Errorf("read error:%v", err)
		return
	}

	mtype := message.MessageType(bts[0] >> 4)
	var msg message.Message
	msg, err = mtype.New()
	if err != nil {
		return
	}

	_, err = msg.Decode(bts)
	if err != nil {
		return
	}

	log.Debugf("(received msg:%v", msg)

	d := time.Since(begin)
	log.Debugf("conn read end: %v", d.String())

	err = writeMessage(c, msg)
	if err != nil {
		log.Errorf("write message error:%v", err)
	}

	d = time.Since(begin)
	log.Debugf("conn write end: %v", d.String())

	return err
}

func Writer(c net.Conn, p []byte) (n int, err error) {
	msg := message.NewPublishMessage()
	msg.SetPayload(p)
	msg.SetTopic([]byte("topic"))
	return 0, writeMessage(c, msg)
}
