package main

import (
	aocshared "aoc_shared"
	"fmt"
	"strings"
)

type Grid[TCell any] struct {
	rows [][]TCell

	queuedOps []func()

	sizeX int
	sizeY int
}

func GridCreate[TCell any](rows [][]TCell) *Grid[TCell] {
	if len(rows) < 1 {
		panic("grid must have at least one row")
	}

	sizeY := len(rows)
	sizeX := len(rows[0])

	return &Grid[TCell]{
		rows:      rows,
		sizeX:     sizeX,
		sizeY:     sizeY,
		queuedOps: make([]func(), 0),
	}
}

func GridQueueOp[TCell any](grid *Grid[TCell], op func()) {
	grid.queuedOps = append(grid.queuedOps, op)
}

func GridQueuedOpsApply[TCell any](grid *Grid[TCell]) {
	for _, op := range grid.queuedOps {
		op()
	}

	grid.queuedOps = []func(){}
}

func GridGetAt[TCell any](grid Grid[TCell], x, y int) (TCell, error) {
	if x >= grid.sizeX || y >= grid.sizeY || x < 0 || y < 0 {
		var zero TCell
		return zero, fmt.Errorf("invalid coords")
	}

	return grid.rows[y][x], nil
}

func GridSetAt[TCell any](grid *Grid[TCell], x, y int, cell TCell) error {
	if x >= grid.sizeX || y >= grid.sizeY || x < 0 || y < 0 {
		return fmt.Errorf("invalid coords")
	}

	grid.rows[y][x] = cell

	return nil
}

func GridGetAdjacencies[TCell any](grid Grid[TCell], x, y int) [8]TCell {
	adjacencies := [8]TCell{}

	if cell, err := GridGetAt(grid, x, y-1); err == nil { // top
		adjacencies[0] = cell
	}

	if cell, err := GridGetAt(grid, x+1, y-1); err == nil { // top-right-diagonal
		adjacencies[1] = cell
	}

	if cell, err := GridGetAt(grid, x+1, y); err == nil { // right
		adjacencies[2] = cell
	}

	if cell, err := GridGetAt(grid, x+1, y+1); err == nil { // bottom-right-diagonal
		adjacencies[3] = cell
	}

	if cell, err := GridGetAt(grid, x, y+1); err == nil { // bottom
		adjacencies[4] = cell
	}

	if cell, err := GridGetAt(grid, x-1, y+1); err == nil { // bottom-left-diagonal
		adjacencies[5] = cell
	}

	if cell, err := GridGetAt(grid, x-1, y); err == nil { // left
		adjacencies[6] = cell
	}

	if cell, err := GridGetAt(grid, x-1, y-1); err == nil { // top-left-diagonal
		adjacencies[7] = cell
	}

	return adjacencies
}

func GridForEach[TCell any](grid Grid[TCell], exec func(cell TCell, x, y int)) {
	for y := 0; y < grid.sizeY; y++ {
		for x := 0; x < grid.sizeX; x++ {
			cell, _ := GridGetAt(grid, x, y)
			exec(cell, x, y)
		}
	}
}

func GridDebug[TCell any](grid Grid[TCell], cellFormatter func(cell TCell) string) {
	if grid.sizeY == 0 {
		fmt.Println("(empty grid)")
		return
	}

	strRows := mapRowsToStrings(grid, cellFormatter)
	widths := getColWidths(strRows)

	printGridAligned(strRows, widths)
}

func mapRowsToStrings[TCell any](grid Grid[TCell], cellFormatter func(cell TCell) string) [][]string {
	strRows := make([][]string, grid.sizeY)
	for y := 0; y < grid.sizeY; y++ {
		strRows[y] = make([]string, grid.sizeX)
		for x := 0; x < grid.sizeX; x++ {
			strRows[y][x] = cellFormatter(grid.rows[y][x])
		}
	}
	return strRows
}

func getColWidths(rows [][]string) []int {
	if len(rows) == 0 {
		return nil
	}

	width := len(rows[0])
	colWidths := make([]int, width)

	for _, row := range rows {
		for x, val := range row {
			if len(val) > colWidths[x] {
				colWidths[x] = len(val)
			}
		}
	}
	return colWidths
}

func printGridAligned(rows [][]string, widths []int) {
	for _, row := range rows {
		printRow(row, widths)
	}
}

func printRow(row []string, widths []int) {
	var sb strings.Builder
	for i, val := range row {
		format := fmt.Sprintf("%%-%ds ", widths[i])
		sb.WriteString(fmt.Sprintf(format, val))
	}
	fmt.Println(sb.String())
}

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

	grid := GridCreate(convertedRows)
	rollsOfPaperCanLift = solveIteration(grid)
	rollsOfPaperCanLiftOverTime = rollsOfPaperCanLift
	GridQueuedOpsApply(grid)

	for {
		if canMove := solveIteration(grid); canMove == 0 {
			break
		} else {
			rollsOfPaperCanLiftOverTime += canMove
			GridQueuedOpsApply(grid)
		}
	}

	fmt.Printf("Can Lift Rolls (P1): %d\n", rollsOfPaperCanLift)
	fmt.Printf("Can Lift Rolls (P2): %d\n", rollsOfPaperCanLiftOverTime)
}

func solveIteration(grid *Grid[bool]) int {
	rollsOfPaperCanLift := 0

	GridForEach(*grid, func(cell bool, x, y int) {
		if !cell {
			return
		}

		adjacencies := GridGetAdjacencies(*grid, x, y)
		if canLift(adjacencies) {
			rollsOfPaperCanLift++
			GridQueueOp(grid, func() {
				GridSetAt(grid, x, y, false)
			})
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
