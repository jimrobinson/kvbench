package main

import (
	"log"
	"sync"
	"time"
)

// WatchDog provides a mechanism to time out on
// a benchmark if too much time has elapsed since
// for an operation.
type WatchDog struct {
	sync.Mutex

	// aborted is set to true when a
	// timeout has been reached
	aborted bool
}

func NewWatchDog() *WatchDog {
	guard := &WatchDog{
		aborted: false,
	}
	return guard
}

// Timer returns a channel that will return
// true if the max time has been reached
func (guard *WatchDog) Timer(dur time.Duration) (ch <-chan time.Time) {
	return time.NewTimer(dur).C
}

// Abort sets the watchdog to the aborted state,
// indicating that a timeout has been triggered.
func (guard *WatchDog) Abort(msg string) {
	guard.Lock()
	guard.aborted = true
	guard.Unlock()
	log.Println(msg)
}

// IsAborted returns true if the watchdog
// has been aborted due to a timeout being
// reached.
func (guard *WatchDog) IsAborted() bool {
	guard.Lock()
	defer guard.Unlock()
	return guard.aborted
}
