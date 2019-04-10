package petri

import (
	"fmt"
	"log"
	"math"
	"strings"
)

type Transition struct {
	timeModeling   float64
	name           string
	buffer         int
	priority       int
	probability    float64
	minTime        float64
	timeServing    float64
	avgTimeServing float64
	avgDeviation   float64
	distribution   string

	timeout               []float64
	inPlaces              []int
	inPlacesWithInfo      []int
	counterInPlaces       []int
	counterPlacesWithInfo []int
	outPlaces             []int
	counterOutPlaces      []int

	iMultiChannel int
	number        int
	mean          float64
	observedMin   float64
	observedMax   float64
}

type BuildTransition interface {
	setTimeModeling(float64) BuildTransition
	setMean(float64) BuildTransition
	setPriority(int) BuildTransition
	setProbability(float64) BuildTransition
	setBuffer(int) BuildTransition
	setDistribution(string, float64) BuildTransition
	setDeviation(float64) BuildTransition
	setAvgTimeServing(float64) BuildTransition
	setName(string) BuildTransition
	setIMultiChannel(int) BuildTransition
	setNumber(int) BuildTransition

	generateTimeServing() float64

	addInPlace(int) BuildTransition
	addOutPlace(int) BuildTransition

	createInPlaces([]Place, []Linker) BuildTransition
	createOutPlaces([]Place, []Linker) BuildTransition

	condition([]Place) bool
	actIn([]Place, float64) BuildTransition
	actOut([]Place) BuildTransition
	initNext(*globalCounter) BuildTransition
	minEvent() BuildTransition

	print()

	clone() BuildTransition
}

func (t *Transition) build(transitionName string, timeDelay float64, probability float64, c *globalCounter) Transition {
	t.name = transitionName
	t.avgTimeServing = timeDelay
	t.avgDeviation = 0
	t.timeServing = t.avgTimeServing
	t.buffer = 0
	t.minTime = math.MaxFloat64
	t.iMultiChannel = 0
	t.mean = 0
	t.observedMax = float64(t.buffer)
	t.observedMin = float64(t.buffer)
	t.probability = probability
	t.priority = 0
	t.distribution = ""
	t.number = c.transition
	c.transition++
	t.timeout = append(t.timeout, math.MaxFloat64)
	t.minEvent()

	return *t
}

func (t *Transition) setTimeModeling(m float64) BuildTransition {
	t.timeModeling = m
	return t
}

func (t *Transition) initNext(c *globalCounter) BuildTransition {
	c.transition = 0
	return t
}

func (t *Transition) setMean(m float64) BuildTransition {
	t.mean += (float64(t.buffer) - t.mean) * m
	return t
}

func (t *Transition) setPriority(p int) BuildTransition {
	t.priority = p
	return t
}

func (t *Transition) setProbability(p float64) BuildTransition {
	t.probability = p
	return t
}

func (t *Transition) setBuffer(b int) BuildTransition {
	t.buffer = b
	return t
}

func (t *Transition) setDistribution(d string, param float64) BuildTransition {
	t.distribution = d
	t.avgTimeServing = param
	t.timeServing = t.avgTimeServing
	return t
}

func (t *Transition) setAvgTimeServing(v float64) BuildTransition {
	t.avgTimeServing = v
	t.timeServing = t.avgTimeServing
	return t
}

func (t *Transition) setDeviation(v float64) BuildTransition {
	t.avgDeviation = v
	return t
}

func (t *Transition) setIMultiChannel(v int) BuildTransition {
	t.iMultiChannel = v
	return t
}

func (t *Transition) setNumber(v int) BuildTransition {
	t.number = v
	return t
}

func (t *Transition) generateTimeServing() float64 {
	if t.distribution != "" {
		switch strings.ToLower(t.distribution) {
		case "exp":
			t.timeServing = Exp(t.avgTimeServing)
			break
		case "unif":
			t.timeServing = Uniform(t.avgTimeServing-t.avgDeviation, t.avgTimeServing+t.avgDeviation)
			break
		case "norm":
			t.timeServing = Normal(t.avgTimeServing, t.avgDeviation)
			break
		}
	} else {
		t.timeServing = t.avgTimeServing
	}

	return t.timeServing
}

func (t *Transition) setName(n string) BuildTransition {
	t.name = n
	return t
}

func (t *Transition) setMultiChannel(m int) BuildTransition {
	t.iMultiChannel = m
	return t
}

func (t *Transition) setTransition(v int) BuildTransition {
	t.number = v
	return t
}

func (t *Transition) addInPlace(n int) BuildTransition {
	t.inPlaces = append(t.inPlaces, n)
	return t
}

func (t *Transition) addOutPlace(n int) BuildTransition {
	t.outPlaces = append(t.outPlaces, n)
	return t
}

func (t *Transition) createInPlaces(places []Place, links []Linker) BuildTransition {
	t.inPlacesWithInfo = t.inPlacesWithInfo[:0]
	t.counterPlacesWithInfo = t.counterPlacesWithInfo[:0]
	t.inPlaces = t.inPlaces[:0]
	t.counterInPlaces = t.counterInPlaces[:0]

	for _, link := range links {
		if float64(link.counterTransitions) == t.avgTimeServing {
			if link.isInfo() {
				t.inPlacesWithInfo = append(t.inPlacesWithInfo, link.getCounterPlaces())
				t.counterPlacesWithInfo = append(t.counterPlacesWithInfo, link.getQuantity())
			} else {
				t.inPlaces = append(t.inPlaces, link.getCounterPlaces())
				t.counterInPlaces = append(t.counterInPlaces, link.getQuantity())
			}
		}
	}

	if len(t.inPlaces) == 0 {
		log.Fatalln(fmt.Errorf("transition %s hasn't input positions", t.name))
	}

	return t
}

func (t *Transition) createOutPlaces(places []Place, links []Linker) BuildTransition {
	t.outPlaces = t.outPlaces[:0]
	t.counterOutPlaces = t.counterOutPlaces[:0]

	for _, link := range links {
		if float64(link.getCounterTransitions()) == t.avgTimeServing {
			t.outPlaces = append(t.outPlaces, link.getCounterPlaces())
			t.counterOutPlaces = append(t.counterOutPlaces, link.getQuantity())
		}
	}

	if len(t.outPlaces) == 0 {
		log.Fatalln(fmt.Errorf("transition %s hasn't input positions", t.name))
	}

	return t
}

func (t *Transition) condition(places []Place) bool {
	var a = true
	var b = true

	for i, place := range t.inPlaces {
		if places[place].getMark() < float64(t.counterInPlaces[i]) {
			a = false
			break
		}
	}

	for i, place := range t.inPlacesWithInfo {
		if places[place].getMark() < float64(t.counterPlacesWithInfo[i]) {
			b = false
			break
		}
	}

	return a == true && b == true
}

func (t *Transition) actIn(places []Place, currentTime float64) BuildTransition {
	if t.condition(places) {
		for i, place := range t.inPlaces {
			places[place].decrMark(float64(t.counterInPlaces[i]))
		}

		if t.buffer == 0 {
			t.timeout = make([]float64, 1)
			t.timeout[0] = currentTime + t.timeServing
		} else {
			t.timeout = append(t.timeout, currentTime+t.timeServing)
		}

		t.buffer++
		if t.observedMax < float64(t.buffer) {
			t.observedMax = float64(t.buffer)
		}

		t.minEvent()
	} else {
		log.Println("Condition not true")
	}

	return t
}

func (t *Transition) actOut(places []Place) BuildTransition {
	if t.buffer > 0 {
		for i, place := range t.outPlaces {
			if !places[place].isExternal() {
				places[place].incrMark(float64(t.counterOutPlaces[i]))
			}
		}

		if t.iMultiChannel == 0 && len(t.timeout) == 1 {
			t.timeout[0] = math.MaxFloat64
		} else {
			t.timeout = append(t.timeout[:t.iMultiChannel], t.timeout[t.iMultiChannel+1:]...)
		}

		t.buffer--
		if t.observedMin > float64(t.buffer) {
			t.observedMin = float64(t.buffer)
		}

	}

	return t
}

func (t *Transition) minEvent() BuildTransition {
	var minTime = math.MaxFloat64
	if len(t.timeout) > 0 {
		for i, timeout := range t.timeout {
			if timeout < minTime {
				minTime = t.timeout[i]
				t.iMultiChannel = i
			}
		}
	}

	return t
}

func (t *Transition) print() {
	fmt.Printf("%+v", t)
}

func (t *Transition) clone() BuildTransition {
	var n Transition
	n = *t
	n.timeout = t.timeout[:]
	n.inPlaces = t.inPlaces[:]
	n.inPlacesWithInfo = t.inPlacesWithInfo[:]
	n.counterInPlaces = t.counterInPlaces[:]
	n.counterPlacesWithInfo = t.counterPlacesWithInfo[:]
	n.outPlaces = t.outPlaces[:]
	n.counterOutPlaces = t.counterOutPlaces[:]
	return &n
}
