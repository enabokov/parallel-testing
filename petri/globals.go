package petri

type GNext struct {
	NextArcIn int
	NextArcOut int
	NextPlace int
	NextTransition int
	NextSimulator int
}

type GTimeModeling struct {
	TimeModelingTransition float64
	TimeModelingSimulation float64
}

type GLimitArrayExtInputs struct {
	LimitSimulation int
}

type GTimeCurrent struct {
	TimeCurrentSimulation float64
}

type GChannel struct {
	ChannelSimulation chan int
}
