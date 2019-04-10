package petri

import "fmt"

type Place struct {
	Mark   float64
	Name   string
	Number int
	Mean   float64

	ObservedMax float64
	ObservedMin float64

	External bool
}

type BuildPlace interface {
	Build(name string, mark float64, c *GlobalCounter) Place

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

	initNext(*GlobalCounter) BuildPlace
	isExternal() bool
	setExternal(bool) BuildPlace

	print()

	clone() BuildPlace
}

func (p *Place) Build(name string, mark float64, c *GlobalCounter) Place {
	p.Name = name
	p.Mark = mark
	p.Mean = 0
	p.Number = c.Place
	c.Place++
	p.ObservedMax = mark
	p.ObservedMin = mark
	return *p
}

func (p *Place) getMean() float64 {
	return p.Mean
}

func (p *Place) setMean(m float64) BuildPlace {
	p.Mean = p.Mean + (p.Mark-p.Mean)*m
	return p
}

func (p *Place) getMark() float64 {
	return p.Mark
}

func (p *Place) setMark(m float64) BuildPlace {
	p.Mark = m
	return p
}

func (p *Place) incrMark(m float64) {
	p.Mark += m
	if p.ObservedMax < p.Mark {
		p.ObservedMax = p.Mark
	}

	if p.ObservedMin > p.Mark {
		p.ObservedMin = p.Mark
	}
}

func (p *Place) decrMark(m float64) {
	p.Mark -= m
	if p.ObservedMax < p.Mark {
		p.ObservedMax = p.Mark
	}

	if p.ObservedMin > p.Mark {
		p.ObservedMin = p.Mark
	}
}

func (p *Place) getObservedMax() float64 {
	return p.ObservedMax
}

func (p *Place) getObservedMin() float64 {
	return p.ObservedMin
}

func (p *Place) getName() string {
	return p.Name
}

func (p *Place) setName(n string) BuildPlace {
	p.Name = n
	return p
}

func (p *Place) getNumber() int {
	return p.Number
}

func (p *Place) setNumber(n int) BuildPlace {
	p.Number = n
	return p
}

func (p *Place) initNext(c *GlobalCounter) BuildPlace {
	c.Place = 0
	return p
}

func (p *Place) isExternal() bool {
	return p.External
}

func (p *Place) setExternal(e bool) BuildPlace {
	p.External = e
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
