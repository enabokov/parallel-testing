package petri

import (
	"log"
)

type Place struct {
	Counter *GlobalCounter
	Mark    float64
	Name    string
	Number  int
	Mean    float64

	ObservedMax float64
	ObservedMin float64

	External bool
}

type BuildPlace interface {
	Build(name string, mark float64, c *GlobalCounter) *Place

	GetMean() float64
	SetMean(float64) BuildPlace // synchronized

	GetMark() float64           // synchronized
	SetMark(float64) BuildPlace // synchronized
	IncrMark(float64)           // synchronized
	DecrMark(float64)           // synchronized

	GetObservedMax() float64
	GetObservedMin() float64

	GetName() string
	SetName(string) BuildPlace

	GetNumber() int
	SetNumber(int) BuildPlace

	InitNext(*GlobalCounter) BuildPlace
	IsExternal() bool
	SetExternal(bool) BuildPlace

	Print()

	Clone() BuildPlace
}

func (p *Place) Build(name string, mark float64, c *GlobalCounter) *Place {
	p.Name = name
	p.Mark = mark
	p.Mean = 0
	p.Counter = c
	p.initNumber()
	p.incr()
	p.Number = p.Counter.Place
	p.ObservedMax = mark
	p.ObservedMin = mark
	return p
}

func (p *Place) initNumber() {
	p.Counter.Lock()
	p.Number = p.Counter.Place
	p.Counter.Unlock()
}

func (p *Place) incr() {
	p.Counter.Lock()
	p.Counter.Place++
	p.Counter.Unlock()
}

func (p *Place) GetMean() float64 {
	return p.Mean
}

func (p *Place) SetMean(m float64) BuildPlace {
	p.Mean += (p.Mark - p.Mean) * m
	return p
}

func (p *Place) GetMark() float64 {
	return p.Mark
}

func (p *Place) SetMark(m float64) BuildPlace {
	p.Mark = m
	return p
}

func (p *Place) IncrMark(m float64) {
	p.Mark += m
	if p.ObservedMax < p.Mark {
		p.ObservedMax = p.Mark
	}

	if p.ObservedMin > p.Mark {
		p.ObservedMin = p.Mark
	}
}

func (p *Place) DecrMark(m float64) {
	p.Mark -= m
	if p.ObservedMax < p.Mark {
		p.ObservedMax = p.Mark
	}

	if p.ObservedMin > p.Mark {
		p.ObservedMin = p.Mark
	}
}

func (p *Place) GetObservedMax() float64 {
	return p.ObservedMax
}

func (p *Place) GetObservedMin() float64 {
	return p.ObservedMin
}

func (p *Place) GetName() string {
	return p.Name
}

func (p *Place) SetName(n string) BuildPlace {
	p.Name = n
	return p
}

func (p *Place) GetNumber() int {
	return p.Number
}

func (p *Place) SetNumber(n int) BuildPlace {
	p.Number = n
	return p
}

func (p *Place) InitNext(c *GlobalCounter) BuildPlace {
	c.Place = 0
	return p
}

func (p *Place) IsExternal() bool {
	return p.External
}

func (p *Place) SetExternal(e bool) BuildPlace {
	p.External = e
	return p
}

func (p *Place) Print() {
	log.Printf("Place %s has such params:\n number: %d, mark: %f\n", p.Name, p.Number, p.Mark)
}

func (p *Place) Clone() BuildPlace {
	var n Place
	n = *p
	return &n
}
