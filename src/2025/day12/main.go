package main

import (
	aocshared "aoc_shared"
	"fmt"
	"strconv"
	"strings"
)

const test = false

func main() {
	day12 := day12{}

	aocshared.DebugAndLogTasks(
		"2025 day 12",
		aocshared.Task{Name: "Get Input", Run: day12.GetInput},
		aocshared.Task{Name: "Parse Input", Run: day12.ParseInput},
		aocshared.Task{Name: "Solve Pt1", Run: day12.SolvePart1},
		aocshared.Task{Name: "Print Results", Run: day12.PrintResults},
	)
}

type day12 struct {
	input   string
	shapes  []Shape
	regions []Region

	validRegions int
}

type Shape struct {
	Idx          int
	Parts        [9]bool // 0,0 1,0; 2,0; 0,1 1,1 2,1; 0,2 1,2 2,2
	NumPartsTrue int
}

type Region struct {
	Width         int
	Height        int
	PresentCounts []int
}

func (d *day12) GetInput() {
	if test {
		d.input = aocshared.GetTestInput(2025, 12)
	} else {
		d.input = aocshared.GetInput(2025, 12)
	}
}

// Parser written by Gemini, as parsing is not what this problem is about.

func (d *day12) ParseInput() {
	lines := strings.Split(d.input, "\n")
	d.shapes = make([]Shape, 0)
	d.regions = make([]Region, 0)

	shapeParsing := true

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		if shapeParsing && strings.HasSuffix(lines[i], ":") {
			if i+3 >= len(lines) {
				panic(fmt.Errorf("malformed shape definition starting at line %d", i+1))
			}

			s := d.parseShapeBlock(lines[i], lines[i+1], lines[i+2], lines[i+3])
			d.shapes = append(d.shapes, s)
			i += 3
			continue
		}

		shapeParsing = false

		region, err := d.parseRegionLine(line)
		if err != nil {
			panic(fmt.Errorf("unhandled line format (after shapes) at line %d: %s, error: %w", i+1, line, err))
		}
		d.regions = append(d.regions, region)
	}

	fmt.Printf("Parsed %d shapes and %d regions.\n", len(d.shapes), len(d.regions))
}

func (d *day12) parseShapeBlock(idxLine, r1, r2, r3 string) Shape {
	idxStr := strings.TrimSuffix(idxLine, ":")
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		panic(fmt.Errorf("invalid shape index format: %s", idxLine))
	}

	rows := []string{r1, r2, r3}
	s := Shape{Idx: idx}
	s.NumPartsTrue = 0

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			arrIdx := (y * 3) + x

			if x >= len(rows[y]) {
				s.Parts[arrIdx] = false
			} else {
				isPart := rows[y][x] == '#'
				s.Parts[arrIdx] = isPart
				if isPart {
					s.NumPartsTrue++
				}
			}
		}
	}
	return s
}

func (d *day12) parseRegionLine(line string) (Region, error) {
	parts := strings.SplitN(line, ": ", 2)
	if len(parts) != 2 {
		return Region{}, fmt.Errorf("region line missing dimensions or counts: %s", line)
	}

	dimParts := strings.Split(parts[0], "x")
	if len(dimParts) != 2 {
		return Region{}, fmt.Errorf("region dimensions malformed (expected WxH): %s", parts[0])
	}
	width, errW := strconv.Atoi(dimParts[0])
	height, errH := strconv.Atoi(dimParts[1])
	if errW != nil || errH != nil {
		return Region{}, fmt.Errorf("invalid region dimension integers: %s", parts[0])
	}

	countStrs := strings.Fields(parts[1])
	counts := make([]int, len(countStrs))
	for i, s := range countStrs {
		count, err := strconv.Atoi(s)
		if err != nil {
			return Region{}, fmt.Errorf("invalid present count integer: %s", s)
		}
		counts[i] = count
	}

	return Region{
		Width:         width,
		Height:        height,
		PresentCounts: counts,
	}, nil
}

func (d *day12) SolvePart1() {
	numValid := 0

	for _, region := range d.regions {
		if solveRegion(region, d.shapes) {
			numValid++
		}
	}

	d.validRegions = numValid
}

func solveRegion(region Region, shapes []Shape) bool {
	shapeMap := make(map[int][]aocshared.ShapeOffset)
	shapeCounts := make(map[int]int)
	totalAreaNeeded := 0

	for shapeIdx, required := range region.PresentCounts {
		if required == 0 {
			continue
		}

		shape := shapes[shapeIdx]
		offsets := convertShapeToOffsets(shape)
		shapeMap[shapeIdx] = offsets
		shapeCounts[shapeIdx] = required
		totalAreaNeeded += shape.NumPartsTrue * required
	}

	if len(shapeMap) == 0 {
		return true
	}

	// Early exit: if total area needed exceeds grid area, impossible
	gridArea := region.Width * region.Height
	if totalAreaNeeded > gridArea {
		return false
	}

	grid := createEmptyGrid(region.Width, region.Height)

	flags := aocshared.PlacementFlags{
		AllowRotate: true,
		AllowFlip:   true,
	}

	return aocshared.GridCanFitShapes(
		grid,
		shapeMap,
		shapeCounts,
		flags,
		func(cell bool, x, y int) bool {
			return !cell
		},
		true,
	)
}

func convertShapeToOffsets(shape Shape) []aocshared.ShapeOffset {
	offsets := make([]aocshared.ShapeOffset, 0)

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			arrIdx := (y * 3) + x
			if shape.Parts[arrIdx] {
				offsets = append(offsets, aocshared.ShapeOffset{DX: x, DY: y})
			}
		}
	}

	return offsets
}

func createEmptyGrid(width, height int) *aocshared.Grid[bool] {
	rows := make([][]bool, height)
	for y := 0; y < height; y++ {
		rows[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			rows[y][x] = false
		}
	}
	return aocshared.GridCreate(rows)
}

func (d *day12) PrintResults() {
	fmt.Printf("Number of Valid Regions: %d\n", d.validRegions)
	fmt.Printf("Pt 2: %d\n", 0)
}
