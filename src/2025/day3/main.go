package main

import (
	aocshared "aoc_shared"
	"fmt"
	"strings"
)

func main() {
	aocshared.DebugAndLogTask("2025, day 3", run)
}

func run() {
	input := aocshared.GetTestInput(2025, 3)
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
	arraySize := len(bank)
	currentNeedleIdx := 0

	bankV := 0

	for i := 1; i <= numDigits; i++ {
		digitsLeftToSeekAfter := numDigits - i

		v, vI := getMaxWithIdx(bank, currentNeedleIdx, arraySize-digitsLeftToSeekAfter)

		bankV = (bankV * 10) + v
		currentNeedleIdx = vI + 1
	}

	*joltSum += bankV
}

func getMaxWithIdx(s string, startIdx, endIdx int) (v int, idx int) {
	maxV := -1
	maxI := -1

	for i := startIdx; i < endIdx; i++ {
		v := int(rune(s[i]) - '0')

		if v == 9 {
			return v, i
		}

		if v > maxV {
			maxV = v
			maxI = i
		}
	}

	return maxV, maxI
}
