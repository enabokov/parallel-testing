package petri

import (
	"log"
	"math/rand"
	"sort"
	"sync"
)

type Model struct {
	gtime           *globalTime
	objects         []Simulator
	timeMod         float64
	t               float64
	isProtocolPrint bool
	isStatistics    bool
}

type BuildModel interface {
	getNextEventTime() float64
	modelInput()
	sortObj([]Simulator)
	chooseObj([]Simulator) Simulator
	parallelGo(float64)
	goRun(float64)
}

func (m *Model) getNextEventTime() float64 {
	min := m.objects[0].timeMin
	for _, m := range m.objects {
		if m.timeMin < min {
			min = m.timeMin
		}
	}

	return min
}

func (m *Model) modelInput() {
	m.sortObj(m.objects)
	var wg sync.WaitGroup

	for _, obj := range m.objects {
		wg.Add(1)
		go func() {
			obj.input()
			wg.Done()
		}()
	}

	wg.Wait()
}

func (m *Model) sortObj(s []Simulator) {
	sort.SliceStable(s, func(i, j int) bool {
		return s[i].priority < s[j].priority
	})
}

func (m *Model) chooseObj(s []Simulator) Simulator {
	var num int
	var max int

	if len(s) > 1 {
		max = len(s)
		m.sortObj(s)
		for i := 1; i < len(s); i++ {
			if s[i].priority < s[i-1].priority {
				max = i - 1
				break
			}
		}

		if max == 0 {
			num = 0
		} else {
			num = rand.Intn(max)
		}
	} else {
		num = 0
	}

	return s[num]
}

func (m *Model) parallelGo(timeModeling float64) {
	m.timeMod = timeModeling

	t := 0.0
	var min float64

	if m.isProtocolPrint {
		log.Println("Start marking objects:")
		// print marks
	}

	var conflictObj []Simulator
	for t < timeModeling {
		conflictObj = []Simulator{}

		// maybe conditions changed
		m.modelInput()
		if m.isProtocolPrint {
			log.Println("Enter markers into transitions")
			// print marks
		}

		// search the closest event
		min = m.getNextEventTime()
		if m.isStatistics {
			for _, obj := range m.objects {

				// statistics within delta t
				// statistics is collected only once for all common positions
				obj.doStatisticsWithInterval((min - t) / min)
			}
		}

		// pass time further
		t = min

		m.gtime.currentTime = t

		if m.isProtocolPrint {
			log.Printf("Passing time further. t: %f", t)
		}

		if t <= timeModeling {
			for _, e := range m.objects {
				if t == e.timeMin {
					conflictObj = append(conflictObj, e)
				}
			}

			if m.isProtocolPrint {
				log.Println("List of conflicting objects")
				for i := 0; i < len(conflictObj); i++ {
					log.Printf("K[%d] = %s\n", i, conflictObj[i].name)
				}
			}

			chosen := m.chooseObj(conflictObj)
			if m.isProtocolPrint {
				log.Printf("Chosen object %s\nNext event time: %f\nEvent %s starts for object %s\n", chosen.name, t, chosen.getEventMin().name, chosen.name)
			}

			chosen.doT()

			// proceed event
			chosen.stepEvent()

			if m.isProtocolPrint {
				log.Println("Exit from markers:")
				// print markers
			}
		}

	}

}

func (m *Model) goRun(timeModeling float64) {
	m.gtime.modTime = timeModeling

	t := 0.0
	var min float64

	m.sortObj(m.objects)
	for _, e := range m.objects {
		e.input()
	}

	if m.isProtocolPrint {
		for _, e := range m.objects {
			e.printMark()
		}
	}

	var K []Simulator
	for t < timeModeling {
		K = []Simulator{}

		min = m.getNextEventTime()
		if m.isStatistics {
			for _, e := range m.objects {
				e.doStatisticsWithInterval((min - t) / min)
			}
		}

		t = min
		m.gtime.currentTime = t

		if m.isProtocolPrint {
			log.Printf("Pass time through t: %f\n", t)
		}

		if t <= timeModeling {
			for _, e := range m.objects {
				if t == e.timeMin {
					K = append(K, e)
				}
			}

			var (
				num int
				max int
			)

			if m.isProtocolPrint {
				log.Println("List of conflicting objects")
				for i := 0; i < len(K); i++ {
					log.Printf("K[%d] = %s\n", i, K[i].name)
				}
			}

			if len(K) > 1 {
				max = len(K)
				m.sortObj(K)
				for i := 1; i < len(K); i++ {
					if K[i].priority < K[i-1].priority {
						max = i - 1
						break
					}
				}
				if max == 0 {
					num = 0
				} else {
					num = rand.Intn(max)
				}
			} else {
				num = 0
			}

			if m.isProtocolPrint {
				log.Printf("Chosen object %s -- next event\n", K[num].name)
			}

			for _, e := range m.objects {
				if e.numObject == K[num].numObject {
					if m.isProtocolPrint {
						log.Printf("time: %f -- event %s starts for object %s\n", t, e.getEventMin().name, e.name)
					}
					e.doT()
					e.stepEvent()
				}
			}

			if m.isProtocolPrint {
				log.Println("Exit markers from transitions")
				for _, e := range m.objects {
					e.printMark()
				}
			}

			m.sortObj(m.objects)
			for _, e := range m.objects {
				// check all conditions
				e.input()
			}

			if m.isProtocolPrint {
				log.Println("Enter markers into transitions")
				for _, e := range m.objects {
					e.printMark()
				}
			}
		}
	}
}
