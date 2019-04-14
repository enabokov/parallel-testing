package petri

import (
	"fmt"
	"log"
	"math"
	"strings"
)

type Transition struct {
	// golang has no static vars
	// thus it refers to global vars
	GTimeMod *GTimeModeling

	Name string

	Buffer int
	Priority int
	Probability float64

	MinTime float64
	TimeServ float64

	// avg time serving
	Parametr float64

	// avg deviation time serving
	ParamDeviation float64
	Distribution string
	TimeOut []float64
	InP []int
	InPWithInf []int
	QuantIn []int
	QuantInWithInf []int
	OutP []int
	QuantOut []int

	Num int
	Number int
	Mean float64
	ObservedMax int
	ObservedMin int
}

func (t *Transition) GetTimeModeling() float64 {
	return t.GTimeMod.TimeModelingTransition
}

func (t *Transition) SetTimeModeling(aTimeModeling float64) {
	t.GTimeMod.TimeModelingTransition = aTimeModeling
}

func (t *Transition) Build(n string, ts float64, gnext *GNext, gtimeModeling *GTimeModeling) *Transition {
	t.GTimeMod = gtimeModeling

	t.Name = n
	t.Parametr = ts
	t.ParamDeviation = 0
	t.TimeServ = t.Parametr
	t.Buffer = 0

	t.MinTime = math.MaxFloat64
	t.Num = 0
	t.Mean = 0
	t.ObservedMin = t.Buffer
	t.ObservedMax = t.Buffer
	t.Priority = 0
	t.Probability = 1.0
	t.Distribution = ""

	t.Number = gnext.NextTransition
	gnext.NextTransition++

	t.TimeOut = append(t.TimeOut, math.MaxFloat64)
	t.MinEvent()
	return t
}

func (t *Transition) BuildWithPriority(n string, timeDelay float64, priority int, gnext *GNext, gtimeModeling *GTimeModeling) *Transition {
	t.GTimeMod = gtimeModeling

	t.Name = n
	t.Parametr = timeDelay
	t.ParamDeviation = 0
	t.TimeServ = t.Parametr
	t.Buffer = 0

	t.MinTime = math.MaxFloat64
	t.Num = 0
	t.Mean = 0
	t.ObservedMax = t.Buffer
	t.ObservedMin = t.Buffer
	t.Priority = priority
	t.Probability = 1
	t.Distribution = ""

	t.Number = gnext.NextTransition
	gnext.NextTransition++

	t.TimeOut = append(t.TimeOut, math.MaxFloat64)
	t.MinEvent()
	return t
}

func (t *Transition) BuildWithProbability(n string, timeDelay float64, probability float64, gnext *GNext, gtimeModeling *GTimeModeling) *Transition {
	t.GTimeMod = gtimeModeling

	t.Name = n
	t.Parametr = timeDelay
	t.ParamDeviation = 0
	t.TimeServ = t.Parametr
	t.Buffer = 0

	t.MinTime = math.MaxFloat64
	t.Num = 0
	t.Mean = 0
	t.ObservedMax = t.Buffer
	t.ObservedMin = t.Buffer
	t.Priority = 0
	t.Distribution = ""

	t.Number = gnext.NextTransition
	gnext.NextTransition++

	t.TimeOut = append(t.TimeOut, math.MaxFloat64)
	t.MinEvent()
	return t
}

func (t *Transition) changeMean(a float64) {
	t.Mean += (float64(t.Buffer) - t.Mean) * a
}

func (t *Transition) getMean() float64 {
	return t.Mean
}

func (t *Transition) getObservedMax() int {
	return t.ObservedMax
}

func (t *Transition) getObservedMin() int {
	return t.ObservedMin
}

func (t *Transition) GetPriority() int {
	return t.Priority
}

func (t *Transition) setPriority(r int) {
	t.Priority = r
}

func (t *Transition) getProbability() float64 {
	return t.Probability
}

func (t *Transition) setProbability(v float64) {
	t.Probability = v
}

func (t *Transition) getBuffer() int {
	return t.Buffer
}

func (t *Transition) setDistribution(s string, param float64) {
	t.Distribution = s
	t.Parametr = param
	t.TimeServ = t.Parametr
}

func (t *Transition) getTimeServing() float64 {
	var a = t.TimeServ
	if t.Distribution != "" {
		a = t.generateTimeServ()
	}

	return a
}

func (t *Transition) getParametr() float64 {
	return t.Parametr
}

func (t *Transition) setParametr(p float64) {
	t.Parametr = p
	t.TimeServ = t.Parametr
}

func (t *Transition) generateTimeServ() float64 {
	if t.Distribution != "" {
		if strings.ToLower(t.Distribution) == "exp" {
			t.TimeServ = Exp(t.Parametr)
		} else if strings.ToLower(t.Distribution) == "unif" {
			t.TimeServ = Uniform(t.Parametr - t.ParamDeviation, t.Parametr + t.ParamDeviation)
		} else if strings.ToLower(t.Distribution) == "norm" {
			t.TimeServ = Normal(t.Parametr, t.ParamDeviation)
		}
	} else {

		// determined value
		t.TimeServ = t.Parametr
	}

	return t.TimeServ
}

func (t *Transition) GetName() string {
	return t.Name
}

func (t *Transition) setName(s string) {
	t.Name = s
}

func (t *Transition) getMinTime() float64 {
	t.MinEvent()
	return t.MinTime
}

func (t *Transition) getNum() int {
	return t.Num
}

func (t *Transition) getNumber() int {
	return t.Number
}

func (t *Transition) addInP(n int) {
	t.InP = append(t.InP, n)
}

func (t *Transition) addOutP(n int) {
	t.OutP = append(t.OutP, n)
}

func (t *Transition) createInP(inPP []Place, ties []ArcIn) {
	// remove elements, leave allocated mem
	t.InPWithInf = t.InPWithInf[:0]
	t.QuantInWithInf = t.QuantInWithInf[:0]
	t.InP = t.InP[:0]
	t.QuantIn = t.QuantIn[:0]

	for i := 0; i < len(ties); i++ {
		if ties[i].getNumT() == t.getNumber() {
			if ties[i].getIsInf() {
				t.InPWithInf = append(t.InPWithInf, ties[i].getNumP())
				t.QuantInWithInf = append(t.QuantInWithInf, ties[i].getQuantity())
			} else {
				t.InP = append(t.InP, ties[i].getNumP())
				t.QuantIn = append(t.QuantIn, ties[i].getQuantity())
			}
		}
	}

	if len(t.InP) == 0 {
		log.Println(fmt.Errorf("transition %s hasn't input positions", t.GetName()))
	}
}

func (t *Transition) createOutP(inPP []Place, arcs []ArcOut) {
	t.OutP = t.OutP[:0]
	t.QuantOut = t.QuantOut[:0]

	for i := 0; i < len(arcs); i++ {
		if arcs[i].getNumT() == t.getNumber() {
			t.OutP = append(t.OutP, arcs[i].getNumP())
			t.QuantOut = append(t.QuantOut, arcs[i].getQuantity())
		}
	}

	if len(t.OutP) == 0 {
		log.Println(fmt.Errorf("transition %s hasn't output positions", t.GetName()))
	}
}

func (t *Transition) setNum(n int) {
	t.Num = n
}

func (t *Transition) setNumber(n int) {
	t.Number = n
}

func (t *Transition) condition(pp []Place) bool {
	var a = true
	var b = true

	for i := 0; i < len(t.InP); i++ {
		if pp[t.InP[i]].getMark() < t.QuantIn[i] {
			a = false
			break
		}
	}

	for i := 0; i < len(t.InPWithInf); i++ {
		if pp[t.InPWithInf[i]].getMark() < t.QuantInWithInf[i] {
			b = false
			break
		}
	}

	return a == true && b == true
}

func (t *Transition) actIn(pp []Place, currentTime float64) {
	if t.condition(pp) {
		for i := 0; i < len(t.InP); i++ {
			pp[t.InP[i]].decreaseMark(t.QuantIn[i])
		}
		if t.Buffer == 0 {
			t.TimeOut[0] = currentTime + t.getTimeServing()
		} else {
			t.TimeOut = append(t.TimeOut, currentTime + t.getTimeServing())
		}

		t.Buffer++
		if t.ObservedMax < t.Buffer {
			t.ObservedMax = t.Buffer
		}

		t.MinEvent()
	} else {
		log.Println("Condition not true")
	}
}

func (t *Transition) actOut(pp []Place) {
	if t.Buffer > 0 {
		for i := 0; i < len(t.OutP); i++ {
			if !pp[t.OutP[i]].isExternal() {
				pp[t.OutP[i]].IncreaseMark(t.QuantOut[i])
			}
		}
		if t.Num == 0 && len(t.TimeOut) == 1 {
			t.TimeOut[0] = math.MaxFloat64
		} else {
			t.TimeOut = append(t.TimeOut[:t.Num], t.TimeOut[t.Num+1:]...)
		}

		t.Buffer--
		if t.ObservedMin > t.Buffer {
			t.ObservedMin = t.Buffer
		}
	} else {
		log.Println("Buffer is null")
	}
}

func (t *Transition) MinEvent() {
	t.MinTime = math.MaxFloat64
	if len(t.TimeOut) > 0 {
		for i := 0; i < len(t.TimeOut); i++ {
			if t.TimeOut[i] < t.MinTime {
				t.MinTime = t.TimeOut[i]
				t.Num = i
			}
		}
	}
}

func (t *Transition) print() {
	for _, time := range t.TimeOut {
		log.Printf("%f %s", time, t.GetName())
	}
}

func (t *Transition) printParameters() {
	log.Printf("%+v", t)
	log.Printf("Time of service (generate) %f\n", t.getTimeServing())
}

func (t *Transition) getInP() []int {
	return t.InP
}

func (t *Transition) getOutP() []int {
	return t.OutP
}

func (t *Transition) isEmptyInputPlacesList() bool {
	return len(t.InP) == 0
}

func (t *Transition) isEmptyOutputPlacesList() bool {
	return len(t.OutP) == 0
}

func (t *Transition) setBuffer(buff int) {
	t.Buffer = buff
}

func (t *Transition) getDistribution() string {
	return t.Distribution
}

func (t *Transition) getParamDeviation() float64 {
	return t.ParamDeviation
}

func (t *Transition) setParamDeviation(parameter float64) {
	t.ParamDeviation = parameter
}
