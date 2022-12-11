package gol

import (
	"fmt"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	inputWorld := makeNewWorld(p.ImageHeight, p.ImageWidth)

	c.ioCommand <- ioInput

	var inputFilename string
	inputFilename = fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	c.ioFilename <- inputFilename

	// get input image from the ioInput
	for h := 0; h < p.ImageHeight; h++ {
		for w := 0; w < p.ImageWidth; w++ {
			inputWorld[h][w] = <-c.ioInput
			cell := util.Cell{
				X: w,
				Y: h,
			}
			if inputWorld[h][w] == 255 {
				c.events <- CellFlipped{
					CompletedTurns: 0,
					Cell:           cell,
				}
			}
		}
	}

	turn := 0

	// TODO: Execute all turns of the Game of Life.

	client, _ := rpc.Dial("tcp", "54.88.56.104:8030")
	defer client.Close()

	request := stubs.Request{
		Turns:       p.Turns,
		Threads:     p.Threads,
		ImageWidth:  p.ImageWidth,
		ImageHeight: p.ImageHeight,
		OldWorld:    inputWorld,
	}

	response := new(stubs.WorkerResponse)

	err := client.Call(stubs.Evaluate, request, response)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	var liveCell []util.Cell
	for h := 0; h < p.ImageHeight; h++ {
		for w := 0; w < p.ImageWidth; w++ {
			c.ioOutput <- response.NewWorld[h][w]

			if response.NewWorld[h][w] == 255 {
				cell := util.Cell{
					X: w,
					Y: h,
				}
				liveCell = append(liveCell, cell)
			}
		}
	}

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          liveCell,
	}

	c.ioCommand <- ioOutput
	c.ioFilename <- inputFilename + fmt.Sprintf("x%d", p.Turns)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

func makeNewWorld(height, width int) [][]uint8 {
	newWorld := make([][]uint8, height)
	for i := range newWorld {
		newWorld[i] = make([]uint8, width)
	}
	return newWorld
}
