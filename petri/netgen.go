package petri

import (
	"fmt"
)

func CreateNetGenerator(timeModeling float64, timeGen float64, distribution string, counter *GlobalCounter) Net {
	var places []*Place
	var transitions []*Transition
	var linksIn []*Linker
	var linksOut []*Linker

	places = append(places,
		(&Place{}).Build("P0", 1, counter),
		(&Place{}).Build("P1", 0, counter),
	)

	transitions = append(transitions,
		(&Transition{}).Build("coming", timeGen, timeModeling, counter),
	)
	transitions[0].SetDistribution(distribution, transitions[0].TimeServing)

	linksIn = append(linksIn,
		(&Linker{}).Build(places[0], transitions[0], 1, false, counter, `i`),
	)

	linksOut = append(linksOut,
		(&Linker{}).Build(places[0], transitions[0], 1, false, counter, `o`),
		(&Linker{}).Build(places[1], transitions[0], 1, false, counter, `o`),
	)

	net := (&Net{}).Build("Generator supplying requirement for serving", places, transitions, linksIn, linksOut)
	counter = &GlobalCounter{}
	return net
}

func CreateNetSMOGroup(numInGroup float64, numChannel int, timeMean float64, name string, c *GlobalCounter) Net {
	var places []*Place
	var transitions []*Transition
	var linksIn []*Linker
	var linksOut []*Linker

	places = append(places,
		(&Place{}).Build("P0", 0, c),
	)

	for i := 0; i < int(numInGroup); i++ {
		places = append(places,
			(&Place{}).Build(fmt.Sprintf("P%d", 2*i+1), float64(numChannel), c),
			(&Place{}).Build(fmt.Sprintf("P%d", 2*i+2), 0, c),
		)

		transitions = append(transitions,
			(&Transition{}).Build(fmt.Sprintf("T%d", i), timeMean, 1, c),
		)

		transitions[i].SetDistribution("exp", transitions[i].TimeServing)
		transitions[i].SetDeviation(0.0)

		linksIn = append(linksIn,
			(&Linker{}).Build(places[2*i], transitions[i], 1, false, c, `i`),
			(&Linker{}).Build(places[2*i+1], transitions[i], 1, false, c, `i`),
		)

		linksOut = append(linksOut,
			(&Linker{}).Build(places[2*i+1], transitions[i], 1, false, c, `o`),
			(&Linker{}).Build(places[2*i+2], transitions[i], 1, false, c, `o`),
		)
	}

	net := (&Net{}).Build(name, places, transitions, linksIn, linksOut)

	// start counter over
	c = &GlobalCounter{}

	return net
}

func CreateNetFork(timeModeling float64, numberWay int, probabilities []float64) Net {
	var places []*Place
	var transitions []*Transition
	var linksIn []*Linker
	var linksOut []*Linker
	var counter GlobalCounter

	p1 := Place{}
	places = append(places, p1.Build("P0", 0, &counter))
	for i := 0; i < numberWay; i++ {
		pi := Place{}
		places = append(places, pi.Build(fmt.Sprintf("P%d", i+1), 0, &counter))
	}

	for i := 0; i < numberWay; i++ {
		ti := Transition{}
		transitions = append(transitions, ti.Build(fmt.Sprintf("choice route %d", i+1), 0, timeModeling, &counter))
	}

	for i := 0; i < len(transitions); i++ {
		transitions[i].SetProbability(probabilities[i])
	}

	for i := 0; i < numberWay; i++ {
		li := Linker{}
		linksIn = append(linksIn,
			li.Build(places[0], transitions[i], 1, false, &counter, `i`))
	}

	for i := 0; i < numberWay; i++ {
		li := Linker{}
		linksOut = append(linksOut,
			li.Build(places[i+1], transitions[i], 1, false, &counter, `o`))
	}

	return (&Net{}).Build("branching route ", places, transitions, linksIn, linksOut)
}
