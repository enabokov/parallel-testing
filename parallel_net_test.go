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
	for i := 0; i < len(model.Objects); i++ {
		wg.Add(1)
		obj := model.Objects[i]
		go func() {
			if strings.ToLower(obj.Name) == "smowithoutqueue" {
				for i := 0; i < len(obj.TNet.Places)/2; i++ {
					log.Printf("mean queue in SMO %d %f", obj.NumObject-1, obj.TNet.Places[2*i].GetMean())
				}
			}
			wg.Done()
		}()
	}

	log.Printf("Wait %d goroutines\n", len(model.Objects))
	wg.Wait()
}
