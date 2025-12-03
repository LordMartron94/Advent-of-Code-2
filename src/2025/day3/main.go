package main

import (
	aocshared "aoc_shared"
	"fmt"
	"math"
	"strings"
)

func main() {
	input := aocshared.GetInput(2025, 3)
	banks := strings.Split(input, "\n")

	joltSum := 0
	joltSum2 := 0

	for _, bank := range banks {
		trimmed := strings.TrimSpace(bank)
		if trimmed == "" {
			continue
		}

		processBank(trimmed, 2, &joltSum)
		processBank(trimmed, 12, &joltSum2)
	}

	fmt.Printf("Sum Jolt P1: %d\n", joltSum)
	fmt.Printf("Sum Jolt P2: %d\n", joltSum2)
}

func processBank(bank string, numDigits int, joltSum *int) {
	arr := make([]int, len(bank))

	for i, char := range bank {
		v := int(char - '0')
		arr[i] = v
	}

	arraySize := len(arr)
	currentNeedleIdx := 0

	bankV := 0

	for i := 1; i <= numDigits; i++ {
		digitsLeftToSeekAfter := numDigits - i

		shiftFactor := int(math.Pow(10, float64(digitsLeftToSeekAfter)))
		v, vI := getMaxWithIdx(arr, currentNeedleIdx, arraySize-digitsLeftToSeekAfter)

		bankV += v * shiftFactor
		currentNeedleIdx = vI + 1
	}

	*joltSum += bankV
}

func getMaxWithIdx(s []int, startIdx, endIdx int) (v int, idx int) {
	maxV := -1
	maxI := -1

	for i := startIdx; i < endIdx; i++ {
		v := s[i]

		if v > maxV {
			maxV = v
			maxI = i
		}
	}

	return maxV, maxI
}
