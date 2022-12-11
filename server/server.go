package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type EvaluateNextBoard struct{}

func (e *EvaluateNextBoard) Evaluate(request stubs.Request, response *stubs.WorkerResponse) (err error) {

	// send the request current board state
	oldWorld := request.OldWorld

	nextWorld := makeNewWorld(request.ImageHeight, request.ImageWidth)

	for h := 0; h < request.ImageHeight; h++ {
		for w := 0; w < request.ImageWidth; w++ {
			nextWorld[h][w] = oldWorld[h][w]
		}
	}

	response.NewWorld = makeNewWorld(request.ImageHeight, request.ImageWidth)

	if request.Turns == 0 {
		response.NewWorld = oldWorld
	}

	if request.Turns > 0 {
		for turn := 0; turn < request.Turns; turn++ {
			nextWorld = workerRoutine(request.ImageHeight, 0, request.ImageWidth, nextWorld)
		}
		response.NewWorld = nextWorld
	}

	return
}

func workerRoutine(endY, StartX, endX int, oldWorld [][]uint8) [][]uint8 {
	newWorld := makeNewWorld(endY, endX)

	for h := 0; h < endY; h++ {
		for w := StartX; w < endX; w++ {
			live := CalcLiveNeighbor(oldWorld, endY, endX, w, h)
			if oldWorld[h][w] == 255 {
				if live < 2 {
					newWorld[h][w] = 0
				}
				if live == 2 || live == 3 {
					newWorld[h][w] = 255
				}
				if live > 3 {
					newWorld[h][w] = 0
				}
			}

			if oldWorld[h][w] == 0 {
				if live == 3 {
					newWorld[h][w] = 255
				}
			}
		}
	}
	return newWorld
}

func CalcLiveNeighbor(world [][]uint8, h int, w int, x int, y int) int {
	liveSum := 0

	//x-1, y-1
	//(x-1+w)%w, (y-1+h)%h
	if world[(y-1+h)%h][(x-1+w)%w] == 255 {
		liveSum++
	}

	//x, y-1
	//x, (y-1+h)%h
	if world[(y-1+h)%h][x] == 255 {
		liveSum++
	}

	//x+1, y-1
	//(x+1)%w, (y-1+h)%h
	if world[(y-1+h)%h][(x+1)%w] == 255 {
		liveSum++
	}

	//x-1, y
	//(x-1+w)%w, y
	if world[y][(x-1+w)%w] == 255 {
		liveSum++
	}

	//x+1, y
	//(x+1)%w, y
	if world[y][(x+1)%w] == 255 {
		liveSum++
	}

	//x-1, y+1
	//(x-1+w)%w, (y+1)%h
	if world[(y+1)%h][(x-1+w)%w] == 255 {
		liveSum++
	}

	//x, y+1
	//x, (y+1)%h
	if world[(y+1)%h][x] == 255 {
		liveSum++
	}

	//x+1, y+1
	//(x+1)%w, (y+1)%h
	if world[(y+1)%h][(x+1)%w] == 255 {
		liveSum++
	}

	return liveSum
}

func makeNewWorld(height, width int) [][]uint8 {
	newWorld := make([][]uint8, height)
	for i := range newWorld {
		newWorld[i] = make([]uint8, width)
	}
	return newWorld
}

func main() {
	portAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()

	err := rpc.Register(&EvaluateNextBoard{})
	if err != nil {
		fmt.Println("Error in RPC Register")
	}

	listener, _ := net.Listen("tcp", ":"+*portAddr)
	defer listener.Close()

	rpc.Accept(listener)
}
