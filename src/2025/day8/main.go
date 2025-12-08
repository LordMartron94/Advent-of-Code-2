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
	day8 := Day8{}

	aocshared.DebugAndLogTasks(
		"2025 day 8",
		aocshared.Task{Name: "Gather Input", Run: day8.GatherInput},
		aocshared.Task{Name: "Parse Input", Run: day8.ParseInput},
		aocshared.Task{Name: "Build Edges", Run: day8.BuildEdges},
		aocshared.Task{Name: "Sort Edges", Run: day8.SortEdges},
		aocshared.Task{Name: "Connect Boxes", Run: day8.ConnectBoxes},
		aocshared.Task{Name: "Print Results", Run: day8.PrintResults},
	)
}

type Day8 struct {
	pairsToConsider int
	input           string
	boxes           []Vector3
	edges           []Edge
	circuits        []*Circuit

	part1Result int
	part2Result int
}

type Edge struct {
	idxA, idxB int
	distance   float64
}

type Circuit struct {
	members []int
}

func (c *Circuit) AddBox(idx int) {
	c.members = append(c.members, idx)
}

func (c *Circuit) Members() []int {
	return c.members
}

func (c Circuit) Size() int {
	return len(c.members)
}

type Vector3 struct {
	x, y, z int
}

func (v Vector3) EuclidianDistance(other Vector3) float64 {
	return math.Sqrt(v.EuclidianDistanceSquared(other))
}

func (v Vector3) EuclidianDistanceSquared(other Vector3) float64 {
	dx := float64(v.x - other.x)
	dy := float64(v.y - other.y)
	dz := float64(v.z - other.z)

	return dx*dx + dy*dy + dz*dz
}

func (d *Day8) GatherInput() {
	if test {
		d.input = aocshared.GetTestInput(2025, 8)
		d.pairsToConsider = 10
	} else {
		d.input = aocshared.GetInput(2025, 8)
		d.pairsToConsider = 1000
	}
}

func (d *Day8) ParseInput() {
	lines := strings.Split(d.input, "\n")

	boxes := make([]Vector3, len(lines))
	for i, line := range lines {
		coords := strings.Split(line, ",")

		x, _ := strconv.Atoi(coords[0])
		y, _ := strconv.Atoi(coords[1])
		z, _ := strconv.Atoi(coords[2])

		boxes[i] = Vector3{x: x, y: y, z: z}
	}

	d.boxes = boxes
}

func (d *Day8) BuildEdges() {
	edges := make([]Edge, 0, len(d.boxes)*len(d.boxes)/2)

	for i := range d.boxes {
		for j := i + 1; j < len(d.boxes); j++ {
			edges = append(edges, Edge{
				idxA:     i,
				idxB:     j,
				distance: d.boxes[i].EuclidianDistance(d.boxes[j]),
			})
		}
	}

	d.edges = edges
}

func (d *Day8) SortEdges() {
	slices.SortFunc(d.edges, func(a, b Edge) int {
		if a.distance == b.distance {
			return 0
		}

		if a.distance < b.distance {
			return -1
		}

		return 1
	})
}

func (d *Day8) ConnectBoxes() {
	circuits := make([]*Circuit, len(d.boxes))
	circuitMap := make(map[int]*Circuit, len(d.boxes))

	for idx := range d.boxes {
		c := &Circuit{members: []int{idx}}
		circuits[idx] = c
		circuitMap[idx] = c
	}

	deleteCircuit := func(circuit *Circuit) {
		circuits = slices.DeleteFunc(circuits, func(element *Circuit) bool {
			return element == circuit
		})
	}

	union := func(destination, source *Circuit) {
		for _, member := range source.Members() {
			destination.AddBox(member)
			circuitMap[member] = destination
		}
		deleteCircuit(source)
	}

	pairsProcessed := 0
	part1Done := false

	for _, edge := range d.edges {
		a := circuitMap[edge.idxA]
		b := circuitMap[edge.idxB]

		if a != b {
			if a.Size() < b.Size() {
				union(b, a)
			} else {
				union(a, b)
			}
		}

		pairsProcessed++

		// ----- Part 1: snapshot after N shortest connections -----
		if !part1Done && pairsProcessed == d.pairsToConsider {
			sizes := make([]int, len(circuits))
			for i, c := range circuits {
				sizes[i] = c.Size()
			}
			slices.Sort(sizes)

			n := len(sizes)
			d.part1Result = sizes[n-1] * sizes[n-2] * sizes[n-3]

			part1Done = true
		}

		// ----- Part 2: first time everything is in one circuit -----
		if len(circuits) == 1 {
			finalA := d.boxes[edge.idxA]
			finalB := d.boxes[edge.idxB]

			d.part2Result = finalA.x * finalB.x
			break
		}
	}

	d.circuits = circuits
}

func (d *Day8) PrintResults() {
	fmt.Printf("Solution Pt1: %d\n", d.part1Result)
	fmt.Printf("Solution Pt2: %d\n", d.part2Result)
}
