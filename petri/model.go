package petri

import (
	"math"
	"math/rand"
	"sort"
)

type Model struct {
	ListObj []Simulator
	TimeMod float64
	T float64
	IsStatistica bool
}

func (m *Model) Build(list []Simulator) {
	m.TimeMod = math.MaxFloat64 - 1
	m.ListObj = list
	m.IsStatistica = true
}

func (m *Model) SetIsStatistica(b bool) {
	m.IsStatistica = b
}

func (m *Model) GetListObj() []Simulator {
	return m.ListObj
}

func (m *Model) SetListObj(list []Simulator) {
	m.ListObj = list
}

func (m *Model) getNextEventTime() float64 {
	min := m.ListObj[0].GetTimeMod()
	for i := 0; i < len(m.ListObj); i++ {
		if m.ListObj[i].getTimeMin() < min {
			min = m.ListObj[i].getTimeMin()
		}
	}

	return min
}

func (m *Model) SortObjParallel(tt []Simulator) []Simulator {
	var tmp []Simulator
	for i := 0; i < len(tt); i++ {
		tmp = append(tmp, tt[i])
	}

	if len(tt) > 1 {
		sort.Slice(tmp, func(i, j int) bool {
			return tmp[i].GetPriority() < tmp[j].GetPriority()
		})
	}

	tt = tt[:0]
	for i := 0; i < len(tmp); i++ {
		tt = append(tt, tmp[i])
	}

	return tt
}

func (m *Model) ChooseObj(array []Simulator) Simulator {
	var (
		num int
		max int
	)

	if len(array) > 1 {
		max = len(array)
		m.SortObjParallel(array)

		for i := 1; i < len(array); i++ {
			if array[i].GetPriority() < array[i - 1].GetPriority() {
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

	return array[num]
}

func (m *Model) GetTimeMod() float64 {
	return m.TimeMod
}

func (m *Model) SetTimeMod(timeMod float64) {
	m.TimeMod = timeMod
}
