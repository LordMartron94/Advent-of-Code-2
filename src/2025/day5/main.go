package main

import (
	aocshared "aoc_shared"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func main() {
	aocshared.DebugAndLogTask("2025 day 5", solve)
}

func solve() {
	input := aocshared.GetInput(2025, 5)
	lines := strings.Split(input, "\n")

	ranges := make([]ingredientRange, 0)
	ingredients := make([]int, 0)
	for _, line := range lines {
		if strings.Contains(line, "-") {
			buildRange(line, &ranges)
		} else {
			if ingredient, err := strconv.Atoi(line); err == nil {
				ingredients = append(ingredients, ingredient)
			}
		}
	}

	slices.SortStableFunc(ranges, func(a, b ingredientRange) int {
		if a.start < b.start {
			return -1
		}
		if a.start == b.start {
			return 0
		}

		return 1
	})

	minimizedRanges := minimizeRanges(ranges)

	numInRange := ingredientsInRange(minimizedRanges, ingredients)

	fmt.Printf("Fresh Ingredients (P1): %d\n", numInRange)
	fmt.Printf("Fresh Ingredient Ranges (P2): %d\n", numValidIngredients(minimizedRanges))
}

type ingredientRange struct {
	start int
	end   int
}

func (i ingredientRange) IsInRange(num int) bool {
	return num >= i.start && num <= i.end
}

func (i ingredientRange) ValidIngredients() int {
	return (i.end - i.start) + 1
}

func numValidIngredients(minimizedRangeSlice []ingredientRange) int {
	valid := 0

	for _, ingredientRange := range minimizedRangeSlice {
		valid += ingredientRange.ValidIngredients()
	}

	return valid
}

func ingredientsInRange(minimizedRangeSlice []ingredientRange, ingredients []int) int {
	fresh := 0
	for _, ingredient := range ingredients {
		nearestIdx := FindNearestIndex(minimizedRangeSlice, ingredient)
		ingredientRange := minimizedRangeSlice[nearestIdx]
		if ingredientRange.IsInRange(ingredient) {
			fresh++
		}
	}

	return fresh
}

func minimizeRanges(sortedRangeSlice []ingredientRange) []ingredientRange {
	processed := 0
	minimized := make([]ingredientRange, 0)

	originalSize := len(sortedRangeSlice)
	for processed < originalSize {
		currentRange := sortedRangeSlice[processed]
		for s := processed + 1; s < originalSize; s++ {
			next := sortedRangeSlice[s]

			if s >= originalSize || !currentRange.IsInRange(next.start) {
				break
			}

			if next.end > currentRange.end {
				currentRange.end = next.end
			}

			processed++
		}

		minimized = append(minimized, currentRange)
		processed++
	}

	return minimized
}

func buildRange(line string, rangeSlice *[]ingredientRange) {
	parts := strings.Split(line, "-")
	part1Cnv, _ := strconv.Atoi(parts[0])
	part2Cnv, _ := strconv.Atoi(parts[1])

	*rangeSlice = append(*rangeSlice, ingredientRange{
		start: part1Cnv,
		end:   part2Cnv,
	})
}

// FindNearestIndex finds the index of the element closest to the target.
// Complexity: O(log n)
func FindNearestIndex(sorted []ingredientRange, target int) int {
	if len(sorted) == 0 {
		return -1
	}

	idx, found := slices.BinarySearchFunc(sorted, target, func(element ingredientRange, target int) int {
		if element.start < target {
			return -1
		}

		if element.start == target {
			return 0
		}

		return 1
	})

	if found {
		return idx
	}

	if idx == 0 {
		return 0
	}

	if idx == len(sorted) {
		return len(sorted) - 1
	}

	return idx - 1
}
