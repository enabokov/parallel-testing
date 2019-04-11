package petri

import (
	"fmt"
	"log"
)

type Net struct {
	Name              string
	CounterPlace      int
	CounterTransition int
	CounterIn         int
	CounterOut        int

	Places      []*Place
	Transitions []*Transition
	LinksIn     []*Linker
	LinksOut    []*Linker
}

type BuildNet interface {
	Build(string, []*Place, []*Transition, []*Linker, []*Linker) Net
	FindPlaceByName(string) int
	FindTransitionByName(string) int
	GetCurrentMark(string) float64
	GetMeanMark(string) float64
	GetCurrentBuffer(string) int
	GetMeanBuffer(string) float64

	PrintLinks()
	PrintMark()
	PrintBuffer()

	Clone() BuildNet
}

func (n *Net) Build(name string, places []*Place, transitions []*Transition, linksIn []*Linker, linksOut []*Linker) Net {
	n.Name = name
	n.CounterPlace = len(places)
	n.CounterTransition = len(transitions)
	n.CounterIn = len(linksIn)
	n.CounterOut = len(linksOut)

	n.Places = places
	n.Transitions = transitions
	n.LinksIn = linksIn
	n.LinksOut = linksOut

	for i := 0; i < len(n.Transitions); i++ {
		n.Transitions[i].CreateInPlaces(places, linksIn)
		n.Transitions[i].CreateOutPlaces(places, linksOut)

		if len(n.Transitions[i].InPlaces) == 0 {
			log.Println(fmt.Errorf("error: Transition %s has empty list of input places", n.Transitions[i].Name))
		}

		if len(n.Transitions[i].OutPlaces) == 0 {
			log.Println(fmt.Errorf("error: Transition %s has empty list of output places", n.Transitions[i].Name))
		}
	}

	return *n
}

func (n *Net) FindPlaceByName(placeName string) int {
	for i := 0; i < len(n.Places); i++ {
		if placeName == n.Places[i].GetName() {
			return i
		}
	}

	return -1
}

func (n *Net) FindTransitionByName(transitionName string) int {
	for i := 0; i < len(n.Transitions); i++ {
		if transitionName == n.Transitions[i].Name {
			return i
		}
	}

	return -1
}

func (n *Net) GetCurrentMark(placeName string) float64 {
	return n.Places[n.FindPlaceByName(placeName)].GetMark()
}

func (n *Net) GetMeanMark(placeName string) float64 {
	return n.Places[n.FindPlaceByName(placeName)].GetMean()
}

func (n *Net) GetCurrentBuffer(transitionName string) int {
	return n.Transitions[n.FindTransitionByName(transitionName)].Buffer
}

func (n *Net) GetMeanBuffer(transitionName string) float64 {
	return n.Transitions[n.FindTransitionByName(transitionName)].Mean
}

func (n *Net) PrintLinks() {
	fmt.Printf("Petri net %s ties: %b Input links and %b Output links\n", n.Name, len(n.LinksIn), len(n.LinksOut))
	for i := 0; i < len(n.LinksIn); i++ {
		n.LinksIn[i].PrintInfo()
	}

	for i := 0; i < len(n.LinksOut); i++ {
		n.LinksOut[i].PrintInfo()
	}
}

func (n *Net) PrintMark() {
	fmt.Printf("Mark in Net %s: ", n.Name)
	for i := 0; i < len(n.Places); i++ {
		fmt.Print(n.Places[i].GetMean())
	}
	fmt.Println()
}

func (n *Net) PrintBuffer() {
	fmt.Printf("Buffer in Net %s: ", n.Name)
	for i := 0; i < len(n.Transitions); i++ {
		fmt.Print(n.Transitions[i].Buffer)
	}
	fmt.Println()
}

func (n *Net) Clone() BuildNet {
	var v Net
	v = *n
	v.Places = n.Places[:]
	v.Transitions = n.Transitions[:]
	v.LinksIn = n.LinksIn[:]
	v.LinksOut = n.LinksOut[:]
	return &v
}
