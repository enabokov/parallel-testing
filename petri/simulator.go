package petri

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"sort"
)

type Simulator struct {
	Gcounter  *GlobalCounter
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

	Places      []*Place
	Transitions []*Transition
	LinksIn     []*Linker
	LinksOut    []*Linker

	EventMin *Transition
	TNet     Net

	StatisticsPlaces []*Place

	Channel chan int

	PrevObj *Simulator
	NextObj *Simulator

	TimeExternalInput []float64
	OutT              []*Transition
	InT               []*Transition

	BeginWait []string
	EndWait   []string

	Limit   int // 10
	Counter int // 0
}

type BuildSimulator interface {
	Build(Net, *GlobalCounter, *GlobalTime, *GlobalLocker, chan int) *Simulator

	GetEventMin() *Transition
	GetTimeExternalInput() []float64 // atomic
	SetPriority(int) BuildSimulator
	ProcessEventMin()
	FindActiveTransition() []*Transition
	SortTransitionsByPriority([]*Transition) // inplace
	Step()
	IsBufferEmpty() bool
	PrintMark()
	DoConflict([]*Transition) *Transition
	CheckIfOutTransitions([]*Transition, *Transition) bool
	Input()
	Output()
	ReinstateActOut(*Place, *Transition)
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
	PrintState()
	PrintBuffer()
}

func (s *Simulator) Build(n Net, c *GlobalCounter, t *GlobalTime, cond *GlobalLocker, channel chan int) *Simulator {
	s.TNet = n
	s.Name = n.Name
	s.Gcounter = c
	s.Channel = channel
	s.InitNumObj()
	s.IncrCounter()
	s.Gtime = t
	s.TimeLocal = s.Gtime.CurrentTime
	s.TimeMin = math.MaxFloat64
	s.Limit = 10
	s.Counter = 0
	copy(s.Places, n.Places)
	copy(s.Transitions, n.Transitions)
	copy(s.LinksIn, n.LinksIn)
	copy(s.LinksOut, n.LinksOut)
	s.EventMin = s.GetEventMin()
	s.Priority = 0
	copy(s.StatisticsPlaces, s.Places)

	// WARNING READ SOME FILE

	return s
}

func (s *Simulator) InitNumObj() {
	s.Gcounter.Lock()
	s.NumObject = s.Gcounter.Simulator
	s.Gcounter.Unlock()
}

func (s *Simulator) IncrCounter() {
	s.Gcounter.Lock()
	s.Gcounter.Simulator++
	s.Gcounter.Unlock()
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

func (s *Simulator) GetEventMin() *Transition {
	s.ProcessEventMin()
	return s.EventMin
}

func (s *Simulator) ProcessEventMin() {
	var event *Transition
	min := math.MaxFloat64

	for i := 0; i < len(s.Transitions); i++ {
		if s.Transitions[i].MinTime < min {
			event = s.Transitions[i]
			min = s.Transitions[i].MinTime
		}
	}

	s.TimeMin = min
	s.EventMin = event
}

func (s *Simulator) FindActiveTransition() []*Transition {
	var activeTransitions []*Transition

	for i := 0; i < len(s.Transitions); i++ {
		if s.Transitions[i].Condition(s.Places) && s.Transitions[i].Probability != 0 {
			activeTransitions = append(activeTransitions, s.Transitions[i])
		}
	}

	if len(activeTransitions) > 1 {
		log.Printf("Before sorting: %v\n", activeTransitions)
		s.SortTransitionsByPriority(activeTransitions)
		log.Printf("After sorting: %v\n", activeTransitions)
	}

	return activeTransitions
}

func (s *Simulator) SortTransitionsByPriority(t []*Transition) {
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
		for i := 0; i < len(s.Places); i++ {
			s.Places[i].SetMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
		}

		for i := 0; i < len(s.Transitions); i++ {
			s.Transitions[i].SetMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
		}

		// propagating time
		s.Gtime.CurrentTime = s.TimeMin

		return
	}

	for len(activeTransitions) > 0 {
		// resolving conflicts
		tmpTransition := s.DoConflict(activeTransitions)
		tmpTransition.ActIn(s.Places, s.Gtime.CurrentTime)

		// refresh list of active transitions
		activeTransitions = s.FindActiveTransition()
	}

	// find the closest event and its time
	s.ProcessEventMin()

	for i := 0; i < len(s.Places); i++ {
		s.Places[i].SetMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}

	for i := 0; i < len(s.Transitions); i++ {
		s.Transitions[i].SetMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}

	// propagate time
	s.Gtime.Lock()
	s.Gtime.CurrentTime = s.Gtime.ModTime
	s.Gtime.Unlock()

	if s.Gtime.CurrentTime <= s.Gtime.ModTime {

		// exit markers
		s.EventMin.ActOut(s.Places)

		if s.EventMin.Buffer > 0 {
			u := true
			for u {
				s.EventMin.MinEvent()
				if s.EventMin.MinTime == s.Gtime.CurrentTime {
					s.EventMin.ActOut(s.Places)
				} else {
					u = false
				}
			}
		}

		// WARNING: Output from all transitions
		// time of out markers == current time
		for i := 0; i < len(s.Transitions); i++ {
			if s.Transitions[i].Buffer > 0 && s.Transitions[i].MinTime == s.Gtime.CurrentTime {

				// exit markers from transition that responds to the closest time range
				s.Transitions[i].ActOut(s.Places)

				if s.Transitions[i].Buffer > 0 {
					u := true
					for u {
						s.Transitions[i].MinEvent()
						if s.Transitions[i].MinTime == s.Gtime.CurrentTime {
							s.Transitions[i].ActOut(s.Places)
						} else {
							u = false
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) DoConflict(t []*Transition) *Transition {
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
	for i := 0; i < len(s.Transitions); i++ {
		if s.Transitions[i].Buffer > 0 {
			c = false
			break
		}
	}

	return c
}

func (s *Simulator) PrintMark() {
	log.Printf("Mark in Net %s", s.Name)
	for i := 0; i < len(s.Places); i++ {
		log.Printf("- %f -", s.Places[i].Mark)
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
			t.ActIn(s.Places, s.TimeLocal)
			activeTransitions = s.FindActiveTransition()
		}

		s.ProcessEventMin()
	}
}

func (s *Simulator) Output() {
	var externalPlace *Place
	if s.NextObj != nil {
		externalPlace = s.Places[len(s.Places)-1]
		externalPlace.External = true
	}

	for i := 0; i < len(s.Transitions); i++ {
		if s.Transitions[i].MinTime == s.TimeLocal && s.Transitions[i].Buffer > 0 {
			s.Transitions[i].ActOut(s.Places)

			if s.NextObj != nil && s.CheckIfOutTransitions(s.OutT, s.Transitions[i]) {
				s.NextObj.AddTimeExternalInput(s.TimeLocal)
				s.NextObj.Channel <- 1

				for len(s.NextObj.TimeExternalInput) > s.Limit {
					log.Println("Wait for others")
					<-s.NextObj.Channel
					log.Println("Continue to processed further")
				}
			}

			if s.Transitions[i].Buffer > 0 {
				u := true
				for u {
					s.Transitions[i].MinEvent()
					if s.Transitions[i].MinTime == s.TimeLocal {
						s.Transitions[i].ActOut(s.Places)
						if s.NextObj != nil && s.CheckIfOutTransitions(s.OutT, s.Transitions[i]) {
							s.NextObj.AddTimeExternalInput(s.TimeLocal)
							s.NextObj.Channel <- 1
							for len(s.NextObj.TimeExternalInput) > s.Limit {
								<-s.Channel
							}
						}
					} else {
						u = false
					}
				}
			}

		}
	}
}

func (s *Simulator) CheckIfOutTransitions(t []*Transition, tofind *Transition) bool {
	for _, transition := range t {
		if transition == tofind {
			return true
		}
	}

	return false
}

func (s *Simulator) ReinstateActOut(p *Place, t *Transition) {
	for i := 0; i < len(s.PrevObj.LinksOut); i++ {
		if s.PrevObj.LinksOut[i].CounterTransitions == t.Number && s.PrevObj.LinksOut[i].CounterPlaces == p.Number {
			p.IncrMark(float64(s.PrevObj.LinksOut[i].KVariant))
			s.Counter++
			break
		} else {
			log.Printf("%d == %d && %d == %d", s.PrevObj.LinksOut[i].CounterTransitions, t.Number, s.PrevObj.LinksOut[i].CounterPlaces, p.Number)
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
	for i := 0; i < len(s.Transitions); i++ {
		if s.Transitions[i].Condition(s.Places) {
			return false
		}
		if s.Transitions[i].Buffer > 0 {
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
	for i := 0; i < len(s.Places); i++ {
		s.Places[i].SetMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}

	for i := 0; i < len(s.Transitions); i++ {
		s.Transitions[i].SetMean((s.TimeMin - s.Gtime.CurrentTime) / s.Gtime.ModTime)
	}
}

func (s *Simulator) DoStatisticsWithInterval(interval float64) {
	if interval > 0 {
		for i := 0; i < len(s.StatisticsPlaces); i++ {
			s.StatisticsPlaces[i].SetMean(interval)
		}

		for i := 0; i < len(s.Transitions); i++ {
			s.Transitions[i].SetMean(interval)
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
	s.NextObj.TimeExternalInput = append(s.NextObj.TimeExternalInput, s.TimeLocal)
}

func (s *Simulator) IsStopSerial() bool {
	s.ProcessEventMin()
	return reflect.DeepEqual(s.EventMin, nil)
}

func (s *Simulator) GoUntilConference(limitTime float64) {
	limit := float64(s.Limit)
	for s.TimeLocal < limit {
		for s.IsStop() {
			<-s.Channel
		}

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
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					s.NextObj.Channel <- 1
				}
			}

			if s.PrevObj != nil {
				for len(s.TimeExternalInput) == 0 {
					<-s.Channel
				}
			}

			if s.TimeExternalInput[0] > s.Gtime.ModTime {
				if s.NextObj != nil {
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					s.NextObj.Channel <- 1
				}
			} else if s.TimeExternalInput[0] == s.TimeLocal {
				s.ReinstateActOut(s.PrevObj.Places[len(s.PrevObj.Places)-1], s.PrevObj.OutT[0])
				s.TimeExternalInput = s.TimeExternalInput[1:]

				if len(s.TimeExternalInput) <= s.Limit {
					s.PrevObj.Channel <- 1
				}

				for len(s.TimeExternalInput) == 0 {
					<-s.Channel
				}

				if len(s.TimeExternalInput) > 0 {
					if s.TimeExternalInput[0] > s.Gtime.ModTime {
						if s.NextObj != nil {
							s.NextObj.AddTimeExternalInput(math.MaxFloat64)
							s.NextObj.Channel <- 1
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
		// checking preCondition start Input
		for s.IsStop() {
			log.Printf("%s is waiting for Input...\n", s.Name)
			<-s.Channel
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
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					s.NextObj.Channel <- 1
				}
			} else {
				if s.PrevObj != nil {
					if len(s.TimeExternalInput) == 0 || s.TimeExternalInput[len(s.TimeExternalInput)-1] < math.MaxFloat64 {
						for len(s.TimeExternalInput) == 0 {
							<-s.Channel
						}
					}

					if s.TimeExternalInput[0] > s.Gtime.ModTime {
						s.MoveTimeLocal(s.Gtime.ModTime)
						if s.NextObj != nil {
							s.NextObj.AddTimeExternalInput(math.MaxFloat64)
							s.NextObj.Channel <- 1
						}
					} else {
						s.MoveTimeLocal(limit)
						s.ReinstateActOut(s.PrevObj.Places[len(s.PrevObj.Places)-1], s.PrevObj.OutT[0])
						s.TimeExternalInput = s.TimeExternalInput[1:]

						if len(s.TimeExternalInput) <= s.Limit {
							s.PrevObj.Channel <- 1
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
			log.Printf("Lock: %s\n", s.Name)
			for len(s.TimeExternalInput) == 0 {
				log.Printf("Wait: %s\n", s.Name)
				<-s.Channel
			}

			log.Println("Len", len(s.TimeExternalInput))
			limitTime = s.TimeExternalInput[0]
			if limitTime > s.Gtime.ModTime {
				limitTime = s.Gtime.ModTime
			}

			log.Printf("Unlock: %s\n", s.Name)
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
	s.PrintState()
}

func (s *Simulator) DoT() {}

func (s *Simulator) PrintState() {
	s.PrintMark()
	s.PrintBuffer()
}

func (s *Simulator) PrintBuffer() {
	log.Printf("Buffer in Net %s: ", s.Name)
	for _, t := range s.Transitions {
		log.Printf("%d ", t.Buffer)
	}
	log.Println()
}
