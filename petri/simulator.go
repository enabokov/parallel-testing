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
	Gtime     *GlobalTime
	TimeLocal float64
	Name      string
	NumObject int
	Priority  int
	TimeMin   float64

	NumP   int
	NumT   int
	NumIn  int
	NumOut int

	Places      []Place
	Transitions []Transition
	LinksIn     []Linker
	LinksOut    []Linker

	EventMin Transition
	TNet     Net

	StatisticsPlaces []Place

	Mux sync.Mutex

	PrevObj *Simulator
	NextObj *Simulator

	// mutex locks

	TimeExternalInput []float64
	OutT              []Transition
	InT               []Transition

	BeginWait []string
	EndWait   []string

	Limit   int // 10
	Counter int // 0
}

type BuildSimulator interface {
	Build(Net, *GlobalCounter, *GlobalTime) Simulator

	GetEventMin() Transition
	GetTimeExternalInput() []float64 // atomic
	SetPriority(int) BuildSimulator
	ProcessEventMin()
	FindActiveTransition() []Transition
	SortTransitionsByPriority([]Transition) // inplace
	Step()
	IsBufferEmpty() bool
	PrintMark()
	DoConflict([]Transition) Transition
	CheckIfOutTransitions([]Transition, Transition) bool
	Input()
	Output()
	ReinstateActOut(Place, Transition)
	StepEvent()
	IsStop() bool
	DoStatistics()
	DoStatisticsWithInterval(float64)
	WriteStatistics()
	Goo()
	AddTimeExternalInput(float64)
	IsStopSerial() bool
	GoUntilConference(float64)
	GoUntil(float64)
	MoveTimeLocal(float64)
	DoT()
	Run()
}

func (s *Simulator) Build(n Net, c *GlobalCounter, t *GlobalTime) Simulator {
	s.TNet = n
	s.Name = n.Name
	s.NumObject = c.Simulator
	c.Simulator++
	s.Gtime = t
	s.TimeLocal = s.Gtime.CurrentTime
	s.TimeMin = math.MaxFloat64

	copy(s.Places, n.Places[:])
	copy(s.Transitions, n.Transitions[:])
	copy(s.LinksIn, n.LinksIn[:])
	copy(s.LinksOut, n.LinksOut[:])
	s.EventMin = s.GetEventMin()
	s.Priority = 0
	copy(s.StatisticsPlaces, s.Places)

	// WARNING READ SOME FILE

	return *s
}

func (s *Simulator) SetPriority(p int) BuildSimulator {
	s.Priority = p
	return s
}

func (s *Simulator) GetNet() Net {
	return s.TNet
}

func (s *Simulator) GetTimeExternalInput() []float64 {
	return s.TimeExternalInput
}

func (s *Simulator) GetEventMin() Transition {
	s.ProcessEventMin()
	return s.EventMin
}

func (s *Simulator) ProcessEventMin() {
	var event Transition
	min := math.MaxFloat64

	for _, t := range s.Transitions {
		if t.MinTime < min {
			event = t
			min = t.MinTime
		}
	}

	s.TimeMin = min
	s.EventMin = event
}

func (s *Simulator) FindActiveTransition() []Transition {
	var activeTransitions []Transition
	for _, t := range s.Transitions {
		if t.condition(s.Places) && t.Probability != 0 {
			activeTransitions = append(activeTransitions, t)
		}
	}

	if len(activeTransitions) > 1 {
		log.Printf("Before sorting: %v\n", activeTransitions)
		s.SortTransitionsByPriority(activeTransitions)
		log.Printf("After sorting: %v\n", activeTransitions)
	}

	return activeTransitions
}

func (s *Simulator) SortTransitionsByPriority(t []Transition) {
	sort.SliceStable(t[:], func(i, j int) bool {
		return t[i].Priority < t[j].Priority
	})
}

func (s *Simulator) Step() {
	log.Printf("[next Step] time: %f\n", s.Gtime.CurrentTime)
	s.PrintMark()
	activeTransitions := s.FindActiveTransition()

	if (len(activeTransitions) == 0 && s.IsBufferEmpty()) || (s.Gtime.CurrentTime >= s.Gtime.ModTime) {
		log.Printf("[stop] in Net %s\n", s.Name)
		s.TimeMin = s.Gtime.ModTime
		for _, p := range s.Places {
			p.setMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
		}

		for _, t := range s.Transitions {
			t.setMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
		}

		// propagating time
		s.Gtime.CurrentTime = s.TimeMin

		return
	}

	for len(activeTransitions) > 0 {
		// resolving conflicts
		s.DoConflict(activeTransitions).actIn(s.Places, s.Gtime.CurrentTime)

		// refresh list of active transitions
		activeTransitions = s.FindActiveTransition()
	}

	// find the closest event and its time
	s.ProcessEventMin()

	for _, p := range s.Places {
		p.setMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}

	for _, t := range s.Transitions {
		t.setMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}

	// propagate time
	s.Gtime.CurrentTime = s.Gtime.ModTime

	if s.Gtime.CurrentTime <= s.Gtime.ModTime {

		// exit markers
		s.EventMin.actOut(s.Places)

		if s.EventMin.Buffer > 0 {
			u := true
			for u {
				s.EventMin.minEvent()
				if s.EventMin.MinTime == s.Gtime.CurrentTime {
					s.EventMin.actOut(s.Places)
				} else {
					u = false
				}
			}
		}

		// WARNING: Output from all transitions
		// time of out markers == current time
		for _, t := range s.Transitions {
			if t.Buffer > 0 && t.MinTime == s.Gtime.CurrentTime {

				// exit markers from transition that responds to the closest time range
				t.actOut(s.Places)

				if t.Buffer > 0 {
					u := true
					for u {
						t.minEvent()
						if t.MinTime == s.Gtime.CurrentTime {
							t.actOut(s.Places)
						} else {
							u = false
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) DoConflict(t []Transition) Transition {
	firstT := t[0]
	if len(t) > 1 {
		firstT = t[0]
		i := 0
		for i < len(t) && t[i].Priority == firstT.Priority {
			i++
		}

		if i > 1 {
			r := rand.Float64()

			j := 0
			var sum float64 = 0
			var prob float64

			for j < len(t) && t[j].Priority == firstT.Priority {
				if t[j].Probability == 1.0 {
					prob = 1.0 / float64(i)
				} else {
					prob = t[j].Probability
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

func (s *Simulator) IsBufferEmpty() bool {
	c := true
	for _, t := range s.Transitions {
		if t.Buffer > 0 {
			c = false
			break
		}
	}

	return c
}

func (s *Simulator) PrintMark() {
	log.Printf("Mark in Net %s\n", s.Name)
	for _, p := range s.Places {
		log.Printf("- %f -", p.Mark)
	}

	log.Println()
}

func (s *Simulator) Input() {
	activeTransitions := s.FindActiveTransition()
	if len(activeTransitions) == 0 && s.IsBufferEmpty() {
		s.TimeMin = math.MaxFloat64
	} else {
		for len(activeTransitions) > 0 {
			t := s.DoConflict(activeTransitions)
			t.actIn(s.Places, s.TimeLocal)
			activeTransitions = s.FindActiveTransition()
		}

		s.ProcessEventMin()
	}
}

func (s *Simulator) Output() {
	var externalPlace Place
	if s.NextObj != nil {
		externalPlace = s.Places[len(s.Places)-1]
		externalPlace.External = true
	}

	for _, t := range s.Transitions {
		if t.MinTime == s.TimeLocal && t.Buffer > 0 {
			t.actOut(s.Places)

			if s.NextObj != nil && s.CheckIfOutTransitions(s.OutT, t) {
				s.NextObj.AddTimeExternalInput(s.TimeLocal)
				s.NextObj.Mux.Lock()
				for len(s.NextObj.TimeExternalInput) > s.Limit {
					// wait until others
				}
				s.NextObj.Mux.Unlock()
			}

			if t.Buffer > 0 {
				u := true
				for u {
					t.minEvent()
					if t.MinTime == s.TimeLocal {
						t.actOut(s.Places)
						if s.NextObj != nil && s.CheckIfOutTransitions(s.OutT, t) {
							s.NextObj.AddTimeExternalInput(s.TimeLocal)
							s.NextObj.Mux.Lock()
							for len(s.NextObj.TimeExternalInput) > s.Limit {
								// wait until others
							}
							s.NextObj.Mux.Unlock()
						}
					} else {
						u = false
					}
				}
			}

		}
	}
}

func (s *Simulator) CheckIfOutTransitions(t []Transition, tofind Transition) bool {
	for _, transition := range t {
		if &transition == &tofind {
			return true
		}
	}

	return false
}

func (s *Simulator) ReinstateActOut(p Place, t Transition) {
	for _, l := range s.PrevObj.LinksOut {
		if l.counterTransitions == t.Number && l.counterPlaces == p.Number {
			p.incrMark(float64(l.kVariant))
			s.Counter++
			break
		} else {
			log.Printf("%d == %d && %d == %d", l.counterTransitions, t.Number, l.counterPlaces, p.Number)
		}
	}
}

func (s *Simulator) StepEvent() {
	if s.IsStop() {
		s.TimeMin = math.MaxFloat64
	} else {
		s.Output()
		s.Input()
	}
}

func (s *Simulator) IsStop() bool {
	for _, t := range s.Transitions {
		if t.condition(s.Places) {
			return false
		}
		if t.Buffer > 0 {
			return false
		}
	}

	if s.PrevObj != nil {
		if len(s.TimeExternalInput) > 0 {
			return false
		}
	}

	if s.NextObj != nil {
		if len(s.NextObj.TimeExternalInput) > 10 {
			return false
		}
	}

	return true
}

func (s *Simulator) DoStatistics() {
	for _, p := range s.Places {
		p.setMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}

	for _, t := range s.Transitions {
		t.setMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}
}

func (s *Simulator) DoStatisticsWithInterval(interval float64) {
	if interval > 0 {
		for _, p := range s.StatisticsPlaces {
			p.setMean(interval)
		}

		for _, t := range s.Transitions {
			t.setMean(interval)
		}
	}
}

func (s *Simulator) WriteStatistics() {
	f, err := os.Create("./statistics.txt")
	if err != nil {
		log.Println(err)
	}

	_, err = f.WriteString(fmt.Sprintf("%f\t%f\t%f\n", s.Places[0].Mark, s.TimeLocal, s.Places[0].Mean))
	log.Println(err)
}

func (s *Simulator) Goo() {
	s.Gtime.CurrentTime = 0
	for s.Gtime.CurrentTime <= s.Gtime.ModTime && !s.IsStop() {
		s.Step()
		if s.IsStop() {
			log.Printf("[STOP] in Net %s", s.Name)
		}

		s.PrintMark()
	}
}

func (s *Simulator) AddTimeExternalInput(t float64) {
	s.Mux.Lock()
	s.NextObj.TimeExternalInput = append(s.NextObj.TimeExternalInput, s.TimeLocal)
	s.Mux.Unlock()
}

func (s *Simulator) IsStopSerial() bool {
	s.ProcessEventMin()
	return reflect.DeepEqual(s.EventMin, nil)
}

func (s *Simulator) GoUntilConference(limitTime float64) {
	limit := float64(s.Limit)
	for s.TimeLocal < limit {
		//for s.IsStop() {
		//	s.Mux.Lock()
		//}

		s.Input()
		if s.TimeMin < limit {
			s.DoStatisticsWithInterval((s.TimeMin - s.TimeLocal) / s.TimeMin) // maybe / s.Gtime.ModTime
			s.TimeLocal = s.TimeMin
			s.Output()
		} else {
			if s.PrevObj != nil && len(s.TimeExternalInput) > 0 {
				if s.TimeExternalInput[len(s.TimeExternalInput)-1] < math.MaxFloat64 {
					s.DoStatisticsWithInterval((limit - s.TimeLocal) / limit)
					s.TimeLocal = limit
				} else {
					limit = s.Gtime.ModTime
					s.DoStatisticsWithInterval((limit - s.TimeLocal) / limit)
					s.TimeLocal = limit
				}
			} else {
				s.DoStatisticsWithInterval((limit - s.TimeLocal) / limit)
				s.TimeLocal = limit
			}

			if limit >= s.Gtime.ModTime {
				if s.NextObj != nil {
					s.NextObj.Mux.Lock()
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					s.NextObj.Mux.Unlock()
				}
			}

			if s.PrevObj != nil {
				for len(s.TimeExternalInput) == 0 {
					s.Mux.Lock()
					// check and await
					s.Mux.Unlock()
				}
			}

			if s.TimeExternalInput[0] > s.Gtime.ModTime {
				if s.NextObj != nil {
					s.NextObj.Mux.Lock()
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					s.NextObj.Mux.Unlock()
				}
			} else if s.TimeExternalInput[0] == s.TimeLocal {
				s.ReinstateActOut(s.PrevObj.Places[len(s.PrevObj.Places)-1], s.PrevObj.OutT[0])
				s.Mux.Lock()
				s.TimeExternalInput = s.TimeExternalInput[1:]
				s.Mux.Unlock()

				if len(s.TimeExternalInput) <= s.Limit {
					s.PrevObj.Mux.Lock()
					// lock condition
					s.PrevObj.Mux.Unlock()
				}

				for len(s.TimeExternalInput) == 0 {
					s.Mux.Lock()
					// lock condition
					s.Mux.Unlock()
				}

				if len(s.TimeExternalInput) > 0 {
					if s.TimeExternalInput[0] > s.Gtime.ModTime {
						if s.NextObj != nil {
							s.NextObj.Mux.Lock()
							s.NextObj.AddTimeExternalInput(math.MaxFloat64)
							// lock condition
							s.NextObj.Mux.Unlock()
						}
					} else {
						if s.TimeExternalInput[len(s.TimeExternalInput)-1] < math.MaxFloat64 {
							limit = s.TimeExternalInput[0]
						} else {
							limit = s.TimeExternalInput[len(s.TimeExternalInput)-1]
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) MoveTimeLocal(t float64) {
	s.DoStatisticsWithInterval((t - s.TimeLocal) / t)
	s.TimeLocal = t
}

func (s *Simulator) GoUntil(limitTime float64) {
	limit := limitTime

	// propagate time within interval range
	for s.TimeLocal < limit {
		// checking precondition start Input
		for s.IsStop() {
			log.Printf("%s is waiting for Input...\n", s.Name)
			s.Mux.Lock()
			// lock condition
			s.Mux.Unlock()
		}

		// timeMin changed
		s.Input()
		log.Printf("%s did Input, new value of timeMin: %f and limitTime: %f", s.Name, s.TimeMin, limit)
		if s.TimeMin < limit {
			s.MoveTimeLocal(s.TimeMin)
			s.Output()
		} else {
			if limit >= s.Gtime.ModTime {
				s.MoveTimeLocal(s.Gtime.ModTime)
				if s.NextObj != nil {
					s.NextObj.Mux.Lock()
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					// lock condition
					s.NextObj.Mux.Unlock()
				}
			} else {
				if s.PrevObj != nil {
					if len(s.TimeExternalInput) == 0 || s.TimeExternalInput[len(s.TimeExternalInput)-1] < math.MaxFloat64 {
						for len(s.TimeExternalInput) == 0 {
							s.Mux.Lock()
							// cond lock
							s.Mux.Unlock()
						}
					}

					if s.TimeExternalInput[0] > s.Gtime.ModTime {
						s.MoveTimeLocal(s.Gtime.ModTime)
						if s.NextObj != nil {
							s.NextObj.Mux.Lock()
							// cond lock
							s.NextObj.Mux.Unlock()
						} else {
							s.MoveTimeLocal(limit)
							s.ReinstateActOut(s.PrevObj.Places[len(s.PrevObj.Places)-1], s.PrevObj.OutT[0])
							s.Mux.Lock()
							s.TimeExternalInput = s.TimeExternalInput[1:]
							s.Mux.Unlock()

							if len(s.TimeExternalInput) <= s.Limit {
								s.PrevObj.Mux.Lock()
								// cond lock
								s.PrevObj.Mux.Unlock()
							}
						}
					}
				} else {
					s.MoveTimeLocal(limit)
				}
			}
		}
	}
}

func (s *Simulator) Run() {
	for s.TimeLocal < s.Gtime.ModTime {
		limitTime := s.Gtime.ModTime
		if s.PrevObj != nil {
			s.Mux.Lock()
			for len(s.TimeExternalInput) == 0 {
				// lock cond
			}
			limitTime = s.TimeExternalInput[0]
			if limitTime > s.Gtime.ModTime {
				limitTime = s.Gtime.ModTime
			}

			s.Mux.Unlock()
		} else {
			limitTime = s.Gtime.ModTime
		}

		if s.TimeLocal < limitTime {
			log.Printf("%s will go until %f have local time %f\n", s.Name, limitTime, s.TimeLocal)
			s.GoUntil(limitTime)
		} else {
			log.Printf("%s will not go until %f have local time %f >= time modeling!\n", s.Name, limitTime, s.TimeLocal)
			return
		}
	}

	log.Printf("%s has finished simulation\n", s.Name)
}

func (s *Simulator) DoT() {}
