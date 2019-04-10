package petri

import "fmt"

type GlobalCounter struct {
	Link       int
	Place      int
	Transition int
	Simulator  int
}

type GlobalTime struct {
	CurrentTime float64
	ModTime     float64
}

func CreateNetGenerator(timeModeling float64, timeGen float64, distribution string) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker
	var counter *GlobalCounter

	places = append(places,
		Place{}.Build("P0", 1, counter),
		Place{}.Build("P1", 0, counter),
	)

	transitions = append(transitions, Transition{}.Build("coming", timeGen, timeModeling, counter))
	transitions[0].setDistribution(distribution, transitions[0].TimeServing)

	linksIn = append(linksIn, Linker{}.Build(places[0], transitions[0], 1, false, counter))
	linksOut = append(linksOut,
		Linker{}.Build(places[0], transitions[0], 1, false, counter),
		Linker{}.Build(places[1], transitions[0], 1, false, counter),
	)

	return Net{}.Build("Generator supplying requirement for serving", places, transitions, linksIn, linksOut)
}

func CreateNetSMO(timeModeling float64, numDevices int, timeServing float64, distribution string, c *GlobalCounter) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker

	places = append(places,
		Place{}.Build("pending requirements", 0, c),
		Place{}.Build("free devices", float64(numDevices), c),
		Place{}.Build("served", 0, c),
	)

	transitions = append(transitions, Transition{}.Build("serving", timeServing, timeModeling, c))
	transitions[0].setDistribution(distribution, transitions[0].TimeServing)

	linksIn = append(linksIn,
		Linker{}.Build(places[0], transitions[0], 1, false, c),
		Linker{}.Build(places[1], transitions[0], 1, false, c),
	)

	linksOut = append(linksOut,
		Linker{}.Build(places[1], transitions[0], 1, false, c),
		Linker{}.Build(places[2], transitions[0], 1, false, c),
	)

	return Net{}.Build("SMO with unlimited queue:", places, transitions, linksIn, linksOut)
}

func CreateNetFork(timeModeling float64, numberWay int, probabilities []float64) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker
	var counter GlobalCounter

	places = append(places, Place{}.Build("P0", 0, &counter))
	for i := 0; i < numberWay; i++ {
		places = append(places, Place{}.Build(fmt.Sprintf("P%d", i+1), 0, &counter))
	}

	for i := 0; i < numberWay; i++ {
		transitions = append(transitions, Transition{}.Build(fmt.Sprintf("choice route %d", i+1), 0, timeModeling, &counter))
	}

	for i, transition := range transitions {
		transition.setProbability(probabilities[i])
	}

	for i := 0; i < numberWay; i++ {
		linksIn = append(linksIn, Linker{}.Build(places[0], transitions[i], 1, false, &counter))
	}

	for i := 0; i < numberWay; i++ {
		linksOut = append(linksOut, Linker{}.Build(places[i+1], transitions[i], 1, false, &counter))
	}

	return Net{}.Build("branching route ", places, transitions, linksIn, linksOut)
}
