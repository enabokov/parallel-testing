package petri

import (
	"fmt"
	"log"
	"math"
	"strings"
)

type Transition struct {
	TimeModeling   float64
	Name           string
	Buffer         int
	Priority       int
	Probability    float64
	MinTime        float64
	TimeServing    float64
	AvgTimeServing float64
	AvgDeviation   float64
	Distribution   string // possible names: unif, exp, norm

	Timeout               []float64
	InPlaces              []int
	InPlacesWithInfo      []int
	CounterInPlaces       []int
	CounterPlacesWithInfo []int
	OutPlaces             []int
	CounterOutPlaces      []int

	IMultiChannel int
	Number        int
	Mean          float64
	ObservedMin   float64
	ObservedMax   float64
}

type BuildTransition interface {
	SetTimeModeling(float64) BuildTransition
	SetMean(float64) BuildTransition
	SetPriority(int) BuildTransition
	SetProbability(float64) BuildTransition
	SetBuffer(int) BuildTransition
	SetDistribution(string, float64) BuildTransition
	SetDeviation(float64) BuildTransition
	SetAvgTimeServing(float64) BuildTransition
	SetName(string) BuildTransition
	SetIMultiChannel(int) BuildTransition
	SetNumber(int) BuildTransition

	GenerateTimeServing() float64

	AddInPlace(int) BuildTransition
	AddOutPlace(int) BuildTransition

	CreateInPlaces([]*Place, []*Linker) BuildTransition
	CreateOutPlaces([]*Place, []*Linker) BuildTransition

	Condition([]*Place) bool
	ActIn([]*Place, float64) BuildTransition
	ActOut([]*Place) BuildTransition
	InitNext(*GlobalCounter) BuildTransition
	MinEvent() BuildTransition

	Print()

	Clone() BuildTransition
}

func (t *Transition) Build(transitionName string, timeDelay float64, probability float64, c *GlobalCounter) *Transition {
	t.Name = transitionName
	t.AvgTimeServing = timeDelay
	t.AvgDeviation = 0
	t.TimeServing = t.AvgTimeServing
	t.Buffer = 0
	t.MinTime = math.MaxFloat64
	t.IMultiChannel = 0
	t.Mean = 0
	t.TimeModeling = math.MaxFloat64 - 1
	t.ObservedMax = float64(t.Buffer)
	t.ObservedMin = float64(t.Buffer)
	t.Probability = probability
	t.Priority = 0
	t.Distribution = ""
	t.Number = c.Transition
	c.Transition++
	t.Timeout = append(t.Timeout, math.MaxFloat64)
	t.MinEvent()

	return t
}

func (t *Transition) SetTimeModeling(m float64) BuildTransition {
	t.TimeModeling = m
	return t
}

func (t *Transition) InitNext(c *GlobalCounter) BuildTransition {
	c.Transition = 0
	return t
}

func (t *Transition) SetMean(m float64) BuildTransition {
	t.Mean += (float64(t.Buffer) - t.Mean) * m
	return t
}

func (t *Transition) SetPriority(p int) BuildTransition {
	t.Priority = p
	return t
}

func (t *Transition) SetProbability(p float64) BuildTransition {
	t.Probability = p
	return t
}

func (t *Transition) SetBuffer(b int) BuildTransition {
	t.Buffer = b
	return t
}

func (t *Transition) SetDistribution(d string, param float64) BuildTransition {
	t.Distribution = d
	t.AvgTimeServing = param
	t.TimeServing = t.AvgTimeServing
	return t
}

func (t *Transition) SetAvgTimeServing(v float64) BuildTransition {
	t.AvgTimeServing = v
	t.TimeServing = t.AvgTimeServing
	return t
}

func (t *Transition) SetDeviation(v float64) BuildTransition {
	t.AvgDeviation = v
	return t
}

func (t *Transition) SetIMultiChannel(v int) BuildTransition {
	t.IMultiChannel = v
	return t
}

func (t *Transition) SetNumber(v int) BuildTransition {
	t.Number = v
	return t
}

func (t *Transition) GenerateTimeServing() float64 {
	if t.Distribution != "" {
		switch strings.ToLower(t.Distribution) {
		case "exp":
			t.TimeServing = Exp(t.AvgTimeServing)
			break
		case "unif":
			t.TimeServing = Uniform(t.AvgTimeServing-t.AvgDeviation, t.AvgTimeServing+t.AvgDeviation)
			break
		case "norm":
			t.TimeServing = Normal(t.AvgTimeServing, t.AvgDeviation)
			break
		}
	} else {
		t.TimeServing = t.AvgTimeServing
	}

	return t.TimeServing
}

func (t *Transition) SetName(n string) BuildTransition {
	t.Name = n
	return t
}

func (t *Transition) setMultiChannel(m int) BuildTransition {
	t.IMultiChannel = m
	return t
}

func (t *Transition) setTransition(v int) BuildTransition {
	t.Number = v
	return t
}

func (t *Transition) AddInPlace(n int) BuildTransition {
	t.InPlaces = append(t.InPlaces, n)
	return t
}

func (t *Transition) AddOutPlace(n int) BuildTransition {
	t.OutPlaces = append(t.OutPlaces, n)
	return t
}

func (t *Transition) CreateInPlaces(places []*Place, links []*Linker) BuildTransition {
	t.InPlacesWithInfo = []int{}
	t.CounterPlacesWithInfo = []int{}
	t.InPlaces = []int{}
	t.CounterInPlaces = []int{}

	for i := 0; i < len(links); i++ {
		if links[i].CounterTransitions == t.Number {
			if links[i].IsInfo() {
				t.InPlacesWithInfo = append(t.InPlacesWithInfo, links[i].GetCounterPlaces())
				t.CounterPlacesWithInfo = append(t.CounterPlacesWithInfo, links[i].GetQuantity())
			} else {
				t.InPlaces = append(t.InPlaces, links[i].GetCounterPlaces())
				t.CounterInPlaces = append(t.CounterInPlaces, links[i].GetQuantity())
			}
		}
	}

	if len(t.InPlaces) == 0 {
		log.Println(fmt.Errorf("transition %s hasn't input positions", t.Name))
	}

	return t
}

func (t *Transition) CreateOutPlaces(places []*Place, links []*Linker) BuildTransition {
	t.OutPlaces = []int{}
	t.CounterOutPlaces = []int{}

	for i := 0; i < len(links); i++ {
		if links[i].CounterTransitions == t.Number {
			t.OutPlaces = append(t.OutPlaces, links[i].GetCounterPlaces())
			t.CounterOutPlaces = append(t.CounterOutPlaces, links[i].GetQuantity())
		}
	}

	if len(t.OutPlaces) == 0 {
		log.Println(fmt.Errorf("transition %s hasn't output positions", t.Name))
	}

	return t
}

func (t *Transition) Condition(places []*Place) bool {
	var a = true
	var b = true

	for i := 0; i < len(t.InPlaces); i++ {
		if places[t.InPlaces[i]].GetMark() < float64(t.CounterInPlaces[i]) {
			a = false
			break
		}
	}

	for i := 0; i < len(t.InPlacesWithInfo); i++ {
		if places[t.InPlacesWithInfo[i]].GetMark() < float64(t.CounterPlacesWithInfo[i]) {
			b = false
			break
		}
	}

	return a == true && b == true
}

func (t *Transition) ActIn(places []*Place, currentTime float64) BuildTransition {
	if t.Condition(places) {
		for i := 0; i < len(t.InPlaces); i++ {
			places[t.InPlaces[i]].DecrMark(float64(t.CounterInPlaces[i]))
		}

		if t.Buffer == 0 {
			t.Timeout[0] = currentTime + t.TimeServing
		} else {
			t.Timeout = append(t.Timeout, currentTime+t.TimeServing)
		}

		t.Buffer++
		if t.ObservedMax < float64(t.Buffer) {
			t.ObservedMax = float64(t.Buffer)
		}

		t.MinEvent()
	} else {
		log.Println("Condition not true")
	}

	return t
}

func (t *Transition) ActOut(places []*Place) BuildTransition {
	if t.Buffer > 0 {
		for i := 0; i < len(t.OutPlaces); i++ {
			if !places[t.OutPlaces[i]].IsExternal() {
				places[t.OutPlaces[i]].IncrMark(float64(t.CounterOutPlaces[i]))
			}
		}

		if t.IMultiChannel == 0 && len(t.Timeout) == 1 {
			t.Timeout[0] = math.MaxFloat64
		} else {
			t.Timeout = append(t.Timeout[:t.IMultiChannel], t.Timeout[t.IMultiChannel+1:]...)
		}

		t.Buffer--
		if t.ObservedMin > float64(t.Buffer) {
			t.ObservedMin = float64(t.Buffer)
		}

	}

	return t
}

func (t *Transition) MinEvent() BuildTransition {
	var minTime = math.MaxFloat64
	if len(t.Timeout) > 0 {
		for i := 0; i < len(t.Timeout); i++ {
			if t.Timeout[i] < minTime {
				minTime = t.Timeout[i]
				t.IMultiChannel = i
			}
		}
	}

	return t
}

func (t *Transition) Print() {
	for i := 0; i < len(t.Timeout); i++ {
		log.Printf("%f %s\n", t.Timeout[i], t.Name)
	}
}

func (t *Transition) Clone() BuildTransition {
	var n Transition
	n = *t
	n.Timeout = t.Timeout[:]
	n.InPlaces = t.InPlaces[:]
	n.InPlacesWithInfo = t.InPlacesWithInfo[:]
	n.CounterInPlaces = t.CounterInPlaces[:]
	n.CounterPlacesWithInfo = t.CounterPlacesWithInfo[:]
	n.OutPlaces = t.OutPlaces[:]
	n.CounterOutPlaces = t.CounterOutPlaces[:]
	return &n
}
