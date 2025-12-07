package main

import (
	aocshared "aoc_shared"
	"fmt"
	"strings"
)

func main() {
	aocshared.DebugAndLogTask("2025 day 4", solve)
}

func solve() {
	input := aocshared.GetInput(2025, 3)
	rows := strings.Split(input, "\n")
	convertedRows := make([][]bool, 0)

	for _, row := range rows {
		trimmed := strings.TrimSpace(row)
		if trimmed == "" {
			continue
		}

		cells := make([]bool, len(row))
		for x, char := range row {
			if char == '.' {
				cells[x] = false
			} else {
				cells[x] = true
			}
		}

		convertedRows = append(convertedRows, cells)
	}

	rollsOfPaperCanLift := 0
	rollsOfPaperCanLiftOverTime := 0

	grid := aocshared.GridCreate(convertedRows)
	rollsOfPaperCanLift = solveIteration(grid)
	rollsOfPaperCanLiftOverTime = rollsOfPaperCanLift
	aocshared.GridQueuedOpsApply(grid)

	for {
		if canMove := solveIteration(grid); canMove == 0 {
			break
		} else {
			rollsOfPaperCanLiftOverTime += canMove
			aocshared.GridQueuedOpsApply(grid)
		}
	}

	fmt.Printf("Can Lift Rolls (P1): %d\n", rollsOfPaperCanLift)
	fmt.Printf("Can Lift Rolls (P2): %d\n", rollsOfPaperCanLiftOverTime)
}

func solveIteration(grid *aocshared.Grid[bool]) int {
	rollsOfPaperCanLift := 0

	aocshared.GridForEach(*grid, func(cell bool, x, y int) {
		if !cell {
			return
		}

		adjacencies := aocshared.GridGetAdjacencies(*grid, x, y)
		if canLift(adjacencies) {
			rollsOfPaperCanLift++
			aocshared.GridQueueOp(grid, func() {
				aocshared.GridSetAt(grid, x, y, false)
			})
			// aocshared.GridSetAt(grid, x, y, false)
		}
	})

	return rollsOfPaperCanLift
}

func canLift(s [8]bool) bool {
	counter := 0

	for _, roll := range s {
		if roll {
			counter++
		}

		if counter == 4 {
			return false
		}
	}

	return true
}
