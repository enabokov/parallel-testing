package parallel_testing

import (
	"fmt"
	"github.com/enabokov/parallel-testing/petri"
	"log"
	"sync"
)

func GetModelSMOGroupForTestParallel(numGroups int, numInGroup int, c *petri.GlobalCounter, gtime *petri.GlobalTime, cond *petri.GlobalLocker) *petri.Model {
	fmt.Println("Creating model SMO for parallel testing")
	var list []*petri.Simulator
	var counter *petri.GlobalCounter

	counter = &petri.GlobalCounter{}

	numSMO := numGroups - 1
	list = append(list,
		(&petri.Simulator{}).Build(petri.CreateNetGenerator(2.0, 10, "norm", c),
			c, gtime, cond),
	)
	log.Printf("CREATED OBJECTS %+v\n", counter)
	for i := 0; i < numSMO; i++ {
		list = append(list,
			(&petri.Simulator{}).Build(
				petri.CreateNetSMOGroup(float64(numInGroup), 1, 1.0, fmt.Sprintf("group_%d", i), c),
				c, gtime, cond),
		)
	}

	list[0].TNet.Places[1] = list[1].TNet.Places[0]
	list[0].OutT = append(list[0].OutT, list[0].TNet.Transitions[0])
	list[1].InT = append(list[1].InT, list[1].TNet.Transitions[0])
	list[0].NextObj = list[1]
	list[1].PrevObj = list[0]

	if numSMO > 1 {
		for i := 2; i <= numSMO; i++ {
			last := len(list[i-1].TNet.Places) - 1

			//group1 = > group2, group2 = > group3,...
			list[i].TNet.Places[0] = list[i-1].TNet.Places[last]

			lastT := len(list[i-1].TNet.Transitions) - 1
			list[i-1].OutT = append(list[i-1].OutT, list[i-1].TNet.Transitions[lastT])
			list[i].InT = append(list[i].InT, list[i].TNet.Transitions[0])
			list[i-1].NextObj = list[i]
			list[i].PrevObj = list[i-1]
		}
	}

	for i := 1; i <= numSMO; i++ {
		var positionForStats []*petri.Place
		var listP []*petri.Place
		listP = list[i].TNet.Places
		for j := 0; j < len(listP)-1; j++ {
			positionForStats = append(positionForStats, listP[j])
		}

		list[i].StatisticsPlaces = positionForStats
	}

	return (&petri.Model{}).Build(list, gtime)
}

func PrintResultsForAllObjects(model *petri.Model) {
	var wg sync.WaitGroup
	for _, e := range model.Objects {
		wg.Add(1)
		go func() {
			log.Printf("For SMO %s: tLocal: %f\n", e.Name, e.TimeLocal)
			if e.PrevObj != nil {
				if len(e.TimeExternalInput) > 0 {
					log.Printf("tExternalInput first: %+v\ntExternalInput last: %+v\n", e.TimeExternalInput[0], e.TimeExternalInput[len(e.TimeExternalInput)-1])
				}

				for j := 0; j < len(e.TNet.Places)/2; j++ {
					log.Printf("Mean queue in SMO %s %f, mark in position %f", e.Name, e.TNet.Places[2*j].Mean, e.TNet.Places[2*j].Mark)
				}
			}

			wg.Done()
		}()
	}
}
