package main

import (
	aocshared "aoc_shared"
	"fmt"
	"slices"
	"strings"
)

const test = false
const test2 = false

func main() {
	day11 := day11{}

	aocshared.DebugAndLogTasks(
		"2025 day 11",
		aocshared.Task{Name: "Get Input", Run: day11.GetInput},
		aocshared.Task{Name: "Parse Input", Run: day11.ParseInput},
		aocshared.Task{Name: "Solve Part 1", Run: day11.SolvePart1},
		aocshared.Task{Name: "Solve Part 2", Run: day11.SolvePart2},
		aocshared.Task{Name: "Print Results", Run: day11.PrintResults},
	)
}

type day11 struct {
	input   string
	devices []device

	numPaths  int
	numPaths2 int
}

func (d *day11) GetInput() {
	if test {
		if test2 {
			d.input = aocshared.GetTestInputN(2025, 11, 2)
		} else {
			d.input = aocshared.GetTestInput(2025, 11)
		}
	} else {
		d.input = aocshared.GetInput(2025, 11)
	}
}

type device struct {
	ID          string
	Connections []string
}

func (d *day11) ParseInput() {
	lines := strings.Split(d.input, "\n")
	devices := make([]device, len(lines))

	for i, line := range lines {
		parts := strings.Fields(line)

		devices[i] = device{
			ID:          parts[0][0 : len(parts[0])-1],
			Connections: parts[1:],
		}
	}

	d.devices = devices
}

func (d *day11) SolvePart1() {
	deviceMap := map[string]device{}

	for _, device := range d.devices {
		deviceMap[device.ID] = device
	}

	cache := map[string]int{}
	d.numPaths = solveDevicePt1(deviceMap["you"], cache, deviceMap)
}

func (d *day11) SolvePart2() {
	deviceMap := map[string]device{}

	for _, device := range d.devices {
		deviceMap[device.ID] = device
	}

	cache := map[part2CacheKey]int{}
	d.numPaths2 = solveDevicePt2(deviceMap["svr"], cache, deviceMap, false, false)
}

func solveDevicePt1(device device, cache map[string]int, deviceMap map[string]device) int {
	if v, cached := cache[device.ID]; cached {
		return v
	}

	result := 0

	if slices.Contains(device.Connections, "out") {
		result += 1
	} else {
		for _, otherDeviceID := range device.Connections {
			result += solveDevicePt1(deviceMap[otherDeviceID], cache, deviceMap)
		}
	}

	cache[device.ID] = result
	return result
}

type part2CacheKey struct {
	ID        string
	passedDac bool
	passedFFT bool
}

func solveDevicePt2(device device, cache map[part2CacheKey]int, deviceMap map[string]device, passedDac, passedFFT bool) int {
	key := part2CacheKey{
		ID:        device.ID,
		passedDac: passedDac,
		passedFFT: passedFFT,
	}
	if v, cached := cache[key]; cached {
		return v
	}

	result := 0

	if device.ID == "fft" {
		key.passedFFT = true
	} else if device.ID == "dac" {
		key.passedDac = true
	}

	if slices.Contains(device.Connections, "out") && key.passedDac && key.passedFFT {
		result += 1
	} else {
		for _, otherDeviceID := range device.Connections {
			result += solveDevicePt2(deviceMap[otherDeviceID], cache, deviceMap, key.passedDac, key.passedFFT)
		}
	}

	cache[key] = result
	return result
}

func (d *day11) PrintResults() {
	fmt.Printf("Number of Different Paths (pt1): %d\n", d.numPaths)
	fmt.Printf("Number of Different Paths (pt2): %d\n", d.numPaths2)
}
