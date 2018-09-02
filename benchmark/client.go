package main

import (
	"time"

	"strconv"

	"sync"

	"math/rand"

	"github.com/hb-go/conn"
	"github.com/hb-go/conn/benchmark/handler"
	"github.com/hb-go/conn/pkg/log"
)

func main() {

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)

		client, _ := conn.NewClient("127.0.0.1:8080")
		client.WriteHandler = handler.Writer

		if err := client.Dial(); err != nil {
			log.Panicf("client dial error: %v", err)
		}
		go func() {
			i := 0
			msg := []byte("hello")
			payload := make([]byte, 32)

			d := time.Second * 60
			timeout := time.After(d)

			begin := time.Time{}
		loop:
			for {
				select {
				case <-timeout:
					wg.Done()
					break loop
				default:
					if begin.IsZero() {
						begin = time.Now()
					}
					copy(payload, append(msg, []byte(strconv.Itoa(i))...))
					_, err := client.Send(payload)
					i++
					if err != nil {
						log.Errorf("send message failed", err)
					} else {
						log.Infof("send message success")
					}

					err = handler.ClientReader(client.Conn)
					if err != nil {
						log.Errorf("receive message error: %v", err)
					}

					time.Sleep(time.Second * time.Duration(rand.Int63n(10)))
				}
			}

			duration := time.Since(begin).Seconds()

			log.Infof("tps: %f/s, num: %d, duration: %f", float64(i)/duration, i, duration)

		}()
	}

	wg.Wait()
}
