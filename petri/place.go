package petri

import "log"

type Place struct {
	Mark int
	Name string
	Number int
	Mean float64
	Next *GNext
	ObservedMax int
	ObservedMin int

	// position is external if it's shared for other Petri-object
	External bool
}

func (p *Place) Build(n string, m int, gnext *GNext) {
	p.Name = n
	p.Mark = m
	p.Mean = 0

	// golang has no static vars
	// thus each object keeps pointer on global next
	p.Next = gnext
	p.Number = gnext.NextPlace
	gnext.NextPlace++

	p.ObservedMax = m
	p.ObservedMin = m
}

func (p *Place) InitNext() {
	p.Next.NextPlace = 0
}

func (p *Place) changeMean(a float64) {
	p.Mean += (float64(p.Mark) - p.Mean) * a
}

func (p *Place) getMean() float64 {
	return p.Mean
}

func (p *Place) IncreaseMark(a int) {
	p.Mark += a
	if p.ObservedMax < p.Mark {
		p.ObservedMax = p.Mark
	}

	if p.ObservedMin > p.Mark {
		p.ObservedMin = p.Mark
	}
}

func (p *Place) decreaseMark(a int) {
	p.Mark -= a
	if p.ObservedMax < p.Mark {
		p.ObservedMax = p.Mark
	}

	if p.ObservedMin > p.Mark {
		p.ObservedMin = p.Mark
	}
}

func (p *Place) getMark() int {
	return p.Mark
}

func (p *Place) getObservedMax() int {
	return p.ObservedMax
}

func (p *Place) getObservedMin() int {
	return p.ObservedMin
}

func (p *Place) setMark(a int) {
	p.Mark = a
}

func (p *Place) GetName() string {
	return p.Name
}

func (p *Place) setName(s string) {
	p.Name = s
}

func (p *Place) getNumber() int {
	return p.Number
}

func (p *Place) setNumber(n int) {
	p.Number = n
}

func (p *Place) printParameters() {
	log.Printf("Place %s has such parameters: number %d, mark %d", p.Name, p.Number, p.Mark)
}

func (p *Place) isExternal() bool {
	return p.External
}

func (p *Place) setExternal(external bool) {
	p.External = external
}
