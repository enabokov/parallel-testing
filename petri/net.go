package petri

import (
	"log"
	"strings"
)

type Net struct {
	Name string
	NumP int
	NumT int
	NumIn int
	NumOut int
	ListP []Place
	ListT []Transition
	ListIn []ArcIn
	ListOut []ArcOut
}

func (n *Net) Build(s string, pp []Place, tt []Transition, in []ArcIn, out[]ArcOut) {
	n.Name = s
	n.NumP = len(pp)
	n.NumT = len(tt)
	n.NumIn = len(in)
	n.NumOut = len(out)
	n.ListP = pp
	n.ListT = tt
	n.ListIn = in
	n.ListOut = out

	for i := 0; i < n.NumP; i++ {
		n.ListP[i] = pp[i]
	}

	for i := 0; i < n.NumT; i++ {
		n.ListT[i] = tt[i]
	}

	for i := 0; i < n.NumIn; i++ {
		n.ListIn[i] = in[i]
	}

	for i := 0; i < n.NumOut; i++ {
		n.ListOut[i] = out[i]
	}

	for i := 0; i < len(n.ListT); i++ {
		n.ListT[i].createInP(n.ListP, n.ListIn)
		n.ListT[i].createOutP(n.ListP, n.ListOut)
	}
}

func (n *Net) GetName() string {
	return n.Name
}

func (n *Net) SetName(s string) {
	n.Name = s
}

func (n *Net) GetListP() []Place {
	return n.ListP
}

func (n *Net) GetListT() []Transition {
	return n.ListT
}

func (n *Net) GetArcIn() []ArcIn {
	return n.ListIn
}

func (n *Net) GetArcOut() []ArcOut {
	return n.ListOut
}

func (n *Net) StrToNumP(s string) int {
	a := -1

	for i := 0; i < len(n.ListP); i++ {
		if strings.ToLower(s) == n.ListP[i].GetName() {
			a = n.ListP[i].getNumber()
			break
		}
	}

	return a
}

func (n *Net) StrToNumT(s string) int {
	a := -1

	for i := 0; i < len(n.ListT); i++ {
		if strings.ToLower(s) == n.ListT[i].GetName() {
			a = n.ListT[i].getNumber()
			break
		}
	}

	return a
}

func (n *Net) GetCurrentMark(s string) int {
	return n.ListP[n.StrToNumP(s)].getMark()
}

func (n *Net) getMeanMark(s string) float64 {
	return n.ListP[n.StrToNumP(s)].getMean()
}

func (n *Net) getCurrentBuffer(s string) int {
	return n.ListT[n.StrToNumT(s)].getBuffer()
}

func (n *Net) getMeanBuffer(s string) float64 {
	return n.ListT[n.StrToNumT(s)].getMean()
}

func (n *Net) printTies() {
	log.Printf("Petri net %s ties: %d input ties and %d output ties\n", n.Name, len(n.ListIn), len(n.ListOut))
}

func (n *Net) printMark() {
	log.Printf("Mark in Net %s: ", n.GetName())
	for _, p := range n.ListP {
		log.Print(p.getMark())
	}
	log.Println()
}

func (n *Net) printBuffer() {
	log.Printf("Buffer in Net %s: ", n.GetName())
	for _, t := range n.ListT {
		log.Print(t.getBuffer())
	}
	log.Println()
}
