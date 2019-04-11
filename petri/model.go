package petri

import (
	"log"
	"math/rand"
	"sort"
	"sync"
)

type Model struct {
	Gtime           *GlobalTime
	Objects         []*Simulator
	TimeMod         float64
	T               float64
	IsProtocolPrint bool
	IsStatistics    bool
}

type BuildModel interface {
	Build([]Simulator) Model
	GetNextEventTime() float64
	ModelInput()
	SortObj([]Simulator)
	ChooseObj([]Simulator) Simulator
	ParallelGo(float64)
	GoRun(float64)
}

func (m *Model) Build(s []*Simulator, gtime *GlobalTime) Model {
	m.Objects = s
	m.Gtime = gtime
	return *m
}

func (m *Model) GetNextEventTime() float64 {
	min := m.Objects[0].TimeMin

	for i := 0; i < len(m.Objects); i++ {
		if m.Objects[i].TimeMin < min {
			min = m.TimeMod
		}
	}

	return min
}

func (m *Model) ModelInput() {
	m.SortObj(m.Objects)
	var wg sync.WaitGroup

	for i := 0; i < len(m.Objects); i++ {
		wg.Add(1)
		go func() {
			it := i
			m.Objects[it].Input()
			wg.Done()
		}()
	}

	wg.Wait()
}

func (m *Model) SortObj(s []*Simulator) {
	sort.SliceStable(s, func(i, j int) bool {
		return s[i].Priority < s[j].Priority
	})
}

func (m *Model) ChooseObj(s []*Simulator) *Simulator {
	var num int
	var max int

	if len(s) > 1 {
		max = len(s)
		m.SortObj(s)
		for i := 1; i < len(s); i++ {
			if s[i].Priority < s[i-1].Priority {
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

func (m *Model) ParallelGo(timeModeling float64) {
	m.TimeMod = timeModeling

	m.T = 0.0
	var min float64

	if m.IsProtocolPrint {
		log.Println("Start marking Objects:")
		// Print marks
	}

	var conflictObj []*Simulator
	for m.T < timeModeling {
		conflictObj = []*Simulator{}

		// maybe Conditions changed
		m.ModelInput()
		if m.IsProtocolPrint {
			log.Println("Enter markers into transitions")
			// Print marks
		}

		// search the closest event
		min = m.GetNextEventTime()
		if m.IsStatistics {
			for i := 0; i < len(m.Objects); i++ {
				// statistics within delta m.T
				// statistics is collected only once for all common positions
				m.Objects[i].DoStatisticsWithInterval((min - m.T) / min)
			}
		}

		// pass time further
		m.T = min

		m.Gtime.CurrentTime = m.T

		if m.IsProtocolPrint {
			log.Printf("Passing time further. m.T: %f", m.T)
		}

		if m.T <= timeModeling {
			for i := 0; i < len(m.Objects); i++ {
				if m.T == m.Objects[i].TimeMin {
					conflictObj = append(conflictObj, m.Objects[i])
				}
			}

			if m.IsProtocolPrint {
				log.Println("List of conflicting Objects")
				for i := 0; i < len(conflictObj); i++ {
					log.Printf("K[%d] = %s\n", i, conflictObj[i].Name)
				}
			}

			chosen := m.ChooseObj(conflictObj)
			if m.IsProtocolPrint {
				log.Printf("Chosen object %s\nNext event time: %f\nEvent %s starts for object %s\n", chosen.Name, m.T, chosen.GetEventMin().Name, chosen.Name)
			}

			chosen.DoT()

			// proceed event
			chosen.StepEvent()

			if m.IsProtocolPrint {
				log.Println("Exit from markers:")
				// Print markers
			}
		}

	}

}

func (m *Model) GoRun(timeModeling float64) {
	m.Gtime.Lock()
	m.Gtime.ModTime = timeModeling
	m.Gtime.Unlock()

	m.T = 0.0
	var min float64

	m.SortObj(m.Objects)
	for _, e := range m.Objects {
		e.Input()
	}

	if m.IsProtocolPrint {
		for _, e := range m.Objects {
			e.PrintMark()
		}
	}

	var K []*Simulator
	for m.T < timeModeling {
		K = []*Simulator{}

		min = m.GetNextEventTime()
		if m.IsStatistics {
			for _, e := range m.Objects {
				e.DoStatisticsWithInterval((min - m.T) / min)
			}
		}

		m.T = min
		m.Gtime.CurrentTime = m.T

		if m.IsProtocolPrint {
			log.Printf("Pass time through m.T: %f\n", m.T)
		}

		if m.T <= timeModeling {
			for i := 0; i < len(m.Objects); i++ {
				if m.T == m.Objects[i].TimeMin {
					K = append(K, m.Objects[i])
				}
			}

			var (
				num int
				max int
			)

			if m.IsProtocolPrint {
				log.Println("List of conflicting Objects")
				for i := 0; i < len(K); i++ {
					log.Printf("K[%d] = %s\n", i, K[i].Name)
				}
			}

			if len(K) > 1 {
				max = len(K)
				m.SortObj(K)
				for i := 1; i < len(K); i++ {
					if K[i].Priority < K[i-1].Priority {
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

			if m.IsProtocolPrint {
				log.Printf("Chosen object %s -- next event\n", K[num].Name)
			}

			for _, e := range m.Objects {
				if e.NumObject == K[num].NumObject {
					if m.IsProtocolPrint {
						log.Printf("time: %f -- event %s starts for object %s\n", m.T, e.GetEventMin().Name, e.Name)
					}
					e.DoT()
					e.StepEvent()
				}
			}

			if m.IsProtocolPrint {
				log.Println("Exit markers from transitions")
				for _, e := range m.Objects {
					e.PrintMark()
				}
			}

			m.SortObj(m.Objects)
			for _, e := range m.Objects {
				// check all Conditions
				e.Input()
			}

			if m.IsProtocolPrint {
				log.Println("Enter markers into transitions")
				for _, e := range m.Objects {
					e.PrintMark()
				}
			}
		}
	}
}
