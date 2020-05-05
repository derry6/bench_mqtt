package main

import (
	"flag"
	"fmt"
)

var (
	cfg configuration
)

type configuration struct {
	broker   string
	usr      string
	pwd      string
	topic    string
	qos      int
	msgSize  int
	msgCnt   int
	num      int
	quite    bool
	interval int
}

func parseConfig() error {
	flag.StringVar(&cfg.broker, "b", "localhost:1883", "broker address")
	flag.StringVar(&cfg.usr, "u", "bench", "username prefix")
	flag.StringVar(&cfg.pwd, "p", "cGFzc3dvcmQK", "password")
	flag.StringVar(&cfg.topic, "t", "bench-topic", "topic prefix")
	flag.IntVar(&cfg.qos, "x", 1, "qos")
	flag.IntVar(&cfg.msgSize, "s", 256, "payload size of message")
	flag.IntVar(&cfg.msgCnt, "m", 1, "msgs per client")
	flag.IntVar(&cfg.num, "c", 100, "total clients")
	flag.BoolVar(&cfg.quite, "q", false, "quite")
	flag.IntVar(&cfg.interval, "i", 500, "interval (ms)")
	flag.Parse()

	if cfg.broker == "" {
		return fmt.Errorf("empty server address")
	}

	if cfg.msgSize < 8 {
		return fmt.Errorf("min size of message is 8")
	}
	if cfg.msgCnt < 1 {
		return fmt.Errorf("no msg")
	}
	if cfg.num < 2 {
		return fmt.Errorf("no clients")
	}
	// cfg.num = cfg.num / 2

	if cfg.qos != 0 && cfg.qos != 1 && cfg.qos != 2 {
		return fmt.Errorf("qos must be in range [0, 1, 2]")
	}
	if cfg.interval < 0 {
		cfg.interval = 500
	}
	return nil
}
