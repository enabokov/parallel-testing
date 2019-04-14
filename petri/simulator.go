package petri

import (
	"log"
	"math"
	"math/rand"
	"sort"
)

type Simulator struct {
	GTimeMod *GTimeModeling
	GLimit *GLimitArrayExtInputs
	GTimeCurr *GTimeCurrent
	GChannel chan int

	TimeLocal   float64
	Name        string
	NumObj   int
	Priority    int
	TimeMin     float64

	NumP   int
	NumT   int
	NumIn  int
	NumOut int

	ListP      []Place
	ListT []Transition
	ListIn     []ArcIn
	ListOut    []ArcOut

	EventMin  *Transition
	NNet      *Net

	ListPositionsForStatistica []Place

	PrevObj *Simulator
	NextObj *Simulator

	// mutex locks

	TimeExternalInput []float64
	OutT              []Transition
	InT               []Transition

	BeginWait []string
	endWait   []string

	Counter int
}

func (s *Simulator) Build(pNet *Net, gnext *GNext, gmodeling *GTimeModeling, glimit *GLimitArrayExtInputs, gcurrent *GTimeCurrent, gchannel chan int) Simulator {
	s.GTimeMod = gmodeling
	s.GLimit = glimit
	s.GTimeCurr = gcurrent
	s.GChannel = gchannel

	s.NNet = pNet
	s.Name = s.NNet.GetName()
	s.NumObj = gnext.NextSimulator
	gnext.NextSimulator++

	s.TimeMin = math.MaxFloat64
	s.TimeLocal = s.GTimeCurr.TimeCurrentSimulation
	s.ListP = s.NNet.GetListP()
	s.ListT = s.NNet.GetListT()
	s.ListIn = s.NNet.GetArcIn()
	s.ListOut = s.NNet.GetArcOut()
	s.NumP = len(s.ListP)
	s.NumT = len(s.ListT)
	s.NumIn = len(s.ListIn)
	s.NumOut = len(s.ListOut)
	s.EventMin = s.GetEventMin()
	s.Priority = 0
	s.ListPositionsForStatistica = s.ListP

	// WARNING READ SOME FILE

	return *s
}

func (s *Simulator) GetEventMin() *Transition {
	s.eventMin()
	return s.EventMin
}

func (s *Simulator) AddOutT(transition Transition) {
	s.OutT = append(s.OutT, transition)
}

func (s *Simulator) AddInT(transition Transition) {
	s.InT = append(s.InT, transition)
}

func (s *Simulator) GetNet() *Net {
	return s.NNet
}

func (s *Simulator) GetName() string {
	return s.Name
}

func (s *Simulator) GetListPositionsForStats() []Place {
	return s.ListPositionsForStatistica
}

func (s *Simulator) GetPriority() int {
	return s.Priority
}

func (s *Simulator) GetNumObj() int {
	return s.NumObj
}

func (s *Simulator) SetPrioriry(a int) {
	s.Priority = a
}

func (s *Simulator) doT() {}

func (s *Simulator) eventMin() {
	var event *Transition
	min := math.MaxFloat64
	for i := 0; i < len(s.ListT); i++ {
		if s.ListT[i].getMinTime() < min {
			event = &s.ListT[i]
			min = s.ListT[i].getMinTime()
		}
	}
	s.setTimeMin(min)
	s.EventMin = event
}

func (s *Simulator) getTimeMin() float64 {
	return s.TimeMin
}

func (s *Simulator) sortT(tt []Transition) []Transition {
	sort.SliceStable(tt, func(i, j int) bool {
		return tt[i].GetPriority() < tt[j].GetPriority()
	})

	return tt
}

func (s *Simulator) findActiveT() []Transition {
	var aT []Transition
	for i := 0; i < len(s.ListT); i++ {
		if s.ListT[i].condition(s.GetListP()) && (s.ListT[i].getProbability() != 0) {
			aT = append(aT, s.ListT[i])
		}
 	}

	if len(aT) > 1 {
		aT = s.sortT(aT)
	}

	return aT
}

func (s *Simulator) GetTimeCurr() float64 {
	return s.GTimeCurr.TimeCurrentSimulation
}

func (s *Simulator) SetTimeCurr(aTimeCurr float64) {
	s.GTimeCurr.TimeCurrentSimulation = aTimeCurr
}

func (s *Simulator) isBufferEmpty() bool {
	c := true
	for _, e := range s.GetListT() {
		if e.getBuffer() > 0 {
			c = false
			break
		}
	}

	return c
}

func (s *Simulator) GetListP() []Place {
	return s.ListP
}

func (s *Simulator) GetListT() []Transition {
	return s.ListT
}

func (s *Simulator) setTimeMin(timeMin float64) {
	s.TimeMin = timeMin
}

func (s *Simulator) PrintMark() {
	log.Printf("Marks %s\n", s.GetName())
}

func (s *Simulator) DoConflict(tt []Transition) Transition {
	aT := tt[0]
	if len(tt) > 1 {
		aT = tt[0]

		i := 0
		for i < len(tt) && tt[i].GetPriority() == aT.GetPriority() {
			i++
		}

		if i > 1 {
			r := rand.Float64()
			var j int
			var sum float64
			var prob float64

			for j < len(tt) && tt[j].GetPriority() == aT.GetPriority() {
				if tt[j].getProbability() == 1.0 {
					prob = 1.0 / float64(i)
				} else {
					prob = tt[j].getProbability()
				}

				sum += prob
				if r < sum {
					aT = tt[j]
					break
				} else {
					j++
				}
			}
		}
	}

	return aT
}

func (s *Simulator) Step() {
	log.Printf("Next Step time: %f\n", s.GetTimeCurr())
	s.PrintMark()

	activeT := s.findActiveT()
	if (len(activeT) == 0 && s.isBufferEmpty()) || (s.GetTimeCurr() >= s.GetTimeMod()) {
		log.Printf("STOP in Net %s\n", s.GetName())
		s.setTimeMin(s.GetTimeMod())
		for i := 0; i < len(s.ListP); i++ {
			s.ListP[i].changeMean((s.getTimeMin() - s.GetTimeCurr()) / s.GetTimeMod())
		}

		for i := 0; i < len(s.ListT); i++ {
			s.ListT[i].changeMean((s.getTimeMin() - s.GetTimeCurr()) / s.GetTimeMod())
		}

		s.SetTimeCurr(s.getTimeMin()) // pass time
	} else {
		for len(activeT) > 0 {
			t := s.DoConflict(activeT)
			t.actIn(s.GetListP(), s.GetTimeCurr())
			activeT = s.findActiveT()
		}

		s.eventMin()

		for i := 0; i < len(s.GetListP()); i++ {
			s.ListP[i].changeMean((s.getTimeMin() - s.GetTimeCurr()) / s.GetTimeMod())
		}

		for i := 0; i < len(s.GetListT()); i++ {
			s.ListT[i].changeMean((s.getTimeMin() - s.GetTimeCurr()) / s.GetTimeMod())
		}

		s.SetTimeCurr(s.getTimeMin())

		if s.GetTimeCurr() <= s.GetTimeMod() {
			s.EventMin.actOut(s.GetListP())
			if s.EventMin.getBuffer() > 0 {
				u := true
				for u {
					s.EventMin.MinEvent()
					if s.EventMin.getMinTime() == s.GetTimeCurr() {
						s.EventMin.actOut(s.GetListP())
					} else {
						u = false
					}
				}
			}

			for i := 0; i < len(s.GetListT()); i++ {
				if s.ListT[i].getBuffer() > 0 && s.ListT[i].getMinTime() == s.GetTimeCurr() {
					s.ListT[i].actOut(s.GetListP())
					if s.ListT[i].getBuffer() > 0 {
						u := true
						for u {
							s.ListT[i].MinEvent()
							if s.ListT[i].getMinTime() == s.GetTimeCurr() {
								s.ListT[i].actOut(s.GetListP())
							} else {
								u = false
							}
						}
					}
				}
			}
		}
	}
}

func (s *Simulator) GetTimeLocal() float64 {
	return s.TimeLocal
}

func (s *Simulator) Input() {
	activeT := s.findActiveT()
	if len(activeT) == 0 && s.isBufferEmpty() {
		s.setTimeMin(math.MaxFloat64)
	} else {
		for len(activeT) > 0 {
			transition := s.DoConflict(activeT)
			transition.actIn(s.GetListP(), s.GetTimeLocal())
			activeT = s.findActiveT()
		}

		s.eventMin()
	}
}

func (s *Simulator) Contains(t []Transition, find Transition) bool {
	for i := 0; i < len(t); i++ {
		if t[i].Number == find.Number {
			return true
		}
	}

	return false
}

func (s *Simulator) GetTimeExternalInput() []float64 {
	return s.TimeExternalInput
}

func (s *Simulator) GetLimitArrayExtInputs() int {
	return s.GLimit.LimitSimulation
}

func (s *Simulator) SetLimitArrayExtInputs(limit int) {
	s.GLimit.LimitSimulation = limit
}

func (s *Simulator) AddTimeExternalInput(t float64) {
	// lock
	s.TimeExternalInput = append(s.TimeExternalInput, t)
	// unlock
}

func (s *Simulator) Output() {
	var externalPosition Place
	if s.NextObj != nil {
		externalPosition = s.ListP[len(s.ListP) - 1]
		externalPosition.setExternal(true)
	}

	for i := 0; i < len(s.GetListT()); i++ {
		if s.ListT[i].getMinTime() == s.GetTimeLocal() && s.ListT[i].getBuffer() > 0 {
			s.ListT[i].actOut(s.GetListP())
			if s.NextObj != nil && s.Contains(s.OutT, s.ListT[i]) {
				s.NextObj.AddTimeExternalInput(s.GetTimeLocal())


				// lock
				// signal
				// unlock
				log.Printf("Signal %s", s.GetName())
				s.GChannel <- 1
				// lock
				for len(s.NextObj.GetTimeExternalInput()) > s.GetLimitArrayExtInputs() {
					// await
					log.Printf("Wait %s", s.GetName())
					<-s.GChannel
				}

				// unlock
			}

			if s.ListT[i].getBuffer() > 0 {
				u := true
				for u {
					s.ListT[i].MinEvent()
					if s.ListT[i].getMinTime() == s.GetTimeLocal() {
						s.ListT[i].actOut(s.GetListP())
						if s.NextObj != nil && s.Contains(s.OutT, s.ListT[i]) {
							s.NextObj.AddTimeExternalInput(s.GetTimeLocal())

							// lock
							// signal
							// unlock
							log.Printf("Signal %s", s.GetName())
							s.GChannel <- 1
							// lock
							for len(s.NextObj.GetTimeExternalInput()) > s.GetLimitArrayExtInputs() {
								// await
								log.Printf("Wait %s", s.GetName())
								<-s.GChannel
							}

							// unlock
						}
					} else {
						u = false
					}
				}
			}
		}
	}
}

func (s *Simulator) IncreaseCounter() {
	s.Counter++
}

func (s *Simulator) DecreaseCounter() {
	s.Counter--
}

func (s *Simulator) ReinstateActOut(transition Transition, place Place) {
	for i := 0; i < len(s.PrevObj.ListOut); i++ {
		if s.PrevObj.ListOut[i].getNumT() == transition.getNumber() && s.PrevObj.ListOut[i].getNumP() == place.getNumber() {
			place.IncreaseMark(s.PrevObj.ListOut[i].getQuantity())
			break
		} else {
			log.Println("Transition reinstae act out")
		}
	}
}

func (s *Simulator) IsStop() bool {
	for i := 0; i < len(s.ListT); i++ {
		if s.ListT[i].condition(s.ListP) {
			return false
		}
		if s.ListT[i].getBuffer() > 0 {
			return false
		}
	}

	if s.PrevObj != nil {
		if len(s.TimeExternalInput) > 0 {
			return false
		}
	}

	if s.NextObj != nil {
		if len(s.NextObj.GetTimeExternalInput()) > 10 {
			return false
		}
	}

	return true
}

func (s *Simulator) StepEvent() {
	if s.IsStop() {
		s.setTimeMin(math.MaxFloat64)
	} else {
		s.Output()
		s.Input()
	}
}

func (s *Simulator) DoStatistics(dt float64) {
	if dt > 0 {
		for i := 0; i < len(s.GetListPositionsForStats()); i++ {
			s.ListPositionsForStatistica[i].changeMean(dt)
		}

		for i := 0; i < len(s.GetListT()); i++ {
			s.ListT[i].changeMean(dt)
		}
	}
}

func (s *Simulator) SetTimeLocal(tLocal float64) {
	s.TimeLocal = tLocal
}

func (s *Simulator) MoveTimeLocal(t float64) {
	s.DoStatistics((t - s.GetTimeLocal()) / t)
	s.SetTimeLocal(t)
}

func (s *Simulator) GoUntil(limitTime float64) {
	log.Println("Run go until", s.GetName())
	limit := limitTime
	for s.GetTimeLocal() < limit {
		for s.IsStop() {
			// lock
			// await
			// unlock
			log.Printf("Wait %s \n", s.GetName())
			a, ok := <-s.GChannel
			if !ok {
				log.Fatalln("ERRORRRRRRR")
			}
			log.Println(a)
			log.Println("checking is stop")
		}

		s.Input()
		if s.getTimeMin() < limit {
			s.MoveTimeLocal(s.getTimeMin())
			s.Output()
		} else {
			if limit >= s.GetTimeMod() {
				s.MoveTimeLocal(s.GetTimeMod())
				if s.NextObj != nil {
					// lock
					s.NextObj.AddTimeExternalInput(math.MaxFloat64)
					log.Printf("Signal %s \n", s.GetName())
					s.GChannel <- 1
					// signal
					// unclock
				} else {
					if s.PrevObj != nil {
						if len(s.GetTimeExternalInput()) == 0 || s.GetTimeExternalInput()[len(s.GetTimeExternalInput())- 1] < math.MaxFloat64 {
							for len(s.GetTimeExternalInput()) == 0 {
								// lock
								// await
								// unlock
								log.Printf("Wait %s \n", s.GetName())
								<-s.GChannel
							}
						}

						if s.GetTimeExternalInput()[0] > s.GetTimeMod() {
							s.MoveTimeLocal(s.GetTimeMod())
							if s.NextObj != nil {
								// lock
								s.NextObj.AddTimeExternalInput(math.MaxFloat64)
								// signal
								// unlock
								log.Printf("Signal %s \n", s.GetName())
								s.GChannel <- 1
							}
						} else {
							s.MoveTimeLocal(limit)
							s.ReinstateActOut(s.PrevObj.OutT[0], s.PrevObj.GetListP()[len(s.PrevObj.GetListP()) -1 ])
							// lock
							s.TimeExternalInput = s.GetTimeExternalInput()[1:]
							// unlock

							if len(s.GetTimeExternalInput()) <= s.GetLimitArrayExtInputs() {
								// lock prevObj
								// signal
								// unlock
								log.Printf("Signal %s \n", s.GetName())
								s.GChannel <- 1
							}
						}
					} else {
						s.MoveTimeLocal(limit)
					}
				}
			}
		}
	}
}

func (s *Simulator) Run() {
	log.Println(s.GetName(), s.NextObj != nil, s.PrevObj != nil)
	for s.GetTimeLocal() < s.GetTimeMod() {
		limitTime := s.GetTimeMod()
		if s.PrevObj != nil {
			// lock
			for len(s.GetTimeExternalInput()) == 0 {
				// await
				log.Printf("Wait %s \n", s.GetName())
				<-s.GChannel
			}

			limitTime = s.GetTimeExternalInput()[0]
			if limitTime > s.GetTimeMod() {
				limitTime = s.GetTimeMod()
			}
		} else {
			limitTime = s.GetTimeMod()
		}

		if s.GetTimeLocal() < limitTime {
			log.Printf("%s will go until %f have local time %f\n", s.GetName(), limitTime, s.GetTimeLocal())
			s.GoUntil(limitTime)
			log.Printf("%s counter: %d", s.GetName(), s.Counter)
		} else {
			return
		}
	}

	log.Printf("%s has finished simulation", s.GetName())
	s.printState()
}

func (s *Simulator) printBuffer() {
	log.Printf("%+v\n")
}

func (s *Simulator) printState() {
	s.PrintMark()
	s.printBuffer()
}

func (s *Simulator) SetTimeMod(aTimeMod float64) {
	s.GTimeMod.TimeModelingSimulation = aTimeMod
}

func (s *Simulator) GetTimeMod() float64 {
	return s.GTimeMod.TimeModelingSimulation
}
