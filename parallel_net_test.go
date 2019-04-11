package parallel_testing

import (
	"github.com/enabokov/parallel-testing/petri"
	"log"
	"strings"
	"sync"
	"testing"
)

func TestParallelObj(t *testing.T) {
	var gtime petri.GlobalTime
	var c petri.GlobalCounter
	var cond petri.GlobalLocker

	//cond.Cond = sync.NewCond(&cond.Mux)

	numObj := 2
	model := GetModelSMOGroupForTestParallel(numObj, 10, &c, &gtime, &cond)
	timeModeling := 1000.0
	model.GoRun(timeModeling)

	var wg sync.WaitGroup
	log.Printf("Start %d goroutines\n", len(model.Objects))
	for _, e := range model.Objects {
		wg.Add(1)
		go func() {
			if strings.ToLower(e.Name) == "smowithoutqueue" {
				for i := 0; i < len(e.TNet.Places)/2; i++ {
					log.Printf("mean queue in SMO %d %f", e.NumObject-1, e.TNet.Places[2*i].GetMean())
				}
			}
			wg.Done()
		}()
	}

	log.Printf("Wait %d goroutines\n", len(model.Objects))
	wg.Wait()
}
