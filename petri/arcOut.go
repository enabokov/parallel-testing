package petri

import "log"

type ArcOut struct {
	NumP int
	NumT int
	K int
	NameP string
	NameT string
	Number int
}

func (a *ArcOut) Build(t Transition, p Place, k int, gnext *GNext) {
	a.NumP = p.getNumber()
	a.NumT = t.getNumber()
	a.K = k
	a.NameP = p.GetName()
	a.NameT = t.GetName()

	a.Number = gnext.NextArcOut
	gnext.NextArcOut++
}

func (a *ArcOut) getQuantity() int {
	return a.K
}

func (a *ArcOut) setQuantity(k int) {
	a.K = k
}

func (a *ArcOut) getNumP() int {
	return a.NumP
}

func (a *ArcOut) setNumP(n int) {
	a.NumP = n
}

func (a *ArcOut) getNumT() int {
	return a.NumT
}

func (a *ArcOut) setNumT(n int) {
	a.NumT = n
}

func (a *ArcOut) getNameT() string {
	return a.NameT
}

func (a *ArcOut) setNameT(s string) {
	a.NameT = s
}

func (a *ArcOut) getNameP() string {
	return a.NameP
}

func (a *ArcOut) setNameP(s string) {
	a.NameP = s
}

func (a *ArcOut) print() {
	if a.NameP != "" && a.NameT != "" {
		log.Printf("P: %s, T: %s, k: %d\n", a.NameP, a.NameT, a.getQuantity())
	} else {
		log.Printf("P: P%d, T: T%d, k: %d\n", a.NumP, a.NumT, a.getQuantity())
	}
}

func (a *ArcOut) printParameters() {
	log.Printf("This tie has direction from  place with number %d to transition with number %d and has %d value of multiplicity.\n", a.NumP, a.NumT, a.K)
}
