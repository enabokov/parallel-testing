package petri

import "fmt"

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
	findPlaceByName(string) int
	getCurrentMark(string) float64
	getMeanMark(string) float64
	getCurrentBuffer(string) int
	getMeanBuffer(string) float64

	printLinks()
	PrintMark()
	printBuffer()

	clone() BuildNet
}

func (n *Net) Build(name string, places []Place, transitions []Transition, linksIn []Linker, linksOut []Linker) Net {
	n.Name = name
	n.CounterPlace = len(places)
	n.CounterTransition = len(transitions)
	n.CounterIn = len(linksIn)
	n.CounterOut = len(LinksOut)

	n.Places = places[:]
	n.Transitions = transitions[:]
	n.LinksIn = linksIn[:]
	n.LinksOut = LinksOut[:]

	for _, t := range n.Transitions {
		t.createInPlaces(places, linksIn)
		t.createOutPlaces(places, LinksOut)
	}

	return *n
}

func (n *Net) findPlaceByName(placeName string) int {
	for i, place := range n.Places {
		if placeName == place.getName() {
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

func (n *Net) getCurrentMark(placeName string) float64 {
	return n.Places[n.findPlaceByName(placeName)].getMark()
}

func (n *Net) getMeanMark(placeName string) float64 {
	return n.Places[n.findPlaceByName(placeName)].getMean()
}

func (n *Net) getCurrentBuffer(transitionName string) int {
	return n.Transitions[n.findTransitionByName(transitionName)].Buffer
}

func (n *Net) getMeanBuffer(transitionName string) float64 {
	return n.Transitions[n.findTransitionByName(transitionName)].Mean
}

func (n *Net) printLinks() {
	fmt.Printf("Petri net %s ties: %b Input links and %b Output links\n", n.Name, len(n.LinksIn), len(n.LinksOut))
	for _, l := range n.LinksIn {
		l.printInfo()
	}

	for _, l := range n.LinksOut {
		l.printInfo()
	}
}

func (n *Net) PrintMark() {
	fmt.Printf("Mark in Net %s: ", n.Name)
	for _, p := range n.Places {
		fmt.Print(p.getMean())
	}
	fmt.Println()
}

func (n *Net) printBuffer() {
	fmt.Printf("Buffer in Net %s: ", n.Name)
	for _, t := range n.Transitions {
		fmt.Print(t.Buffer)
	}
	fmt.Println()
}

func (n *Net) clone() BuildNet {
	var v Net
	v = *n
	v.Places = n.Places[:]
	v.Transitions = n.Transitions[:]
	v.LinksIn = n.LinksIn[:]
	v.LinksOut = n.LinksOut[:]
	return &v
}
