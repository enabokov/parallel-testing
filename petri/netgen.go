package petri

import "fmt"

func CreateNetGenerator(timeMean float64, gnext *GNext, gtimeModeling *GTimeModeling) *Net {
	var places []Place
	var transitions []Transition
	var arcsIn []ArcIn
	var arcsOut []ArcOut

	p1 := Place{}
	p2 := Place{}
	p1.Build("P1", 1, gnext)
	p2.Build("P2", 0, gnext)
	places = append(places, p1, p2)

	t1 := Transition{}
	t1.Build("T1", timeMean, gnext, gtimeModeling)
	transitions = append(transitions, t1)

	transitions[0].setDistribution("exp", transitions[0].getTimeServing())
	transitions[0].setParamDeviation(0.0)

	lin1 := ArcIn{}
	lin1.BuildWithK(p1, t1, 1, gnext)
	arcsIn = append(arcsIn, lin1)

	arcOut1 := ArcOut{}
	arcOut2 := ArcOut{}
	arcOut1.Build(t1, p2, 1, gnext)
	arcOut2.Build(t1, p1, 1, gnext)
	arcsOut = append(arcsOut, arcOut1, arcOut2)

	net := Net{}
	net.Build("Generator", places, transitions, arcsIn, arcsOut)

	gnext = &GNext{}

	return &net
}

func CreateNetSMOgroup(numInGroup int, numChannel int, timeMean float64, name string, gnext *GNext, gtimeModeling *GTimeModeling) *Net {
	var places []Place
	var transitions []Transition
	var arcsIn []ArcIn
	var arcsOut []ArcOut

	p1 := Place{}
	p1.Build("P0", 0, gnext)
	places = append(places, p1)
	for i := 0; i < numInGroup; i++ {
		p2 := Place{}
		p2.Build(fmt.Sprintf("P%d", 2 * i + 1), numChannel, gnext)
		p3 := Place{}
		p3.Build(fmt.Sprintf("P%d", 2 *i + 2), 0, gnext)
		places = append(places, p2, p3)

		t1 := Transition{}
		t1.Build(fmt.Sprintf("T%d", i), timeMean, gnext, gtimeModeling)
		transitions = append(transitions, t1)
		transitions[i].setDistribution("exp", transitions[i].getTimeServing())
		transitions[i].setParamDeviation(0.0)

		arcIn1 := ArcIn{}
		arcIn2 := ArcIn{}
		arcIn1.BuildWithK(places[2 * i], transitions[i], 1, gnext)
		arcIn2.BuildWithK(places[2 * i + 1], transitions[i], 1, gnext)
		arcsIn = append(arcsIn, arcIn1, arcIn2)

		arcOut1 := ArcOut{}
		arcOut2 := ArcOut{}
		arcOut1.Build(transitions[i], places[2 *i + 1], 1, gnext)
		arcOut2.Build(transitions[i], places[2 *i + 1], 1, gnext)
		arcsOut = append(arcsOut, arcOut1, arcOut2)
	}

	net := Net{}
	net.Build(name, places, transitions, arcsIn, arcsOut)

	gnext = &GNext{}

	return &net
}

//func CreateNetSMO(timeModeling float64, numDevices int, timeServing float64, distribution string) Net {
//	var places []Place
//	var transitions []Transition
//	var arcsIn []ArcIn
//	var arcsOut []ArcOut
//	var counter *GNext
//
//	p1 := Place{}
//	p1.Build("")
//
//	places = append(places,
//		Place{}.build("pending requirements", 0, &counter),
//		Place{}.build("free devices", float64(numDevices), &counter),
//		Place{}.build("served", 0, &counter),
//	)
//
//	transitions = append(transitions, Transition{}.build("serving", timeServing, timeModeling, &counter))
//	transitions[0].setDistribution(distribution, transitions[0].timeServing)
//
//	linksIn = append(linksIn,
//		Linker{}.build(places[0], transitions[0], 1, false, &counter),
//		Linker{}.build(places[1], transitions[0], 1, false, &counter),
//	)
//
//	linksOut = append(linksOut,
//		Linker{}.build(places[1], transitions[0], 1, false, &counter),
//		Linker{}.build(places[2], transitions[0], 1, false, &counter),
//	)
//
//	return Net{}.build("SMO with unlimited queue:", places, transitions, linksIn, linksOut)
//}
//
//func CreateNetFork(timeModeling float64, numberWay int, probabilities []float64) Net {
//	var places []Place
//	var transitions []Transition
//	var linksIn []Linker
//	var linksOut []Linker
//	var counter globalCounter
//
//	places = append(places, Place{}.build("P0", 0, &counter))
//	for i := 0; i < numberWay; i++ {
//		places = append(places, Place{}.build(fmt.Sprintf("P%d", i+1), 0, &counter))
//	}
//
//	for i := 0; i < numberWay; i++ {
//		transitions = append(transitions, Transition{}.build(fmt.Sprintf("choice route %d", i+1), 0, timeModeling, &counter))
//	}
//
//	for i, transition := range transitions {
//		transition.setProbability(probabilities[i])
//	}
//
//	for i := 0; i < numberWay; i++ {
//		linksIn = append(linksIn, Linker{}.build(places[0], transitions[i], 1, false, &counter))
//	}
//
//	for i := 0; i < numberWay; i++ {
//		linksOut = append(linksOut, Linker{}.build(places[i+1], transitions[i], 1, false, &counter))
//	}
//
//	return Net{}.build("branching route ", places, transitions, linksIn, linksOut)
//}
