package main

import (
	"fmt"
	"sync"
	"time"
)

// Subscriber ...
type Subscriber struct {
	topic    string
	qos      int
	count    int
	interval int
	cliOpts  *ClientOptions
}

// NewSubscriber create subscriber
func NewSubscriber(opts *ClientOptions, topic string, qos int, count int, interval int) *Subscriber {
	return &Subscriber{
		cliOpts:  opts,
		topic:    topic,
		qos:      qos,
		count:    count,
		interval: interval,
	}
}

// Run ...
func (s *Subscriber) Run(wg *sync.WaitGroup, cleanup func(c *Client)) {
	defer wg.Done()
	cli, err := NewClient(s.cliOpts)
	if err != nil {
		// errorf("Failed to create client %q: %v", s.cliOpts.ClientID, err)
		return
	}

	cleanup(cli)

	for i := 0; i < s.count; i++ {
		begin := time.Now().UnixNano()
		topic := fmt.Sprintf("%s/%d", s.topic, i)
		statisticer.AddSubscribe(ActionBegin, 0)
		token := cli.MQTT().Subscribe(topic, byte(s.qos), nil)
		if token.Wait() && token.Error() != nil {
			// net.OpError{}.Err
			errorf("[%q] subscribe %q failed: %v", s.cliOpts.ClientID, topic, token.Error())
			statisticer.AddSubscribe(ActionFailed, 0)
			continue
		}
		infof("Subscriber %q subscribe %q done", s.cliOpts.ClientID, topic)
		statisticer.AddSubscribe(ActionDone, time.Now().UnixNano()-begin)
		time.Sleep(time.Duration(s.interval) * time.Millisecond)
	}
}
