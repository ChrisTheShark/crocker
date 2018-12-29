package main

import (
	"fmt"
	"net"
	"time"
)

func waitForNetwork() error {
	maxWait := time.Second * 5
	checkInterval := time.Second
	timeStarted := time.Now()

	for {
		if interfaces, err := net.Interfaces(); err != nil {
			return err
		} else if len(interfaces) > 1 {
			return nil
		}

		if time.Since(timeStarted) > maxWait {
			return fmt.Errorf("timeout after %d seconds waiting for network", checkInterval)
		}

		time.Sleep(checkInterval)
	}
}
