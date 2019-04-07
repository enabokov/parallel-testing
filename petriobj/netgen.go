package petriobj

import "fmt"

func createNetGenerator(timeModeling float64, timeGen float64, distribution string) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker

	places = append(places,
		Place{}.build("P0", 1),
		Place{}.build("P1", 0),
	)

	transitions = append(transitions, Transition{}.build("coming", timeGen, timeModeling))
	transitions[0].setDistribution(distribution, transitions[0].timeServing)

	linksIn = append(linksIn, Linker{}.build(places[0], transitions[0], 1, false))
	linksOut = append(linksOut,
		Linker{}.build(places[0], transitions[0], 1, false),
		Linker{}.build(places[1], transitions[0], 1, false),
	)

	return Net{}.build("Generator supplying requirement for serving", places, transitions, linksIn, linksOut)
}

func createNetSMO(timeModeling float64, numDevices int, timeServing float64, distribution string) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker

	places = append(places,
		Place{}.build("pending requirements", 0),
		Place{}.build("free devices", float64(numDevices)),
		Place{}.build("served", 0),
	)

	transitions = append(transitions, Transition{}.build("serving", timeServing, timeModeling))
	transitions[0].setDistribution(distribution, transitions[0].timeServing)

	linksIn = append(linksIn,
		Linker{}.build(places[0], transitions[0], 1, false),
		Linker{}.build(places[1], transitions[0], 1, false),
	)

	linksOut = append(linksOut,
		Linker{}.build(places[1], transitions[0], 1, false),
		Linker{}.build(places[2], transitions[0], 1, false),
	)

	return Net{}.build("SMO with unlimited queue:", places, transitions, linksIn, linksOut)
}

func createNetFork(timeModeling float64, numberWay int, probabilities []float64) Net {
	var places []Place
	var transitions []Transition
	var linksIn []Linker
	var linksOut []Linker

	places = append(places, Place{}.build("P0", 0))
	for i := 0; i < numberWay; i++ {
		places = append(places, Place{}.build(fmt.Sprintf("P%d", i+1), 0))
	}

	for i := 0; i < numberWay; i++ {
		transitions = append(transitions, Transition{}.build(fmt.Sprintf("choice route %d", i+1), 0, timeModeling))
	}

	for i, transition := range transitions {
		transition.setProbability(probabilities[i])
	}

	for i := 0; i < numberWay; i++ {
		linksIn = append(linksIn, Linker{}.build(places[0], transitions[i], 1, false))
	}

	for i := 0; i < numberWay; i++ {
		linksOut = append(linksOut, Linker{}.build(places[i+1], transitions[i], 1, false))
	}

	return Net{}.build("branching route ", places, transitions, linksIn, linksOut)
}
