package main

import (
	aocshared "aoc_shared"
	"fmt"
	"slices"
	"strings"
)

type CellType int

const (
	SPACE CellType = iota
	SPLITTER
	BEAM
)

type Vector2 struct {
	X int
	Y int
}

func main() {
	aocshared.DebugAndLogTask("2025 Day 7", solve)
}

func solve() {
	input := aocshared.GetInput(2025, 7)
	lines := strings.Split(input, "\n")

	rows := make([][]CellType, len(lines))
	cellsPerRow := len(lines[0])

	initialBeamPosition := Vector2{X: -1, Y: -1}
	for y, line := range lines {
		cellsInRow := make([]CellType, cellsPerRow)

		for x, char := range line {
			cellType := getCellType(char)

			if cellType == BEAM {
				initialBeamPosition.X = x
				initialBeamPosition.Y = y
			}

			cellsInRow[x] = cellType
		}

		rows[y] = cellsInRow
	}

	grid := aocshared.GridCreate(rows)
	part2Grid := aocshared.GridClone(grid)

	splitTimes := getSplitTimes(grid, initialBeamPosition)
	activeTimelines := getActiveTimelines(part2Grid, initialBeamPosition)

	fmt.Printf("Split Times (P1): %d\n", splitTimes)
	fmt.Printf("Active Timelines (P2): %d\n", activeTimelines)
}

func debugGrid(grid *aocshared.Grid[CellType]) {
	aocshared.GridDebug(*grid, func(cell CellType) string {
		switch cell {
		case SPACE:
			return "."
		case SPLITTER:
			return "^"
		case BEAM:
			return "|"
		default:
			panic(fmt.Errorf("unknown character: '%v'", cell))
		}
	})
}

func getSplitTimes(grid *aocshared.Grid[CellType], initialPosition Vector2) int {
	splitTimes := 0

	currentBeamPositionsToBeProcessed := []Vector2{initialPosition}

	for i := 0; len(currentBeamPositionsToBeProcessed) != 0; i++ {
		// fmt.Printf("Iteration: %d (START) - grid:\n", i)
		// debugGrid(grid)
		// fmt.Printf("-----------------\n")

		beamsInQueue := len(currentBeamPositionsToBeProcessed)
		for _, beamInQueue := range currentBeamPositionsToBeProcessed {
			southX, southY := aocshared.S.ApplyDelta(beamInQueue.X, beamInQueue.Y)
			cell, err := aocshared.GridGetAt(*grid, southX, southY)

			if err != nil {
				continue
			}

			switch cell {
			case SPACE:
				currentBeamPositionsToBeProcessed = append(currentBeamPositionsToBeProcessed, Vector2{X: southX, Y: southY})
				aocshared.GridSetAt(grid, southX, southY, BEAM)
			case SPLITTER:
				splitTimes += 1
				eastX, eastY := aocshared.E.ApplyDelta(southX, southY)
				westX, westY := aocshared.W.ApplyDelta(southX, southY)

				if cellType, err := aocshared.GridGetAt(*grid, eastX, eastY); cellType == SPACE && err == nil {
					currentBeamPositionsToBeProcessed = append(currentBeamPositionsToBeProcessed, Vector2{X: eastX, Y: eastY})
					aocshared.GridSetAt(grid, eastX, eastY, BEAM)
				}

				if cellType, err := aocshared.GridGetAt(*grid, westX, westY); cellType == SPACE && err == nil {
					currentBeamPositionsToBeProcessed = append(currentBeamPositionsToBeProcessed, Vector2{X: westX, Y: westY})
					aocshared.GridSetAt(grid, westX, westY, BEAM)
				}
			case BEAM:

			default:
				panic(fmt.Errorf("unknown cellType: '%v'", cell)) // impossible to happen
			}
		}

		currentBeamPositionsToBeProcessed = slices.Delete(currentBeamPositionsToBeProcessed, 0, beamsInQueue)
	}

	return splitTimes
}

func getActiveTimelines(grid *aocshared.Grid[CellType], initialPosition Vector2) int {
	cache := map[Vector2]int{}
	return solvePosition(grid, cache, initialPosition)
}

func solvePosition(grid *aocshared.Grid[CellType], cache map[Vector2]int, position Vector2) int {
	if cached, ok := cache[position]; ok {
		return cached
	}

	result := 0

	southX, southY := aocshared.S.ApplyDelta(position.X, position.Y)
	if cell, err := aocshared.GridGetAt(*grid, southX, southY); err != nil {
		result += 1
	} else {
		switch cell {
		case SPACE:
			result += solvePosition(grid, cache, Vector2{X: southX, Y: southY})
		case SPLITTER:
			eastX, eastY := aocshared.E.ApplyDelta(southX, southY)
			westX, westY := aocshared.W.ApplyDelta(southX, southY)
			result += solvePosition(grid, cache, Vector2{X: eastX, Y: eastY})
			result += solvePosition(grid, cache, Vector2{X: westX, Y: westY})
		}
	}

	cache[position] = result

	return result
}

//go:inline
func getCellType(char rune) CellType {
	switch char {
	case '.':
		return SPACE
	case '^':
		return SPLITTER
	case 'S', '|':
		return BEAM
	default:
		panic(fmt.Errorf("unknown character: '%s'", string(char)))
	}
}
