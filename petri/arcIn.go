package petri

import "log"

type ArcIn struct {
	NumP int
	NumT int
	K int
	Inf bool
	NameP string
	NameT string
	Number int
}

func (a *ArcIn) Build(p Place, t Transition, gnext *GNext) {
	a.NumP = p.getNumber()
	a.NumT = t.getNumber()
	a.K = 1
	a.Inf = false
	a.NameP = p.GetName()
	a.NameT = t.GetName()

	a.Number = gnext.NextArcIn
	gnext.NextArcIn++
}

func (a *ArcIn) BuildWithK(p Place, t Transition, k int, gnext *GNext) {
	a.NumP = p.getNumber()
	a.NumT = t.getNumber()
	a.K = k
	a.Inf = false
	a.NameP = p.GetName()
	a.NameT = t.GetName()

	a.Number = gnext.NextArcIn
	gnext.NextArcIn++
}

func (a *ArcIn) BuildWithKAndInf(p Place, t Transition, k int, isInf bool, gnext *GNext) {
	a.NumP = p.getNumber()
	a.NumT = t.getNumber()
	a.K = k
	a.Inf = isInf
	a.NameP = p.GetName()
	a.NameT = t.GetName()

	a.Number = gnext.NextArcIn
	gnext.NextArcIn++
}

func (a *ArcIn) getQuantity() int {
	return a.K
}

func (a *ArcIn) setQuantity(k int) {
	a.K = k
}

func (a *ArcIn) getNumP() int {
	return a.NumP
}

func (a *ArcIn) setNumP(n int) {
	a.NumP = n
}

func (a *ArcIn) getNumT() int {
	return a.NumT
}

func (a *ArcIn) setNumT(n int) {
	a.NumT = n
}

func (a *ArcIn) getNameT() string {
	return a.NameT
}

func (a *ArcIn) setNameT(s string) {
	a.NameT = s
}

func (a *ArcIn) getNameP() string {
	return a.NameP
}

func (a *ArcIn) setNameP(s string) {
	a.NameP = s
}

func (a *ArcIn) getIsInf() bool {
	return a.Inf
}

func (a *ArcIn) setInf(i bool) {
	a.Inf = i
}

func (a *ArcIn) print() {
	if a.NameP != "" && a.NameT != "" {
		log.Printf("P: %s, T: %s, inf: %v, k: %d\n", a.NameP, a.NameT, a.getIsInf(), a.getQuantity())
	} else {
		log.Printf("P: P%d, T: T%d, inf: %v, k: %d\n", a.NumP, a.NumT, a.getIsInf(), a.getQuantity())
	}
}

func (a *ArcIn) printParameters() {
	log.Printf("This tie has direction from  place with number %d to transition with number %d and has %d value of multiplicity, \n", a.NumP, a.NumT, a.K)
	if a.Inf {
		log.Println(" and is informational")
	}
}
