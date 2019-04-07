package petriobj

import "fmt"

type Place struct {
	mark   float64
	name   string
	number int
	mean   float64

	observedMax float64
	observedMin float64

	next     int
	external bool
}

type BuildPlace interface {
	build(name string, mark float64) Place

	getMean() float64
	setMean(float64) BuildPlace // synchronized

	getMark() float64           // synchronized
	setMark(float64) BuildPlace // synchronized
	incrMark(float64)           // synchronized
	decrMark(float64)           // synchronized

	getObservedMax() float64
	getObservedMin() float64

	getName() string
	setName(string) BuildPlace

	getNumber() int
	setNumber(int) BuildPlace

	initNext() BuildPlace
	isExternal() bool
	setExternal(bool) BuildPlace

	print()

	clone() BuildPlace
}

func (p *Place) build(name string, mark float64) Place {
	p.name = name
	p.mark = mark
	p.mean = 0
	p.number = p.next
	p.next++
	p.observedMax = mark
	p.observedMin = mark
	return *p
}

func (p *Place) getMean() float64 {
	return p.mean
}

func (p *Place) setMean(m float64) BuildPlace {
	p.mean = p.mean + (p.mark-p.mean)*m
	return p
}

func (p *Place) getMark() float64 {
	return p.mark
}

func (p *Place) setMark(m float64) BuildPlace {
	p.mark = m
	return p
}

func (p *Place) incrMark(m float64) {
	p.mark += m
	if p.observedMax < p.mark {
		p.observedMax = p.mark
	}

	if p.observedMin > p.mark {
		p.observedMin = p.mark
	}
}

func (p *Place) decrMark(m float64) {
	p.mark -= m
	if p.observedMax < p.mark {
		p.observedMax = p.mark
	}

	if p.observedMin > p.mark {
		p.observedMin = p.mark
	}
}

func (p *Place) getObservedMax() float64 {
	return p.observedMax
}

func (p *Place) getObservedMin() float64 {
	return p.observedMin
}

func (p *Place) getName() string {
	return p.name
}

func (p *Place) setName(n string) BuildPlace {
	p.name = n
	return p
}

func (p *Place) getNumber() int {
	return p.number
}

func (p *Place) setNumber(n int) BuildPlace {
	p.number = n
	return p
}

func (p *Place) initNext() BuildPlace {
	p.next = 0
	return p
}

func (p *Place) isExternal() bool {
	return p.external
}

func (p *Place) setExternal(e bool) BuildPlace {
	p.external = e
	return p
}

func (p *Place) print() {
	fmt.Printf("%+v", p)
}

func (p *Place) clone() BuildPlace {
	var n Place
	n = *p
	return &n
}
