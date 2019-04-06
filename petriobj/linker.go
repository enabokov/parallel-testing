package petriobj

import (
	"fmt"
	"strings"
)

type Linker struct {
	namePlace string
	counterPlaces int

	nameTransition string
	counterTransitions int

	kVariant int
	info bool

	next int
	number int
}

type BuildLink interface {
	build(namePlace string, counterPlaces int, nameTransition string, counterTransitions int, kVariant int, isInfo bool) BuildLink

	getQuantity() int
	setQuantity(int) BuildLink

	getNamePlace() string
	setNamePlace(string) BuildLink

	getCounterPlaces() int
	setCounterPlaces(int) BuildLink

	getNameTransition() string
	setNameTransition(string) BuildLink

	getCounterTransitions() int
	setCounterTransitions(int) BuildLink

	initNext() BuildLink
	isInfo() bool
	setInfo(bool) BuildLink

	printInfo()
	printParams()

	clone() BuildLink
}

func (l *Linker) build(namePlace string, counterPlaces int, nameTransition string, counterTransitions int, kVariant int, info bool) BuildLink {
	l.namePlace = namePlace
	l.counterPlaces = counterPlaces
	l.nameTransition = nameTransition
	l.counterTransitions = counterTransitions
	l.kVariant = kVariant
	l.info = info
	l.number = l.next
	l.next++
	return l
}

func (l *Linker) getQuantity() int {
	return l.kVariant
}

func (l *Linker) setQuantity(q int) BuildLink {
	l.kVariant = q
	return l
}

func (l *Linker) getNamePlace() string {
	return l.namePlace
}

func (l *Linker) setNamePlace(n string) BuildLink {
	l.namePlace = n
	return l
}

func (l *Linker) getCounterPlaces() int {
	return l.counterPlaces
}

func (l *Linker) setCounterPlaces(c int) BuildLink {
	l.counterPlaces = c
	return l
}

func (l *Linker) getNameTransition() string {
	return l.nameTransition
}

func (l *Linker) setNameTransition(n string) BuildLink {
	l.nameTransition = n
	return l
}

func (l *Linker) getCounterTransitions() int {
	return l.counterTransitions
}

func (l *Linker) setCounterTransitions(c int) BuildLink {
	l.counterTransitions = c
	return l
}

func (l *Linker) initNext() BuildLink {
	l.next = 0
	return l
}

func (l *Linker) isInfo() bool {
	return l.info
}

func (l *Linker) setInfo(i bool) BuildLink {
	l.info = i
	return l
}

func (l *Linker) printInfo() {
	fmt.Printf("%s\n%+v\n",
		strings.Repeat("=", 10), l,
	)
}

func (l *Linker) printParams() {
	fmt.Println("Test")
}

func (l *Linker) clone() BuildLink {
	var n Linker
	n = *l
	return &n
}
