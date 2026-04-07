package main

import "sync"

var linuxListenerOnce sync.Once

func startGlobalListener() {
	linuxListenerOnce.Do(func() {
		// TODO: Linux implementation
	})
}
