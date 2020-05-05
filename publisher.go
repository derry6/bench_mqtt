package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// Publisher ...
type Publisher struct {
	topic    string
	qos      int
	count    int
	size     int
	interval int
	cliOpts  *ClientOptions
}

func (p *Publisher) getPayload(msgSize int) []byte {
	now := time.Now().UnixNano()
	data := bytes.Repeat([]byte{'#'}, msgSize)
	binary.BigEndian.PutUint64(data, uint64(now))
	return data
}

// Run ...
func (p *Publisher) Run(wg *sync.WaitGroup, cleanup func(c *Client)) {
	defer wg.Done()
	cli, err := NewClient(p.cliOpts)
	if err != nil {
		// errorf("[%q] create failed: %v", p.cliOpts.ClientID, err)
		return
	}
	cleanup(cli)

	for i := 0; i < p.count; i++ {
		data := p.getPayload(p.size)
		begin := time.Now().UnixNano()

		statisticer.AddPublish(ActionBegin, 0)
		topic := fmt.Sprintf("%s/%d", p.topic, i)
		token := cli.MQTT().Publish(topic, byte(p.qos), false, data)
		if token.Wait() && token.Error() != nil {
			errorf("[%q] publish %q failed: %v", p.cliOpts.ClientID, topic, token.Error())
			statisticer.AddPublish(ActionFailed, 0)
			continue
		}
		infof("Publisher %q publish %q done", p.cliOpts.ClientID, topic)
		statisticer.AddPublish(ActionDone, time.Now().UnixNano()-begin)
		time.Sleep(time.Duration(p.interval) * time.Millisecond)
	}
}

// NewPublisher ...
func NewPublisher(cliOpts *ClientOptions, topic string, qos, count, msgSize, interval int) *Publisher {
	p := &Publisher{
		cliOpts:  cliOpts,
		topic:    topic,
		qos:      qos,
		count:    count,
		size:     msgSize,
		interval: interval,
	}
	return p
}
