package main

import (
	"fmt"
	"github.com/enabokov/parallel-testing/petri"
	"log"
	"math"
	"sync"
)

func main() {
	var time = 100000.0
	var numObj = 8

	var gnext *petri.GNext
	var gtimeMod *petri.GTimeModeling
	var glimit *petri.GLimitArrayExtInputs
	var gcurrent *petri.GTimeCurrent

	lock := sync.Mutex{}
	gcond := sync.NewCond(&lock)

	gnext = &petri.GNext{}
	gtimeMod = &petri.GTimeModeling{}
	glimit = &petri.GLimitArrayExtInputs{}
	gcurrent = &petri.GTimeCurrent{}

	gnext.NextSimulator = 1
	gtimeMod.TimeModelingSimulation = math.MaxFloat64 - 1
	gtimeMod.TimeModelingTransition = math.MaxFloat64 - 1
	glimit.LimitSimulation = 10
	gcurrent.TimeCurrentSimulation = 0

	model := getModelSMOgroupForTestParallel(numObj, 10, gnext, gtimeMod, glimit, gcurrent, gcond)
	log.Printf("Quantity of objects %d, quantity of positions in one object %d", len(model.GetListObj()), len(model.GetListObj()[1].GetListP()))
	model.SetTimeMod(time)

	// define global time mod and limit for all simulators at once
	model.ListObj[0].SetTimeMod(time)
	model.ListObj[0].SetLimitArrayExtInputs(2)

	var wg sync.WaitGroup
	for i := 0; i < len(model.GetListObj()); i++ {
		wg.Add(1)
		toRun := model.GetListObj()[i]
		go func() {
			defer wg.Done()
			toRun.Run()
		}()
	}

	log.Println("Wait for goroutines")
	wg.Wait()


	log.Println("Done")
}

func getModelSMOgroupForTestParallel(
	numGroups int,
	numInGroup int,
	gnext *petri.GNext,
	gtimeMod *petri.GTimeModeling,
	glimit *petri.GLimitArrayExtInputs,
	gcurrent *petri.GTimeCurrent, gcond *sync.Cond) *petri.Model {

	var list []petri.Simulator

	gtimeMod = &petri.GTimeModeling{}
	numSMO := numGroups - 1

	net := petri.CreateNetGenerator(2.0, gnext, gtimeMod)
	sim1 := petri.Simulator{}
	list = append(list, sim1.Build(net, gnext, gtimeMod, glimit, gcurrent))
	for i := 0; i < numSMO; i++ {
		sim2 := petri.Simulator{}
		list = append(list, sim2.Build(
			petri.CreateNetSMOgroup(numInGroup, 1, 1.0, fmt.Sprintf("group_%d", i), gnext, gtimeMod),
			gnext, gtimeMod, glimit, gcurrent))
	}

	list[0].GetNet().GetListP()[1] = list[1].GetNet().GetListP()[0]
	list[0].AddOutT(list[0].GetNet().GetListT()[0])
	list[1].AddInT(list[1].GetNet().GetListT()[0])
	list[0].NextObj = &list[1]
	list[1].PrevObj = &list[0]

	if numSMO > 1 {
		for i := 2; i <= numSMO; i++ {
			last := len(list[i-1].GetNet().GetListP()) - 1
			list[i].GetNet().GetListP()[0] = list[i-1].GetNet().GetListP()[last]

			lastT := len(list[i-1].GetNet().GetListT()) - 1
			list[i-1].AddOutT(list[i-1].GetNet().GetListT()[lastT])
			list[i].AddInT(list[i].GetNet().GetListT()[0])
			list[i-1].NextObj = &list[i]
			list[i].PrevObj = &list[i-1]
		}
	}

	for i := 1; i <= numSMO; i++ {
		var positionForStatistics []petri.Place
		listP := list[i].GetNet().GetListP()
		for j := 0; j < len(listP) -1; j++ {
			positionForStatistics = append(positionForStatistics, listP[j])
		}

		list[i].ListPositionsForStatistica = positionForStatistics
	}

	model := petri.Model{}
	model.Build(list)

	return &model
}