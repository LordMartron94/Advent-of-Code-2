package main

import (
	aocshared "aoc_shared"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

const test = false

func main() {
	day9 := Day9{}

	aocshared.DebugAndLogTasks(
		"2025 day 9",
		aocshared.Task{Name: "Gather Input", Run: day9.GatherInput},
		aocshared.Task{Name: "Parse Input", Run: day9.ParseInput},
		aocshared.Task{Name: "Prepare Pairs", Run: day9.PreparePairs},
		aocshared.Task{Name: "Build Pairs", Run: day9.BuildPairs},
		aocshared.Task{Name: "Sort Pairs", Run: day9.SortPairs},
		aocshared.Task{Name: "Print Results", Run: day9.PrintResults},
	)
}

type CompressedSpace struct {
	xRanks map[int]int
	yRanks map[int]int

	satGrid [][]int
}

func buildCompressedSpace(coords []Vector2, boundaryLookup map[Vector2]bool) CompressedSpace {
	uniqueX := make(map[int]struct{})
	uniqueY := make(map[int]struct{})
	for _, coord := range coords {
		uniqueX[coord.x] = struct{}{}
		uniqueY[coord.y] = struct{}{}
	}

	sortedX := make([]int, 0, len(uniqueX))
	for x := range uniqueX {
		sortedX = append(sortedX, x)
	}
	slices.Sort(sortedX)

	sortedY := make([]int, 0, len(uniqueY))
	for y := range uniqueY {
		sortedY = append(sortedY, y)
	}
	slices.Sort(sortedY)

	N := len(sortedX) - 1
	M := len(sortedY) - 1

	xRanks := make(map[int]int, len(sortedX))
	for i, x := range sortedX {
		xRanks[x] = i
	}
	yRanks := make(map[int]int, len(sortedY))
	for i, y := range sortedY {
		yRanks[y] = i
	}

	rawGrid := make([][]int, M)

	tempShape := Shape{coords: coords, boundaryLookup: boundaryLookup}

	for j := 0; j < M; j++ {
		rawGrid[j] = make([]int, N)
		for i := 0; i < N; i++ {
			testP := Vector2{x: sortedX[i] + 1, y: sortedY[j] + 1}

			if !tempShape.IsValidCoord(testP) {
				rawGrid[j][i] = 1
			} else {
				rawGrid[j][i] = 0
			}
		}
	}

	sat := make([][]int, M+1)
	for r := 0; r <= M; r++ {
		sat[r] = make([]int, N+1)
	}

	for j := 1; j <= M; j++ {
		for i := 1; i <= N; i++ {
			sat[j][i] = rawGrid[j-1][i-1] + sat[j-1][i] + sat[j][i-1] - sat[j-1][i-1]
		}
	}

	return CompressedSpace{
		xRanks:  xRanks,
		yRanks:  yRanks,
		satGrid: sat,
	}
}

type Day9 struct {
	input    string
	coords   []Vector2
	shape    Shape
	pairsPt1 []Pair
	pairsPt2 []Pair
}

type Pair struct {
	idxA, idxB int
	distance   float64
	area       int
}

type Shape struct {
	coords          []Vector2
	boundaryLookup  map[Vector2]bool
	compressedSpace CompressedSpace
}

func (s Shape) isOnBoundary(coord Vector2) bool {
	_, ok := s.boundaryLookup[coord]
	return ok
}

func (s Shape) countVerticalCrossings(coord Vector2) int {
	numCrossings := 0
	N := len(s.coords)

	for i := 0; i < N; i++ {
		V_A := s.coords[i]
		V_B := s.coords[(i+1)%N]

		if V_A.x == V_B.x {
			wallX := V_A.x

			minY := V_A.y
			maxY := V_B.y
			if minY > maxY {
				minY, maxY = maxY, minY
			}

			if wallX > coord.x {
				if coord.y >= minY && coord.y < maxY {
					numCrossings++
				}
			}
		}
	}

	return numCrossings
}

func (s Shape) IsValidCoord(coord Vector2) bool {
	if s.isOnBoundary(coord) {
		return true
	}

	numCrossings := s.countVerticalCrossings(coord)

	isInside := numCrossings%2 != 0

	return isInside
}

func buildBoundaryLookup(coords []Vector2) map[Vector2]bool {
	boundaryMap := make(map[Vector2]bool)
	N := len(coords)

	if N < 2 {
		return boundaryMap
	}

	for i := 0; i < N; i++ {
		V_A := coords[i]
		V_B := coords[(i+1)%N]

		boundaryMap[V_A] = true

		dx := V_B.x - V_A.x
		dy := V_B.y - V_A.y

		if dy == 0 {
			stepX := 1
			if dx < 0 {
				stepX = -1
			}

			for x := V_A.x; x != V_B.x; x += stepX {
				boundaryMap[Vector2{x: x, y: V_A.y}] = true
			}
		}

		if dx == 0 {
			stepY := 1
			if dy < 0 {
				stepY = -1
			}

			for y := V_A.y; y != V_B.y; y += stepY {
				boundaryMap[Vector2{x: V_A.x, y: y}] = true
			}
		}

		boundaryMap[V_B] = true
	}

	return boundaryMap
}

func ShapeBuild(coords []Vector2) Shape {
	boundaryLookup := buildBoundaryLookup(coords)

	shape := Shape{
		coords:         coords,
		boundaryLookup: boundaryLookup,
	}

	shape.compressedSpace = buildCompressedSpace(coords, boundaryLookup)

	return shape
}

type Vector2 struct {
	x, y int
}

func (v Vector2) Subtract(other Vector2) Vector2 {
	return Vector2{
		x: v.x - other.x,
		y: v.y - other.y,
	}
}

func (v Vector2) AbsoluteDifference(other Vector2) Vector2 {
	difference := v.Subtract(other)

	differenceX := difference.x
	differenceY := difference.y

	if differenceX < 0 {
		differenceX = -differenceX
	}

	if differenceY < 0 {
		differenceY = -differenceY
	}

	return Vector2{
		x: differenceX,
		y: differenceY,
	}
}

func (v Vector2) Area(other Vector2) int {
	absoluteDifference := v.AbsoluteDifference(other)
	width := absoluteDifference.x
	height := absoluteDifference.y

	width += 1
	height += 1

	return width * height
}

func (v Vector2) EuclidianDistance(other Vector2) float64 {
	return math.Sqrt(v.EuclidianDistanceSquared(other))
}

func (v Vector2) EuclidianDistanceSquared(other Vector2) float64 {
	dx := float64(v.x - other.x)
	dy := float64(v.y - other.y)

	return dx*dx + dy*dy
}

func (d *Day9) GatherInput() {
	if test {
		d.input = aocshared.GetTestInput(2025, 9)
	} else {
		d.input = aocshared.GetInput(2025, 9)
	}
}

func (d *Day9) ParseInput() {
	lines := strings.Split(d.input, "\n")

	allCoords := make([]Vector2, len(lines))
	for i, line := range lines {
		coords := strings.Split(line, ",")

		x, _ := strconv.Atoi(coords[0])
		y, _ := strconv.Atoi(coords[1])

		allCoords[i] = Vector2{x: x, y: y}
	}

	d.coords = allCoords
}

func (d *Day9) PreparePairs() {
	d.shape = ShapeBuild(d.coords)
}

func isRectangleValid(shape Shape, cornerA, cornerB Vector2) bool {
	bottomRight := Vector2{x: cornerB.x, y: cornerA.y}
	topLeft := Vector2{x: cornerA.x, y: cornerB.y}

	if !shape.IsValidCoord(bottomRight) || !shape.IsValidCoord(topLeft) {
		return false
	}

	minXRank := min(shape.compressedSpace.xRanks[cornerA.x], shape.compressedSpace.xRanks[cornerB.x])
	maxXRank := max(shape.compressedSpace.xRanks[cornerA.x], shape.compressedSpace.xRanks[cornerB.x])
	minYRank := min(shape.compressedSpace.yRanks[cornerA.y], shape.compressedSpace.yRanks[cornerB.y])
	maxYRank := max(shape.compressedSpace.yRanks[cornerA.y], shape.compressedSpace.yRanks[cornerB.y])

	r2, c2 := maxYRank, maxXRank
	r1, c1 := minYRank, minXRank

	totalSum := shape.compressedSpace.satGrid[r2][c2] -
		shape.compressedSpace.satGrid[r1][c2] -
		shape.compressedSpace.satGrid[r2][c1] +
		shape.compressedSpace.satGrid[r1][c1]

	return totalSum == 0
}

func (d *Day9) BuildPairs() {
	pt1Pairs := make([]Pair, 0, len(d.coords)*len(d.coords)/2)
	pt2Pairs := make([]Pair, 0)

	for i := range d.coords {
		for j := i + 1; j < len(d.coords); j++ {
			coordA := d.coords[i]
			coordB := d.coords[j]

			pt1Pairs = append(pt1Pairs, Pair{
				idxA:     i,
				idxB:     j,
				distance: coordA.EuclidianDistance(coordB),
				area:     coordA.Area(coordB),
			})

			if isRectangleValid(d.shape, coordA, coordB) {
				pt2Pairs = append(pt2Pairs, Pair{
					idxA:     i,
					idxB:     j,
					distance: coordA.EuclidianDistance(coordB),
					area:     coordA.Area(coordB),
				})
			}
		}
	}

	d.pairsPt1 = pt1Pairs
	d.pairsPt2 = pt2Pairs
}

func (d *Day9) SortPairs() {
	slices.SortFunc(d.pairsPt1, func(a, b Pair) int {
		if a.area == b.area {
			return 0
		}

		if a.area > b.area {
			return -1
		}

		return 1
	})

	slices.SortFunc(d.pairsPt2, func(a, b Pair) int {
		if a.area == b.area {
			return 0
		}

		if a.area > b.area {
			return -1
		}

		return 1
	})
}

func (d *Day9) PrintResults() {
	// for i, pair := range d.pairsPt2 {
	// 	fmt.Printf("%03d) A=%v,B=%d;Distance=%.05f;Area=%d\n", i, d.coords[pair.idxA], d.coords[pair.idxB], pair.distance, pair.area)
	// }

	fmt.Printf("Solution Pt1: %d\n", d.pairsPt1[0].area)
	fmt.Printf("Solution Pt2: %d\n", d.pairsPt2[0].area)
}
