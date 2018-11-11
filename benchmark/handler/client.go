package handler

import (
	"io"
	"net"
	"time"

	"github.com/surgemq/message"

	"github.com/hb-go/conn/pkg/log"
)

func ClientReader(c net.Conn) (err error) {
	log.Debug("conn read begin")
	begin := time.Now()
	c.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	bts, err := getMessageBuffer(c)
	c.SetReadDeadline(time.Time{})
	if err != nil {
		// TODO temporary错误重试
		log.Errorf("read error:%v", err)
		return
	}

	d := time.Since(begin)
	log.Debugf("received msg:%v", bts)
	log.Debugf("conn read end: %v", d.String())

	return

	mtype := message.MessageType(bts[0] >> 4)
	var msg message.Message
	msg, err = mtype.New()
	if err != nil {
		log.Errorf("message new err: %v, bts: %v", err, bts)
		return
	}

	_, err = msg.Decode(bts)
	if err != nil {
		log.Errorf("message decode err: %v, bts: %v", err, bts)
		return
	}

	log.Debugf("received msg:%v", msg)

	d = time.Since(begin)
	log.Debugf("conn decode end: %v", d.String())

	switch {
	case err == io.EOF:
		break

	case err != nil:
		log.Errorf("read error:%v", err)
		break

	default:

	}

	return err
}
