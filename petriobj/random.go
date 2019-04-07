package petriobj

import (
	"fmt"
	"math"
	"math/rand"
)

func generate() (r float64) {
	for r == 0 {
		r = rand.Float64()
	}

	return r
}

func Exp(timeMean float64) float64 {
	return -timeMean * math.Log(generate())
}

func Uniform(timeMin float64, timeMax float64) float64 {
	return timeMin + generate()*(timeMax-timeMin)
}

func Normal(timeMean float64, timeDeviation float64) float64 {
	return timeMean + timeDeviation*rand.NormFloat64()
}

func Empiric(x []float64, y []float64) (v float64, err error) {
	n := len(x) - 1
	if y[n] != 1.0 {
		return v, fmt.Errorf("illegal array of points for empiric distribution")
	}

	r := rand.Float64()
	for i := 1; i < len(x)-1; i++ {
		if r > y[i-1] && r <= y[i] {
			return x[i-1] + (r-y[i-1])*(x[i]-x[i-1])/(y[i]-y[i-1]), nil
		}
	}

	return x[n-2] + (r-y[n-2])*(x[n-1]-x[n-2])/(y[n-1]-y[n-2]), nil
}
