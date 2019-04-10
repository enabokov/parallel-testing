package petri

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"sync"
)

type Simulator struct {
	gtime     *globalTime
	timeLocal float64
	name      string
	numObject int
	priority  int
	timeMin   float64

	numP   int
	numT   int
	numIn  int
	numOut int

	places      []Place
	transitions []Transition
	linksIn     []Linker
	linksOut    []Linker

	eventMin Transition
	net      Net

	statisticsPlaces []Place

	mux sync.Mutex

	prevObj *Simulator
	nextObj *Simulator

	// mutex locks

	timeExternalInput []float64
	outT              []Transition
	inT               []Transition

	beginWait []string
	endWait   []string

	limit   int // 10
	counter int // 0
}

type BuildSimulator interface {
	build(Net, *globalCounter, *globalTime) Simulator

	getEventMin() Transition
	getTimeExternalInput() []float64 // atomic
	setPriority(int) BuildSimulator
	processEventMin()
	findActiveTransition() []Transition
	sortTransitionsByPriority([]Transition) // inplace
	step()
	isBufferEmpty() bool
	printMark()
	doConflict([]Transition) Transition
	checkIfOutTransitions([]Transition, Transition) bool
	input()
	output()
	reinstateActOut(Place, Transition)
	stepEvent()
	isStop() bool
	doStatistics()
	doStatisticsWithInterval(float64)
	writeStatistics()
	goo()
	addTimeExternalInput(float64)
	isStopSerial() bool
	goUntilConference(float64)
	goUntil(float64)
	moveTimeLocal(float64)
	doT()
	run()
}

func (s *Simulator) build(n Net, c *globalCounter, t *globalTime) Simulator {
	s.net = n
	s.name = n.name
	s.numObject = c.simulator
	c.simulator++
	s.gtime = t
	s.timeLocal = s.gtime.currentTime
	s.timeMin = math.MaxFloat64

	copy(s.places, n.places[:])
	copy(s.transitions, n.transitions[:])
	copy(s.linksIn, n.linksIn[:])
	copy(s.linksOut, n.linksOut[:])
	s.eventMin = s.getEventMin()
	s.priority = 0
	copy(s.statisticsPlaces, s.places)

	// WARNING READ SOME FILE

	return *s
}

func (s *Simulator) setPriority(p int) BuildSimulator {
	s.priority = p
	return s
}

func (s *Simulator) getTimeExternalInput() []float64 {
	return s.timeExternalInput
}

func (s *Simulator) getEventMin() Transition {
	s.processEventMin()
	return s.eventMin
}

func (s *Simulator) processEventMin() {
	var event Transition
	min := math.MaxFloat64

	for _, t := range s.transitions {
		if t.minTime < min {
			event = t
			min = t.minTime
		}
	}

	s.timeMin = min
	s.eventMin = event
}

func (s *Simulator) findActiveTransition() []Transition {
	var activeTransitions []Transition
	for _, t := range s.transitions {
		if t.condition(s.places) && t.probability != 0 {
			activeTransitions = append(activeTransitions, t)
		}
	}

	if len(activeTransitions) > 1 {
		log.Printf("Before sorting: %v\n", activeTransitions)
		s.sortTransitionsByPriority(activeTransitions)
		log.Printf("After sorting: %v\n", activeTransitions)
	}

	return activeTransitions
}

func (s *Simulator) sortTransitionsByPriority(t []Transition) {
	sort.SliceStable(t[:], func(i, j int) bool {
		return t[i].priority < t[j].priority
	})
}

func (s *Simulator) step() {
	log.Printf("[next step] time: %f\n", s.gtime.currentTime)
	s.printMark()
	activeTransitions := s.findActiveTransition()

	if (len(activeTransitions) == 0 && s.isBufferEmpty()) || (s.gtime.currentTime >= s.gtime.modTime) {
		log.Printf("[stop] in Net %s\n", s.name)
		s.timeMin = s.gtime.modTime
		for _, p := range s.places {
			p.setMean((s.timeMin - s.gtime.currentTime) / s.gtime.modTime)
		}

		for _, t := range s.transitions {
			t.setMean((s.timeMin - s.gtime.currentTime) / s.gtime.modTime)
		}

		// propagating time
		s.gtime.currentTime = s.timeMin

		return
	}

	for len(activeTransitions) > 0 {
		// resolving conflicts
		s.doConflict(activeTransitions).actIn(s.places, s.gtime.currentTime)

		// refresh list of active transitions
		activeTransitions = s.findActiveTransition()
	}

	// find the closest event and its time
	s.processEventMin()

	for _, p := range s.places {
		p.setMean((s.timeMin - s.gtime.currentTime) / s.gtime.modTime)
	}

	for _, t := range s.transitions {
		t.setMean((s.timeMin - s.gtime.currentTime) / s.gtime.modTime)
	}

	// propagate time
	s.gtime.currentTime = s.gtime.modTime

	if s.gtime.currentTime <= s.gtime.modTime {

		// exit markers
		s.eventMin.actOut(s.places)

		if s.eventMin.buffer > 0 {
			u := true
			for u {
				s.eventMin.minEvent()
				if s.eventMin.minTime == s.gtime.currentTime {
					s.eventMin.actOut(s.places)
				} else {
					u = false
				}
			}
		}

		// WARNING: output from all transitions
		// time of out markers == current time
		for _, t := range s.transitions {
			if t.buffer > 0 && t.minTime == s.gtime.currentTime {

				// exit markers from transition that responds to the closest time range
				t.actOut(s.places)

				if t.buffer > 0 {
					u := true
					for u {
						t.minEvent()
						if t.minTime == s.gtime.currentTime {
							t.actOut(s.places)
						} else {
							u = false
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) doConflict(t []Transition) Transition {
	firstT := t[0]
	if len(t) > 1 {
		firstT = t[0]
		i := 0
		for i < len(t) && t[i].priority == firstT.priority {
			i++
		}

		if i > 1 {
			r := rand.Float64()

			j := 0
			var sum float64 = 0
			var prob float64

			for j < len(t) && t[j].priority == firstT.priority {
				if t[j].probability == 1.0 {
					prob = 1.0 / float64(i)
				} else {
					prob = t[j].probability
				}

				sum += prob
				if r < sum {
					firstT = t[j]
					break
				} else {
					j++
				}
			}
		}
	}

	return firstT
}

func (s *Simulator) isBufferEmpty() bool {
	c := true
	for _, t := range s.transitions {
		if t.buffer > 0 {
			c = false
			break
		}
	}

	return c
}

func (s *Simulator) printMark() {
	log.Printf("Mark in Net %s\n", s.name)
	for _, p := range s.places {
		log.Printf("- %f -", p.mark)
	}

	log.Println()
}

func (s *Simulator) input() {
	activeTransitions := s.findActiveTransition()
	if len(activeTransitions) == 0 && s.isBufferEmpty() {
		s.timeMin = math.MaxFloat64
	} else {
		for len(activeTransitions) > 0 {
			t := s.doConflict(activeTransitions)
			t.actIn(s.places, s.timeLocal)
			activeTransitions = s.findActiveTransition()
		}

		s.processEventMin()
	}
}

func (s *Simulator) output() {
	var externalPlace Place
	if s.nextObj != nil {
		externalPlace = s.places[len(s.places)-1]
		externalPlace.external = true
	}

	for _, t := range s.transitions {
		if t.minTime == s.timeLocal && t.buffer > 0 {
			t.actOut(s.places)

			if s.nextObj != nil && s.checkIfOutTransitions(s.outT, t) {
				s.nextObj.addTimeExternalInput(s.timeLocal)
				s.nextObj.mux.Lock()
				for len(s.nextObj.timeExternalInput) > s.limit {
					// wait until others
				}
				s.nextObj.mux.Unlock()
			}

			if t.buffer > 0 {
				u := true
				for u {
					t.minEvent()
					if t.minTime == s.timeLocal {
						t.actOut(s.places)
						if s.nextObj != nil && s.checkIfOutTransitions(s.outT, t) {
							s.nextObj.addTimeExternalInput(s.timeLocal)
							s.nextObj.mux.Lock()
							for len(s.nextObj.timeExternalInput) > s.limit {
								// wait until others
							}
							s.nextObj.mux.Unlock()
						}
					} else {
						u = false
					}
				}
			}

		}
	}
}

func (s *Simulator) checkIfOutTransitions(t []Transition, tofind Transition) bool {
	for _, transition := range t {
		if &transition == &tofind {
			return true
		}
	}

	return false
}

func (s *Simulator) reinstateActOut(p Place, t Transition) {
	for _, l := range s.prevObj.linksOut {
		if l.counterTransitions == t.number && l.counterPlaces == p.number {
			p.incrMark(float64(l.kVariant))
			s.counter++
			break
		} else {
			log.Printf("%d == %d && %d == %d", l.counterTransitions, t.number, l.counterPlaces, p.number)
		}
	}
}

func (s *Simulator) stepEvent() {
	if s.isStop() {
		s.timeMin = math.MaxFloat64
	} else {
		s.output()
		s.input()
	}
}

func (s *Simulator) isStop() bool {
	for _, t := range s.transitions {
		if t.condition(s.places) {
			return false
		}
		if t.buffer > 0 {
			return false
		}
	}

	if s.prevObj != nil {
		if len(s.timeExternalInput) > 0 {
			return false
		}
	}

	if s.nextObj != nil {
		if len(s.nextObj.timeExternalInput) > 10 {
			return false
		}
	}

	return true
}

func (s *Simulator) doStatistics() {
	for _, p := range s.places {
		p.setMean((s.timeMin - s.gtime.currentTime) / s.gtime.modTime)
	}

	for _, t := range s.transitions {
		t.setMean((s.timeMin - s.gtime.currentTime) / s.gtime.modTime)
	}
}

func (s *Simulator) doStatisticsWithInterval(interval float64) {
	if interval > 0 {
		for _, p := range s.statisticsPlaces {
			p.setMean(interval)
		}

		for _, t := range s.transitions {
			t.setMean(interval)
		}
	}
}

func (s *Simulator) writeStatistics() {
	f, err := os.Create("./statistics.txt")
	if err != nil {
		log.Println(err)
	}

	_, err = f.WriteString(fmt.Sprintf("%f\t%f\t%f\n", s.places[0].mark, s.timeLocal, s.places[0].mean))
	log.Println(err)
}

func (s *Simulator) goo() {
	s.gtime.currentTime = 0
	for s.gtime.currentTime <= s.gtime.modTime && !s.isStop() {
		s.step()
		if s.isStop() {
			log.Printf("[STOP] in Net %s", s.name)
		}

		s.printMark()
	}
}

func (s *Simulator) addTimeExternalInput(t float64) {
	s.mux.Lock()
	s.nextObj.timeExternalInput = append(s.nextObj.timeExternalInput, s.timeLocal)
	s.mux.Unlock()
}

func (s *Simulator) isStopSerial() bool {
	s.processEventMin()
	return reflect.DeepEqual(s.eventMin, nil)
}

func (s *Simulator) goUntilConference(limitTime float64) {
	limit := float64(s.limit)
	for s.timeLocal < limit {
		//for s.isStop() {
		//	s.mux.Lock()
		//}

		s.input()
		if s.timeMin < limit {
			s.doStatisticsWithInterval((s.timeMin - s.timeLocal) / s.timeMin) // maybe / s.gtime.modTime
			s.timeLocal = s.timeMin
			s.output()
		} else {
			if s.prevObj != nil && len(s.timeExternalInput) > 0 {
				if s.timeExternalInput[len(s.timeExternalInput)-1] < math.MaxFloat64 {
					s.doStatisticsWithInterval((limit - s.timeLocal) / limit)
					s.timeLocal = limit
				} else {
					limit = s.gtime.modTime
					s.doStatisticsWithInterval((limit - s.timeLocal) / limit)
					s.timeLocal = limit
				}
			} else {
				s.doStatisticsWithInterval((limit - s.timeLocal) / limit)
				s.timeLocal = limit
			}

			if limit >= s.gtime.modTime {
				if s.nextObj != nil {
					s.nextObj.mux.Lock()
					s.nextObj.addTimeExternalInput(math.MaxFloat64)
					s.nextObj.mux.Unlock()
				}
			}

			if s.prevObj != nil {
				for len(s.timeExternalInput) == 0 {
					s.mux.Lock()
					// check and await
					s.mux.Unlock()
				}
			}

			if s.timeExternalInput[0] > s.gtime.modTime {
				if s.nextObj != nil {
					s.nextObj.mux.Lock()
					s.nextObj.addTimeExternalInput(math.MaxFloat64)
					s.nextObj.mux.Unlock()
				}
			} else if s.timeExternalInput[0] == s.timeLocal {
				s.reinstateActOut(s.prevObj.places[len(s.prevObj.places)-1], s.prevObj.outT[0])
				s.mux.Lock()
				s.timeExternalInput = s.timeExternalInput[1:]
				s.mux.Unlock()

				if len(s.timeExternalInput) <= s.limit {
					s.prevObj.mux.Lock()
					// lock condition
					s.prevObj.mux.Unlock()
				}

				for len(s.timeExternalInput) == 0 {
					s.mux.Lock()
					// lock condition
					s.mux.Unlock()
				}

				if len(s.timeExternalInput) > 0 {
					if s.timeExternalInput[0] > s.gtime.modTime {
						if s.nextObj != nil {
							s.nextObj.mux.Lock()
							s.nextObj.addTimeExternalInput(math.MaxFloat64)
							// lock condition
							s.nextObj.mux.Unlock()
						}
					} else {
						if s.timeExternalInput[len(s.timeExternalInput)-1] < math.MaxFloat64 {
							limit = s.timeExternalInput[0]
						} else {
							limit = s.timeExternalInput[len(s.timeExternalInput)-1]
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) moveTimeLocal(t float64) {
	s.doStatisticsWithInterval((t - s.timeLocal) / t)
	s.timeLocal = t
}

func (s *Simulator) goUntil(limitTime float64) {
	limit := limitTime

	// propagate time within interval range
	for s.timeLocal < limit {
		// checking precondition start input
		for s.isStop() {
			log.Printf("%s is waiting for input...\n", s.name)
			s.mux.Lock()
			// lock condition
			s.mux.Unlock()
		}

		// timeMin changed
		s.input()
		log.Printf("%s did input, new value of timeMin: %f and limitTime: %f", s.name, s.timeMin, limit)
		if s.timeMin < limit {
			s.moveTimeLocal(s.timeMin)
			s.output()
		} else {
			if limit >= s.gtime.modTime {
				s.moveTimeLocal(s.gtime.modTime)
				if s.nextObj != nil {
					s.nextObj.mux.Lock()
					s.nextObj.addTimeExternalInput(math.MaxFloat64)
					// lock condition
					s.nextObj.mux.Unlock()
				}
			} else {
				if s.prevObj != nil {
					if len(s.timeExternalInput) == 0 || s.timeExternalInput[len(s.timeExternalInput)-1] < math.MaxFloat64 {
						for len(s.timeExternalInput) == 0 {
							s.mux.Lock()
							// cond lock
							s.mux.Unlock()
						}
					}

					if s.timeExternalInput[0] > s.gtime.modTime {
						s.moveTimeLocal(s.gtime.modTime)
						if s.nextObj != nil {
							s.nextObj.mux.Lock()
							// cond lock
							s.nextObj.mux.Unlock()
						} else {
							s.moveTimeLocal(limit)
							s.reinstateActOut(s.prevObj.places[len(s.prevObj.places)-1], s.prevObj.outT[0])
							s.mux.Lock()
							s.timeExternalInput = s.timeExternalInput[1:]
							s.mux.Unlock()

							if len(s.timeExternalInput) <= s.limit {
								s.prevObj.mux.Lock()
								// cond lock
								s.prevObj.mux.Unlock()
							}
						}
					}
				} else {
					s.moveTimeLocal(limit)
				}
			}
		}
	}
}

func (s *Simulator) run() {
	for s.timeLocal < s.gtime.modTime {
		limitTime := s.gtime.modTime
		if s.prevObj != nil {
			s.mux.Lock()
			for len(s.timeExternalInput) == 0 {
				// lock cond
			}
			limitTime = s.timeExternalInput[0]
			if limitTime > s.gtime.modTime {
				limitTime = s.gtime.modTime
			}

			s.mux.Unlock()
		} else {
			limitTime = s.gtime.modTime
		}

		if s.timeLocal < limitTime {
			log.Printf("%s will go until %f have local time %f\n", s.name, limitTime, s.timeLocal)
			s.goUntil(limitTime)
		} else {
			log.Printf("%s will not go until %f have local time %f >= time modeling!\n", s.name, limitTime, s.timeLocal)
			return
		}
	}

	log.Printf("%s has finished simulation\n", s.name)
}

func (s *Simulator) doT() {}
