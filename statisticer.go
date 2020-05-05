package main

import (
	"encoding/binary"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	gstats "gonum.org/v1/gonum/stat"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	statisticer = NewStatisticer()
)

// StatisticAction ...
type StatisticAction int

const (
	// ActionBegin ...
	ActionBegin StatisticAction = 1
	// ActionDone ...
	ActionDone StatisticAction = 2
	// ActionFailed ...
	ActionFailed StatisticAction = 3
)

// StatisticType ...
type StatisticType int

const (
	// TypeClient ...
	TypeClient StatisticType = 1
	// TypeSubscribe ...
	TypeSubscribe StatisticType = 2
	// TypePublish ...
	TypePublish StatisticType = 3
	// TypeMessage ...
	TypeMessage StatisticType = 4
)

type statistics struct {
	total   int64
	done    int64
	failed  int64
	elapses []float64
}

type statData struct {
	stype   StatisticType
	action  StatisticAction
	elapsed int64
}

// Statisticer ...
type Statisticer struct {
	closed int32
	done   bool
	datas  chan statData
	stats  map[StatisticType]*statistics
}

func (s *Statisticer) onMessageArriaved(c *Client, msg mqtt.Message) {
	infof("Subscriber %q received message %q", c.opts.ClientID, msg.Topic())
	now := time.Now().UnixNano()
	data := msg.Payload()
	if len(data) < 8 {
		errorf("Got invalid message: %v", msg.Topic())
		return
	}
	begin := int64(binary.BigEndian.Uint64(data))
	elapsed := now - begin
	if elapsed < 0 || elapsed >= now {
		errorf("Got invalid message %q: begin=%d, now=%d", msg.Topic(), begin, now)
		return
	}
	s.AddMessage(ActionDone, elapsed)
}

func (s *Statisticer) add(cate StatisticType, action StatisticAction, elapsed int64) {
	if atomic.LoadInt32(&s.closed) == 0 {
		s.datas <- statData{stype: cate, action: action, elapsed: elapsed}
	}
}

// AddClient ...
func (s *Statisticer) AddClient(action StatisticAction, elapsed int64) {
	s.add(TypeClient, action, elapsed)
}

// AddSubscribe ..
func (s *Statisticer) AddSubscribe(action StatisticAction, elapsed int64) {
	s.add(TypeSubscribe, action, elapsed)
}

// AddPublish ...
func (s *Statisticer) AddPublish(action StatisticAction, elapsed int64) {
	s.add(TypePublish, action, elapsed)
}

// AddMessage ...
func (s *Statisticer) AddMessage(action StatisticAction, elapsed int64) {
	s.add(TypeMessage, action, elapsed)
}

func (s *Statisticer) parseElapses(elapses []float64) {
	sort.Float64s(elapses)
	fmt.Printf("  Min       = %.4f\n", elapses[0])
	fmt.Printf("  Max       = %.4f\n", elapses[len(elapses)-1])
	fmt.Printf("  Average   = %.4f\n", gstats.Mean(elapses, nil))

	fmt.Printf("  Median    = %.4f\n", gstats.Quantile(0.50, gstats.Empirical, elapses, nil))
	fmt.Printf("  Line90    = %.4f\n", gstats.Quantile(0.90, gstats.Empirical, elapses, nil))
	fmt.Printf("  Line95    = %.4f\n", gstats.Quantile(0.95, gstats.Empirical, elapses, nil))
	fmt.Printf("  Line99    = %.4f\n", gstats.Quantile(0.99, gstats.Empirical, elapses, nil))
}

func perSecond(values []float64, total int64) float64 {
	s := 0.0
	for _, x := range values {
		s += x
	}
	s /= 1000
	return float64(total) / s
}

func (s *Statisticer) dump(cate StatisticType, name string) {
	v := s.stats[cate]
	fmt.Printf(" =================== [%s] =================== \n", name)
	fmt.Printf("  Samples   = %d\n", v.total)
	if v.total == 0 {
		return
	}
	pctErr := 0.0
	if v.total > 0 {
		pctErr = float64(v.failed) / float64(v.total) * 100.0
	}
	fmt.Printf("  Error     = %d\n", v.failed)
	fmt.Printf("  ErrorPct  = %.4f%%\n", pctErr)
	if len(v.elapses) == 0 {
		return
	}
	fmt.Printf("  PerSec    = %.4f\n", perSecond(v.elapses, v.done))
	s.parseElapses(v.elapses)
}

// Dump ...
func (s *Statisticer) Dump() {
	// clients
	fmt.Printf("\n\n")

	s.dump(TypeClient, "Client Connections")
	s.dump(TypeSubscribe, "Topic subscribes")
	s.dump(TypePublish, "Message publishs")
	// messages
	v := s.stats[TypeMessage]

	fmt.Printf(" =================== [Message latency] =================== \n")
	total := s.stats[TypePublish].done
	lost := total - v.done
	fmt.Printf("  Received  = %d\n", v.done)
	fmt.Printf("  Lost      = %d\n", lost)
	if v.done == 0 {
		return
	}
	pctLost := float64(lost) / float64(total) * 100
	fmt.Printf("  LostPct   = %.4f%%\n", pctLost)
	if len(v.elapses) == 0 {
		return
	}
	fmt.Printf("  PerSec    = %.4f\n", perSecond(v.elapses, v.done))
	s.parseElapses(v.elapses)
}

// Start ...
func (s *Statisticer) Start() {
	for {
		data, ok := <-s.datas
		if !ok {
			return
		}
		v := s.stats[data.stype]
		switch data.action {
		case ActionBegin:
			v.total++
		case ActionDone:
			atomic.AddInt64(&v.done, 1)
			ms := float64(data.elapsed) / 1000000.0
			v.elapses = append(v.elapses, ms)
		case ActionFailed:
			v.failed++
		}
	}
}

// WaitMsgs ...
func (s *Statisticer) WaitMsgs(n int) {
	if atomic.LoadInt32(&s.closed) == 0 {
		infof("Wating messages for 60 seconds ...")
		for i := 0; i < n; i++ {
			pDone := atomic.LoadInt64(&s.stats[TypePublish].done)
			mDone := atomic.LoadInt64(&s.stats[TypeMessage].done)
			if mDone == pDone {
				infof("All messages arriaved")
				// All messages arriaved
				break
			}
			time.Sleep(time.Second)
		}
	}
}

// Stop ..
func (s *Statisticer) Stop() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.datas)
	}
}

// NewStatisticer ...
func NewStatisticer() *Statisticer {
	s := &Statisticer{
		closed: 0,
		datas:  make(chan statData),
		stats:  make(map[StatisticType]*statistics),
	}
	s.stats[TypeClient] = &statistics{}
	s.stats[TypeSubscribe] = &statistics{}
	s.stats[TypePublish] = &statistics{}
	s.stats[TypeMessage] = &statistics{}
	return s
}

// Samples  Min, Max, Average, Medium, line90, line95, line99, error, throughput, kb/sec
