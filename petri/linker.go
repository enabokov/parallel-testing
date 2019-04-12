package petri

import (
	"fmt"
	"strings"
)

type Linker struct {
	Counter       *GlobalCounter
	NamePlace     string
	CounterPlaces int

	NameTransition     string
	CounterTransitions int

	KVariant int
	Info     bool

	Number int
	Label  string
}

type BuildLink interface {
	Build(place *Place, transition *Transition, kVariant int, IsInfo bool, c *GlobalCounter, label string) *Linker

	GetQuantity() int
	SetQuantity(int) BuildLink

	GetNamePlace() string
	SetNamePlace(string) BuildLink

	GetCounterPlaces() int
	SetCounterPlaces(int) BuildLink

	GetNameTransition() string
	SetNameTransition(string) BuildLink

	GetCounterTransitions() int
	SetCounterTransitions(int) BuildLink

	InitNext(*GlobalCounter) BuildLink
	IsInfo() bool
	SetInfo(bool) BuildLink

	PrintInfo()
	PrintParams()

	Clone() BuildLink
}

func (l *Linker) Build(place *Place, transition *Transition, kVariant int, info bool, c *GlobalCounter, label string) *Linker {
	l.CounterPlaces = place.Number
	l.CounterTransitions = transition.Number
	l.KVariant = kVariant
	l.Info = info
	l.NamePlace = place.Name
	l.NameTransition = transition.Name
	l.Label = label

	if label == `i` {
		l.Number = c.LinkIn
	} else if label == `o` {
		l.Number = c.LinkOut
	}
	l.Counter = c
	l.incr()
	return l
}

func (l *Linker) incr() {
	l.Counter.Lock()
	if l.Label == `i` {
		l.Counter.LinkIn++
	} else if l.Label == `o` {
		l.Counter.LinkOut++
	}
	l.Counter.Unlock()
}

func (l *Linker) GetQuantity() int {
	return l.KVariant
}

func (l *Linker) SetQuantity(q int) BuildLink {
	l.KVariant = q
	return l
}

func (l *Linker) GetNamePlace() string {
	return l.NamePlace
}

func (l *Linker) SetNamePlace(n string) BuildLink {
	l.NamePlace = n
	return l
}

func (l *Linker) GetCounterPlaces() int {
	return l.CounterPlaces
}

func (l *Linker) SetCounterPlaces(c int) BuildLink {
	l.CounterPlaces = c
	return l
}

func (l *Linker) GetNameTransition() string {
	return l.NameTransition
}

func (l *Linker) SetNameTransition(n string) BuildLink {
	l.NameTransition = n
	return l
}

func (l *Linker) GetCounterTransitions() int {
	return l.CounterTransitions
}

func (l *Linker) SetCounterTransitions(c int) BuildLink {
	l.CounterTransitions = c
	return l
}

func (l *Linker) InitNext(c *GlobalCounter) BuildLink {
	if l.Label == "i" {
		c.LinkIn = 0
	} else {
		c.LinkOut = 0
	}
	return l
}

func (l *Linker) IsInfo() bool {
	return l.Info
}

func (l *Linker) SetInfo(i bool) BuildLink {
	l.Info = i
	return l
}

func (l *Linker) PrintInfo() {
	fmt.Printf("%s\n%+v\n",
		strings.Repeat("=", 10), l,
	)
}

func (l *Linker) PrintParams() {
	fmt.Println("Test")
}

func (l *Linker) Clone() BuildLink {
	var n Linker
	n = *l
	return &n
}
