package parallel_testing

import (
	"fmt"
	"github.com/enabokov/parallel-testing/petri"
	"log"
	"sync"
	"testing"
)

func TestParallel(t *testing.T) {
	var c petri.GlobalCounter
	var gtime petri.GlobalTime
	var cond petri.GlobalLocker

	//cond.Cond = sync.NewCond()

	time := 100000.0
	numObjects := 8

	// sequence of 10 SMO groups and generator
	model := GetModelSMOGroupForTestParallel(numObjects, 10, &c, &gtime, &cond)
	log.Printf("Quantity of objects %d, quantity of positions in object %d", len(model.Objects), len(model.Objects[1].Places))
	model.TimeMod = time
	gtime.ModTime = time

	model.IsProtocolPrint = true

	var wg sync.WaitGroup
	fmt.Println("START RUNNING")
	log.Printf("Total %d\n", len(model.Objects))
	for i := 0; i < len(model.Objects); i++ {
		wg.Add(1)
		tmp := model.Objects[i]
		go func() {
			defer wg.Done()
			tmp.Run()
		}()
	}
	fmt.Println("Waiting for goroutines")
	wg.Wait()
	fmt.Println("DONE")

	PrintResultsForAllObjects(model)

	//e := model.Objects[numObjects - 1]
	//mean := 0.0
	//n := 0
	//frequency := 1

	//System.out.println("  quantity of Objects   "+model.getListObj().size()+
	//	", quantity of positions in one object "+model.getListObj().get(1).getListP().length);
	t.Log(model)
}
