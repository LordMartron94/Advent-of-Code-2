package main

import (
	aocshared "aoc_shared"
	"fmt"
	"strconv"
	"strings"
)

func floorDiv(n, d int) int {
	if n >= 0 {
		return n / d
	}
	return (n - d + 1) / d
}

func wrapAround(maxNum, current, diff int) (int, int) {
	rangeSize := maxNum + 1
	raw := current + diff

	remainder := raw % rangeSize
	if remainder < 0 {
		remainder += rangeSize
	}

	var zeros int
	if diff > 0 {
		zeros = floorDiv(raw, rangeSize) - floorDiv(current, rangeSize)
	} else if diff < 0 {
		zeros = floorDiv(current-1, rangeSize) - floorDiv(raw-1, rangeSize)
	} else {
		zeros = 0
	}

	return remainder, zeros
}

func main() {
	input := aocshared.GetInput(2025, 1)
	rotations := strings.Split(input, "\n")
	fmt.Printf("Number of rotations: %d\n", len(rotations))

	current := 50
	zeroCounter1 := 0
	zeroCounter2 := 0
	for _, rotation := range rotations {
		if len(rotation) == 0 {
			continue
		}

		direction := string(rotation[0])
		amount, _ := strconv.Atoi(string(rotation[1:]))

		if direction == "L" {
			amount = -amount
		}

		nextVal, zeroCrossings := wrapAround(99, current, amount)

		if nextVal == 0 {
			zeroCounter1++
		}

		zeroCounter2 += zeroCrossings

		current = nextVal
	}

	fmt.Printf("Answer Part 1: %d\n", zeroCounter1)
	fmt.Printf("Answer Part 2: %d\n", zeroCounter2)
}
