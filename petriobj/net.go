package petriobj

import "fmt"

type Net struct {
	name string
	counterPlace int
	counterTransition int
	counterIn int
	counterOut int

	places []Place
	transitions []Transition
	linksIn []Linker
	linksOut []Linker
}

type BuildNet interface {
	findPlaceByName(string) int
	getCurrentMark(string) float64
	getMeanMark(string) float64
	getCurrentBuffer(string) int
	getMeanBuffer(string) float64

	printLinks()
	printMark()
	printBuffer()

	clone() BuildNet
}

func (n *Net) findPlaceByName(placeName string) int {
	for i, place := range n.places {
		if placeName == place.getName() {
			return i
		}
	}

	return -1
}

func (n *Net) findTransitionByName(transitionName string) int {
	for i, transition := range n.transitions {
		if transitionName == transition.name {
			return i
		}
	}

	return -1
}

func (n *Net) getCurrentMark(placeName string) float64 {
	return n.places[n.findPlaceByName(placeName)].getMark()
}

func (n *Net) getMeanMark(placeName string) float64 {
	return n.places[n.findPlaceByName(placeName)].getMean()
}

func (n *Net) getCurrentBuffer(transitionName string) int {
	return n.transitions[n.findTransitionByName(transitionName)].buffer
}

func (n *Net) getMeanBuffer(transitionName string) float64 {
	return n.transitions[n.findTransitionByName(transitionName)].mean
}

func (n *Net) printLinks() {
	fmt.Printf("Petri net %s ties: %b input links and %b output links\n", n.name, len(n.linksIn), len(n.linksOut))
	for _, l := range n.linksIn {
		l.printInfo()
	}

	for _, l := range n.linksOut {
		l.printInfo()
	}
}

func (n *Net) printMark() {
	fmt.Printf("Mark in Net %s: ", n.name)
	for _, p := range n.places {
		fmt.Print(p.getMean())
	}
	fmt.Println()
}

func (n *Net) printBuffer() {
	fmt.Printf("Buffer in Net %s: ", n.name)
	for _, t := range n.transitions {
		fmt.Print(t.buffer)
	}
	fmt.Println()
}

func (n *Net) clone() BuildNet {
	var v Net
	v = *n
	v.places = n.places[:]
	v.transitions = n.transitions[:]
	v.linksIn = n.linksIn[:]
	v.linksOut = n.linksOut[:]
	return &v
}
