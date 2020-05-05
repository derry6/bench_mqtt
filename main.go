package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// ClientManager ...
type ClientManager struct {
	closed  int32
	clients []*Client
	events  chan *Client
}

// Add ...
func (m *ClientManager) Add(cli *Client) {
	if atomic.LoadInt32(&m.closed) == 0 {
		m.events <- cli
	}
}

// Close ...
func (m *ClientManager) Close() {
	if atomic.CompareAndSwapInt32(&m.closed, 0, 1) {
		close(m.events)
		for _, cli := range m.clients {
			cli.Close()
		}
	}
}

// NewClientManager ...
func NewClientManager() *ClientManager {
	m := &ClientManager{
		closed:  0,
		clients: make([]*Client, 0),
		events:  make(chan *Client),
	}
	go func() {
		for {
			cli, ok := <-m.events
			if !ok {
				return
			}
			m.clients = append(m.clients, cli)
		}
	}()
	return m
}

func main() {
	if err := parseConfig(); err != nil {
		log.Fatalf("Invalid configurations: %v", err)
	}
	infof("Starting ...")
	manager := NewClientManager()
	defer manager.Close()

	// Run statisticer
	go statisticer.Start()

	n := 5

	{
		var wg1 sync.WaitGroup
		subNum := cfg.num / 2
		infof("Creating %d subscribers ...", subNum)
		for i := 0; i < subNum; i++ {
			name := fmt.Sprintf("%s/sub/%d", cfg.usr, i)
			cliOpts := &ClientOptions{
				Broker:    cfg.broker,
				Username:  name,
				Password:  cfg.pwd,
				ClientID:  name,
				OnMessage: statisticer.onMessageArriaved,
			}
			topic := fmt.Sprintf("%s/%d", cfg.topic, i)
			sub := NewSubscriber(cliOpts, topic, cfg.qos, cfg.msgCnt, cfg.interval)
			wg1.Add(1)
			go sub.Run(&wg1, manager.Add)
			time.Sleep(time.Duration(cfg.interval) * time.Millisecond)
		}
		wg1.Wait()
		infof("All topics subscribed.")
	}

	{
		var wg2 sync.WaitGroup
		pubNum := cfg.num / 2
		infof("Creating %d publishers ...", pubNum)
		for i := 0; i < pubNum; i++ {
			name := fmt.Sprintf("%s/pub/%d", cfg.usr, i)
			cliOpts := &ClientOptions{
				Broker:   cfg.broker,
				Username: name,
				Password: cfg.pwd,
				ClientID: name,
			}
			topic := fmt.Sprintf("%s/%d", cfg.topic, i)
			pub := NewPublisher(cliOpts, topic, cfg.qos, cfg.msgCnt, cfg.msgSize, cfg.interval)
			wg2.Add(1)
			go pub.Run(&wg2, manager.Add)
			time.Sleep(time.Duration(cfg.interval) * time.Millisecond)
		}
		wg2.Wait()
		infof("All messages published.")
	}
	statisticer.WaitMsgs(n)
	statisticer.Dump()
}
