package petri

import (
	"fmt"
)

type Net struct {
	Name              string
	CounterPlace      int
	CounterTransition int
	CounterIn         int
	CounterOut        int

	Places      []Place
	Transitions []Transition
	LinksIn     []Linker
	LinksOut    []Linker
}

type BuildNet interface {
	Build(string, []Place, []Transition, []Linker, []Linker) Net
	FindPlaceByName(string) int
	GetCurrentMark(string) float64
	GetMeanMark(string) float64
	GetCurrentBuffer(string) int
	GetMeanBuffer(string) float64

	PrintLinks()
	PrintMark()
	PrintBuffer()

	Clone() BuildNet
}

func (n *Net) Build(name string, places []Place, transitions []Transition, linksIn []Linker, linksOut []Linker) Net {
	n.Name = name
	n.CounterPlace = len(places)
	n.CounterTransition = len(transitions)
	n.CounterIn = len(linksIn)
	n.CounterOut = len(linksOut)

	n.Places = places[:]
	n.Transitions = transitions[:]
	n.LinksIn = linksIn[:]
	n.LinksOut = linksOut[:]

	for _, t := range n.Transitions {
		t.CreateInPlaces(places, linksIn)
		t.CreateOutPlaces(places, linksOut)
	}

	return *n
}

func (n *Net) FindPlaceByName(placeName string) int {
	for i, place := range n.Places {
		if placeName == place.GetName() {
			return i
		}
	}

	return -1
}

func (n *Net) findTransitionByName(transitionName string) int {
	for i, transition := range n.Transitions {
		if transitionName == transition.Name {
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
	return n.Transitions[n.findTransitionByName(transitionName)].Buffer
}

func (n *Net) GetMeanBuffer(transitionName string) float64 {
	return n.Transitions[n.findTransitionByName(transitionName)].Mean
}

func (n *Net) PrintLinks() {
	fmt.Printf("Petri net %s ties: %b Input links and %b Output links\n", n.Name, len(n.LinksIn), len(n.LinksOut))
	for _, l := range n.LinksIn {
		l.PrintInfo()
	}

	for _, l := range n.LinksOut {
		l.PrintInfo()
	}
}

func (n *Net) PrintMark() {
	fmt.Printf("Mark in Net %s: ", n.Name)
	for _, p := range n.Places {
		fmt.Print(p.GetMean())
	}
	fmt.Println()
}

func (n *Net) PrintBuffer() {
	fmt.Printf("Buffer in Net %s: ", n.Name)
	for _, t := range n.Transitions {
		fmt.Print(t.Buffer)
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
