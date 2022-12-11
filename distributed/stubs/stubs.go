package stubs

var Evaluate = "EvaluateNextBoard.Evaluate"

type Request struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
	OldWorld    [][]uint8
}

type WorkerResponse struct {
	NewWorld [][]uint8
}
