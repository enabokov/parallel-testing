package petri

import "fmt"

type globalCounter struct {
	link       int
	place      int
	transition int
	simulator  int
}

type globalTime struct {
	currentTime float64
	modTime     float64
}

func CreateNetGenerator(timeModeling float64, timeGen float64, distribution string) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker
	var counter *globalCounter

	places = append(places,
		Place{}.build("P0", 1, counter),
		Place{}.build("P1", 0, counter),
	)

	transitions = append(transitions, Transition{}.build("coming", timeGen, timeModeling, counter))
	transitions[0].setDistribution(distribution, transitions[0].timeServing)

	linksIn = append(linksIn, Linker{}.build(places[0], transitions[0], 1, false, counter))
	linksOut = append(linksOut,
		Linker{}.build(places[0], transitions[0], 1, false, counter),
		Linker{}.build(places[1], transitions[0], 1, false, counter),
	)

	return Net{}.build("Generator supplying requirement for serving", places, transitions, linksIn, linksOut)
}

func CreateNetSMO(timeModeling float64, numDevices int, timeServing float64, distribution string) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker
	var counter globalCounter

	places = append(places,
		Place{}.build("pending requirements", 0, &counter),
		Place{}.build("free devices", float64(numDevices), &counter),
		Place{}.build("served", 0, &counter),
	)

	transitions = append(transitions, Transition{}.build("serving", timeServing, timeModeling, &counter))
	transitions[0].setDistribution(distribution, transitions[0].timeServing)

	linksIn = append(linksIn,
		Linker{}.build(places[0], transitions[0], 1, false, &counter),
		Linker{}.build(places[1], transitions[0], 1, false, &counter),
	)

	linksOut = append(linksOut,
		Linker{}.build(places[1], transitions[0], 1, false, &counter),
		Linker{}.build(places[2], transitions[0], 1, false, &counter),
	)

	return Net{}.build("SMO with unlimited queue:", places, transitions, linksIn, linksOut)
}

func CreateNetFork(timeModeling float64, numberWay int, probabilities []float64) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker
	var counter globalCounter

	places = append(places, Place{}.build("P0", 0, &counter))
	for i := 0; i < numberWay; i++ {
		places = append(places, Place{}.build(fmt.Sprintf("P%d", i+1), 0, &counter))
	}

	for i := 0; i < numberWay; i++ {
		transitions = append(transitions, Transition{}.build(fmt.Sprintf("choice route %d", i+1), 0, timeModeling, &counter))
	}

	for i, transition := range transitions {
		transition.setProbability(probabilities[i])
	}

	for i := 0; i < numberWay; i++ {
		linksIn = append(linksIn, Linker{}.build(places[0], transitions[i], 1, false, &counter))
	}

	for i := 0; i < numberWay; i++ {
		linksOut = append(linksOut, Linker{}.build(places[i+1], transitions[i], 1, false, &counter))
	}

	return Net{}.build("branching route ", places, transitions, linksIn, linksOut)
}
