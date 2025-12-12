package aocshared

import (
	"fmt"
	"sort"
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

type ShapeOffset struct {
	DX, DY int
}

func GridPlace[TCell any](grid *Grid[TCell], shape []ShapeOffset, value TCell, isSuitable func(cell TCell, x, y int) bool) bool {
	if len(shape) == 0 {
		return false
	}

	for anchorY := 0; anchorY < grid.sizeY; anchorY++ {
		for anchorX := 0; anchorX < grid.sizeX; anchorX++ {
			allValid := true
			for _, offset := range shape {
				posX := anchorX + offset.DX
				posY := anchorY + offset.DY

				if posX < 0 || posX >= grid.sizeX || posY < 0 || posY >= grid.sizeY {
					allValid = false
					break
				}

				cell, err := GridGetAt(*grid, posX, posY)
				if err != nil || !isSuitable(cell, posX, posY) {
					allValid = false
					break
				}
			}

			if allValid {
				for _, offset := range shape {
					posX := anchorX + offset.DX
					posY := anchorY + offset.DY
					grid.rows[posY][posX] = value
				}
				return true
			}
		}
	}

	return false
}

type PlacementFlags struct {
	AllowRotate bool
	AllowFlip   bool
}

func rotateShape90(shape []ShapeOffset) []ShapeOffset {
	rotated := make([]ShapeOffset, len(shape))
	for i, offset := range shape {
		rotated[i] = ShapeOffset{DX: -offset.DY, DY: offset.DX}
	}
	return rotated
}

func flipShapeHorizontal(shape []ShapeOffset) []ShapeOffset {
	flipped := make([]ShapeOffset, len(shape))
	for i, offset := range shape {
		flipped[i] = ShapeOffset{DX: -offset.DX, DY: offset.DY}
	}
	return flipped
}

func flipShapeVertical(shape []ShapeOffset) []ShapeOffset {
	flipped := make([]ShapeOffset, len(shape))
	for i, offset := range shape {
		flipped[i] = ShapeOffset{DX: offset.DX, DY: -offset.DY}
	}
	return flipped
}

func normalizeShape(shape []ShapeOffset) []ShapeOffset {
	if len(shape) == 0 {
		return shape
	}

	minX, minY := shape[0].DX, shape[0].DY
	for _, offset := range shape {
		if offset.DX < minX {
			minX = offset.DX
		}
		if offset.DY < minY {
			minY = offset.DY
		}
	}

	normalized := make([]ShapeOffset, len(shape))
	for i, offset := range shape {
		normalized[i] = ShapeOffset{DX: offset.DX - minX, DY: offset.DY - minY}
	}
	return normalized
}

func shapeToKey(shape []ShapeOffset) string {
	coords := make([]string, len(shape))
	for i, offset := range shape {
		coords[i] = fmt.Sprintf("%d,%d", offset.DX, offset.DY)
	}
	sort.Strings(coords)
	return strings.Join(coords, ";")
}

func generateShapeTransformations(shape []ShapeOffset, flags PlacementFlags) [][]ShapeOffset {
	transformations := make([][]ShapeOffset, 0)
	seen := make(map[string]bool)

	addIfNew := func(transformed []ShapeOffset) {
		normalized := normalizeShape(transformed)
		key := shapeToKey(normalized)
		if !seen[key] {
			seen[key] = true
			transformations = append(transformations, normalized)
		}
	}

	base := normalizeShape(shape)
	addIfNew(base)

	if flags.AllowRotate {
		rot90 := rotateShape90(base)
		addIfNew(rot90)

		rot180 := rotateShape90(rot90)
		addIfNew(rot180)

		rot270 := rotateShape90(rot180)
		addIfNew(rot270)
	}

	if flags.AllowFlip {
		flipH := normalizeShape(flipShapeHorizontal(base))
		addIfNew(flipH)

		flipV := normalizeShape(flipShapeVertical(base))
		addIfNew(flipV)

		if flags.AllowRotate {
			flipHRot90 := normalizeShape(rotateShape90(flipH))
			addIfNew(flipHRot90)

			flipHRot180 := normalizeShape(rotateShape90(flipHRot90))
			addIfNew(flipHRot180)

			flipHRot270 := normalizeShape(rotateShape90(flipHRot180))
			addIfNew(flipHRot270)

			flipVRot90 := normalizeShape(rotateShape90(flipV))
			addIfNew(flipVRot90)

			flipVRot180 := normalizeShape(rotateShape90(flipVRot90))
			addIfNew(flipVRot180)

			flipVRot270 := normalizeShape(rotateShape90(flipVRot180))
			addIfNew(flipVRot270)
		}
	}

	return transformations
}

func canPlaceShapeAt[TCell any](grid *Grid[TCell], shape []ShapeOffset, anchorX, anchorY int, isSuitable func(cell TCell, x, y int) bool) bool {
	if len(shape) == 0 {
		return false
	}

	// Early bounds check on first offset
	firstOffset := shape[0]
	firstX := anchorX + firstOffset.DX
	firstY := anchorY + firstOffset.DY
	if firstX < 0 || firstX >= grid.sizeX || firstY < 0 || firstY >= grid.sizeY {
		return false
	}

	// Direct array access for better performance
	if !isSuitable(grid.rows[firstY][firstX], firstX, firstY) {
		return false
	}

	// Check remaining offsets
	for i := 1; i < len(shape); i++ {
		offset := shape[i]
		posX := anchorX + offset.DX
		posY := anchorY + offset.DY

		if posX < 0 || posX >= grid.sizeX || posY < 0 || posY >= grid.sizeY {
			return false
		}

		if !isSuitable(grid.rows[posY][posX], posX, posY) {
			return false
		}
	}
	return true
}

func placeShapeAt[TCell any](grid *Grid[TCell], shape []ShapeOffset, anchorX, anchorY int, value TCell) {
	for _, offset := range shape {
		posX := anchorX + offset.DX
		posY := anchorY + offset.DY
		grid.rows[posY][posX] = value
	}
}

type placementState[TCell any] struct {
	positions [][2]int
	values    []TCell
}

func savePlacementState[TCell any](grid *Grid[TCell], shape []ShapeOffset, anchorX, anchorY int) placementState[TCell] {
	state := placementState[TCell]{
		positions: make([][2]int, len(shape)),
		values:    make([]TCell, len(shape)),
	}
	for i, offset := range shape {
		posX := anchorX + offset.DX
		posY := anchorY + offset.DY
		state.positions[i] = [2]int{posX, posY}
		state.values[i] = grid.rows[posY][posX]
	}
	return state
}

func restorePlacementState[TCell any](grid *Grid[TCell], state placementState[TCell]) {
	for i, pos := range state.positions {
		grid.rows[pos[1]][pos[0]] = state.values[i]
	}
}

func gridCanFitShapesDFS[TCell any](
	grid *Grid[TCell],
	shapeCounts map[int]int,
	shapeTransformations map[int][][]ShapeOffset,
	remainingShapes []int,
	isSuitable func(cell TCell, x, y int) bool,
	markerValue TCell,
) bool {
	if len(remainingShapes) == 0 {
		return true
	}

	shapeID := remainingShapes[0]
	transformations := shapeTransformations[shapeID]

	for _, transformedShape := range transformations {
		if len(transformedShape) == 0 {
			continue
		}

		// Compute bounding box of shape to optimize anchor position iteration
		minDX, maxDX := transformedShape[0].DX, transformedShape[0].DX
		minDY, maxDY := transformedShape[0].DY, transformedShape[0].DY
		for _, offset := range transformedShape {
			if offset.DX < minDX {
				minDX = offset.DX
			}
			if offset.DX > maxDX {
				maxDX = offset.DX
			}
			if offset.DY < minDY {
				minDY = offset.DY
			}
			if offset.DY > maxDY {
				maxDY = offset.DY
			}
		}

		shapeWidth := maxDX - minDX + 1
		shapeHeight := maxDY - minDY + 1

		// Skip if shape is too large for grid
		if shapeWidth > grid.sizeX || shapeHeight > grid.sizeY {
			continue
		}

		// Only check anchor positions where the shape could fit
		maxAnchorY := grid.sizeY - shapeHeight
		maxAnchorX := grid.sizeX - shapeWidth

		for anchorY := 0; anchorY <= maxAnchorY && anchorY < grid.sizeY; anchorY++ {
			for anchorX := 0; anchorX <= maxAnchorX && anchorX < grid.sizeX; anchorX++ {
				// Quick pre-check: if first cell is occupied, skip
				firstOffset := transformedShape[0]
				firstX := anchorX + firstOffset.DX
				firstY := anchorY + firstOffset.DY
				if !isSuitable(grid.rows[firstY][firstX], firstX, firstY) {
					continue
				}

				if !canPlaceShapeAt(grid, transformedShape, anchorX, anchorY, isSuitable) {
					continue
				}

				state := savePlacementState(grid, transformedShape, anchorX, anchorY)

				placeShapeAt(grid, transformedShape, anchorX, anchorY, markerValue)

				shapeCounts[shapeID]--
				nextRemaining := remainingShapes
				if shapeCounts[shapeID] == 0 {
					nextRemaining = remainingShapes[1:]
				}

				if gridCanFitShapesDFS(grid, shapeCounts, shapeTransformations, nextRemaining, isSuitable, markerValue) {
					return true
				}

				shapeCounts[shapeID]++
				restorePlacementState(grid, state)
			}
		}
	}

	return false
}

func GridCanFitShapes[TCell any](
	grid *Grid[TCell],
	shapes map[int][]ShapeOffset,
	shapeCounts map[int]int,
	flags PlacementFlags,
	isSuitable func(cell TCell, x, y int) bool,
	markerValue TCell,
) bool {
	if len(shapes) == 0 {
		return true
	}

	shapeTransformations := make(map[int][][]ShapeOffset)
	remainingShapes := make([]int, 0)
	shapeConstraints := make(map[int]int) // Number of transformations (fewer = more constrained)

	for shapeID, shape := range shapes {
		if shapeCounts[shapeID] <= 0 {
			continue
		}
		transformations := generateShapeTransformations(shape, flags)
		shapeTransformations[shapeID] = transformations
		shapeConstraints[shapeID] = len(transformations)
		remainingShapes = append(remainingShapes, shapeID)
	}

	if len(remainingShapes) == 0 {
		return true
	}

	// Sort shapes by constraint level (fewer transformations = more constrained = try first)
	// This helps prune the search tree faster
	for i := 0; i < len(remainingShapes)-1; i++ {
		for j := i + 1; j < len(remainingShapes); j++ {
			if shapeConstraints[remainingShapes[i]] > shapeConstraints[remainingShapes[j]] {
				remainingShapes[i], remainingShapes[j] = remainingShapes[j], remainingShapes[i]
			}
		}
	}

	shapeCountsCopy := make(map[int]int)
	for k, v := range shapeCounts {
		shapeCountsCopy[k] = v
	}

	return gridCanFitShapesDFS(grid, shapeCountsCopy, shapeTransformations, remainingShapes, isSuitable, markerValue)
}
