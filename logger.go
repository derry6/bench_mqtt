package main

import (
	"fmt"
)

func infof(format string, args ...interface{}) {
	if !cfg.quite {
		fmt.Printf(format+"\n", args...)
	}
}

func errorf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}
