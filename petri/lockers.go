package petri

import "sync"

type GlobalCounter struct {
	sync.Mutex
	LinkIn     int
	LinkOut    int
	Place      int
	Transition int
	Simulator  int
}

type GlobalTime struct {
	sync.Mutex
	CurrentTime float64
	ModTime     float64
}

type GlobalLocker struct {
	Cond *sync.Cond
}
