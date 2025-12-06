package main

import (
	aocshared "aoc_shared"
	"fmt"
	"strconv"
	"strings"
)

func main() {
	aocshared.DebugAndLogTask("2025 day 6", solve)
}

type Operator func(current, next int) int

func MultiplyOp(current, next int) int {
	return current * next
}

func AddOp(current, next int) int {
	return current + next
}

func solve() {
	input := aocshared.GetInput(2025, 6)
	original := strings.Split(input, "\n")
	lines := original[:len(original)-1]

	rawOperators := strings.Fields(original[len(original)-1])

	operators := []Operator{}

	for _, operator := range rawOperators {
		if operator == "*" {
			operators = append(operators, MultiplyOp)
		}
		if operator == "+" {
			operators = append(operators, AddOp)
		}
	}

	numProblems := len(operators)

	part1Problems := make([][][]rune, numProblems)
	part2Problems := make([][][]rune, numProblems) // problemNum -> numberPos -> digits

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Pt1
		numbersInLine := strings.Fields(trimmed)

		for problemNum, number := range numbersInLine {
			if part1Problems[problemNum] == nil {
				part1Problems[problemNum] = make([][]rune, 0)
			}
			part1Problems[problemNum] = append(part1Problems[problemNum], []rune(number))
		}

		// Pt2
		currentProblemNum := 0

		numberStartColumns := make([]int, 0)
		isWhitespace := true

		for i, char := range line {
			if char != ' ' && isWhitespace {
				numberStartColumns = append(numberStartColumns, i)
				isWhitespace = false
			} else if char == ' ' {
				isWhitespace = true
			}
		}

		for colIndex, char := range line {
			currentProblemNum = 0
			for i := 1; i < len(numberStartColumns); i++ {
				if colIndex >= numberStartColumns[i] {
					currentProblemNum = i
				}
			}

			if currentProblemNum < len(part2Problems) {
				if part2Problems[currentProblemNum] == nil {
					part2Problems[currentProblemNum] = make([][]rune, 0)
				}

				for len(part2Problems[currentProblemNum]) <= colIndex {
					part2Problems[currentProblemNum] = append(part2Problems[currentProblemNum], make([]rune, 0))
				}

				if char >= '0' && char <= '9' {
					part2Problems[currentProblemNum][colIndex] = append(part2Problems[currentProblemNum][colIndex], char)
				}
			}
		}
	}

	grandTotal := 0
	grandTotal2 := 0

	for problemNum, problem := range part1Problems {
		numbers := make([]int, len(problem))

		for i, number := range problem {
			numberString := string(number)
			v, _ := strconv.Atoi(numberString)
			numbers[i] = v
		}

		operator := operators[problemNum]
		grandTotal += solveColumn(operator, numbers)
	}

	for problemNum, problem := range part2Problems {
		finalNumbers := make([]int, 0)

		for _, digitRunes := range problem {
			numberString := string(digitRunes)

			if numberString == "" {
				continue
			}

			v, _ := strconv.Atoi(numberString)
			finalNumbers = append(finalNumbers, v)
		}

		operator := operators[problemNum]
		grandTotal2 += solveColumn(operator, finalNumbers)
	}

	fmt.Printf("Total (P1): %d\n", grandTotal)
	fmt.Printf("Total (P2): %d\n", grandTotal2)
}

func solveColumn(problemFn Operator, values []int) int {
	total := 0

	for _, value := range values {
		total = solveProblem(problemFn, value, total)
	}

	return total
}

func solveProblem(problemFn Operator, num, currentNum int) int {
	if currentNum == 0 {
		return num
	}

	return problemFn(currentNum, num)
}
