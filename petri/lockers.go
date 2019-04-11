package petri

import "sync"

type GlobalCounter struct {
	Mux        sync.Mutex
	Link       int
	Place      int
	Transition int
	Simulator  int
}

type GlobalTime struct {
	Mux         sync.Mutex
	CurrentTime float64
	ModTime     float64
}

type GlobalLocker struct {
	Cond *sync.Cond
}
