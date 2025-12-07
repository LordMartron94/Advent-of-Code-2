package aocshared

import (
	"fmt"
	"strings"
)

type Grid[TCell any] struct {
	rows [][]TCell

	queuedOps []func()

	sizeX int
	sizeY int
}

type Direction int

const (
	N Direction = iota
	NE
	E
	SE
	S
	SW
	W
	NW
	DirectionCount
)

var DirDelta = [DirectionCount]struct{ DX, DY int }{
	N:  {0, -1},
	NE: {1, -1},
	E:  {1, 0},
	SE: {1, 1},
	S:  {0, 1},
	SW: {-1, 1},
	W:  {-1, 0},
	NW: {-1, -1},
}

func (d Direction) ApplyDelta(posX, posY int) (x, y int) {
	return posX + DirDelta[d].DX, posY + DirDelta[d].DY
}

func (d Direction) Delta() (dx, dy int) {
	return DirDelta[d].DX, DirDelta[d].DY
}

func (d Direction) Opposite() Direction {
	return (d + 4) % 8
}

func (d Direction) RotateCW() Direction {
	return (d + 1) % 8
}

func (d Direction) RotateCCW() Direction {
	return (d + 7) % 8
}

var deltas = []struct{ dx, dy int }{
	{0, -1}, {1, -1}, {1, 0}, {1, 1},
	{0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
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

func GridClone[TCell any](grid *Grid[TCell]) *Grid[TCell] {
	clonedGrid := &Grid[TCell]{
		sizeX: grid.sizeX,
		sizeY: grid.sizeY,
	}

	clonedGrid.rows = make([][]TCell, grid.sizeY)

	for y := 0; y < grid.sizeY; y++ {
		clonedGrid.rows[y] = make([]TCell, grid.sizeX)
		copy(clonedGrid.rows[y], grid.rows[y])
	}

	clonedGrid.queuedOps = make([]func(), len(grid.queuedOps))
	copy(clonedGrid.queuedOps, grid.queuedOps)

	return clonedGrid
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
	fixedAdjacencies := [8]TCell{}
	for d := Direction(0); d < 8; d++ {
		delta := DirDelta[d]
		if cell, err := GridGetAt(grid, x+delta.DX, y+delta.DY); err == nil {
			fixedAdjacencies[d] = cell
		}
	}
	return fixedAdjacencies
}

func GridForEach[TCell any](grid Grid[TCell], exec func(cell TCell, x, y int)) {
	for y, row := range grid.rows {
		for x, cell := range row {
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
