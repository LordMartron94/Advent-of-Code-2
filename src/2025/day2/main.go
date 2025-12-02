package main

import (
	aocshared "aoc_shared"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ------------------ PRIMITIVE CHECK ------------------

// isSequencePrimitive determines if a sequence 'S' of length 'L' is fundamentally primitive.
// A sequence is primitive if it cannot be represented as a smaller sub-sequence 's'
// repeated 'k' times.
//
// Example:
// 1212 (L=4) is NOT primitive because it is 12 (L=2) repeated 2 times.
// 1234 (L=4) IS primitive.
//
// This check is required for Part 2 to prevent double-counting.
// We must count 121212 as (12)^3, not as (1212) followed by partial overlap.
func isSequencePrimitive(sequence, sequenceLength int) bool {
	// We iterate through all valid sub-sequence lengths (lSub).
	// lSub must be a strict divisor of the total sequenceLength.
	for lSub := 1; lSub <= sequenceLength/2; lSub++ {
		if sequenceLength%lSub != 0 {
			continue
		}

		// kSub is the number of times the sub-sequence would need to repeat.
		kSub := sequenceLength / lSub

		// Calculate the geometric series multiplier to construct the full sequence
		// from the sub-sequence mathematically.
		// Multiplier M = 10^((k-1)lSub) + ... + 10^lSub + 1
		powerOfTenLSub := int(math.Pow(10, float64(lSub)))

		multiplier := 1
		currentPower := 1
		for i := 1; i < kSub; i++ {
			currentPower *= powerOfTenLSub
			multiplier += currentPower
		}

		// If the sequence is perfectly divisible by this multiplier, it implies
		// the sequence has the structure SubSeq * M, which means it is repeating.
		if sequence%multiplier == 0 {
			subSequence := sequence / multiplier

			// Verification: Ensure the sub-sequence doesn't have leading zeros.
			// The sub-sequence must be >= 10^(lSub-1).
			minSubSequenceValue := int(math.Pow(10, float64(lSub-1)))

			if subSequence >= minSubSequenceValue {
				return false // The sequence is composite (repeating); not primitive.
			}
		}
	}

	return true
}

// ------------------ MAIN EXECUTION ------------------

func main() {
	input := aocshared.GetInput(2025, 2)
	ranges := strings.Split(input, ",")

	sumInvalidIDsP1 := 0
	sumInvalidIDsP2 := 0

	for _, rangeElement := range ranges {
		trimmed := strings.TrimSpace(rangeElement)
		parts := strings.Split(trimmed, "-")
		p1, p2 := parts[0], parts[1]

		// Part 1: Strict Seq^2 (k=2, k=2).
		sumInvalidIDsForPart(p1, p2, 2, 2, &sumInvalidIDsP1)

		// Part 2: Seq^k where k >= 2.
		// Passing 0 for maxK triggers the computation of the physical limit (N_max).
		sumInvalidIDsForPart(p1, p2, 2, 0, &sumInvalidIDsP2)
	}

	fmt.Printf("Sum Invalid IDs P1: %d\n", sumInvalidIDsP1)
	fmt.Printf("Sum Invalid IDs P2: %d\n", sumInvalidIDsP2)
}

// ------------------ SUMMATION LOGIC ------------------

// sumInvalidIDsForPart iterates over valid structure definitions (N and k)
// rather than iterating the range itself.
//
// N = Total Digits
// k = Repetition Factor
// L = Sequence Length (N / k)
func sumInvalidIDsForPart(startIDString, endIDString string, minK, maxK int, sumPointer *int) {
	startIDInt, err := strconv.Atoi(startIDString)
	if err != nil {
		panic(fmt.Errorf("error converting start ID: %w", err))
	}

	endIDInt, err := strconv.Atoi(endIDString)
	if err != nil {
		panic(fmt.Errorf("error converting end ID: %w", err))
	}

	minTotalDigits := len(startIDString)
	maxTotalDigits := len(endIDString)

	// Determine the hard upper limit for k.
	// The maximum k occurs when L=1, so k_max = N_max.
	computedMaxK := maxTotalDigits

	if maxK == 0 || maxK > computedMaxK {
		maxK = computedMaxK
	}

	// Loop 1: Iterate all possible Total Lengths (N) that overlap the range.
	for totalDigits := minTotalDigits; totalDigits <= maxTotalDigits; totalDigits++ {

		// Loop 2: Iterate all possible Repetition Factors (k).
		for k := minK; k <= maxK; k++ {

			// Constraint: N must be divisible by k to form valid integer sequences.
			if totalDigits%k != 0 {
				continue
			}

			sequenceLength := totalDigits / k

			// Constraint: Sequence length L must be at least 1.
			if sequenceLength < 1 {
				continue
			}

			// We delegate the sequence generation to 'process'.
			// We pass 'k' for both min and max to force checking this specific structure.
			process(startIDInt, endIDInt, sequenceLength, k, k, sumPointer)
		}
	}
}

// ------------------ GENERATION LOGIC ------------------

func process(startIDInt, endIDInt, sequenceLength, minPatternOccurrence, maxPatternOccurrence int, sumPointer *int) {
	// powerOfTenL is the shift factor (10^L) for this sequence length.
	powerOfTenL := int(math.Pow(10, float64(sequenceLength)))

	// Define the search space for the sequence candidate (e.g., 100 to 999).
	minSequenceValue := int(math.Pow(10, float64(sequenceLength-1)))
	maxSequenceValue := powerOfTenL - 1

	for sequenceCandidate := minSequenceValue; sequenceCandidate <= maxSequenceValue; sequenceCandidate++ {

		// CRITICAL FILTER:
		// We only proceed if the sequence itself is primitive.
		// If sequenceCandidate is 1212, it is not primitive (it's 12 repeated).
		// We skip it here because it will be caught when process is called with L=2.
		if !isSequencePrimitive(sequenceCandidate, sequenceLength) {
			continue
		}

		// Check the specific repetition factor k.
		for k := minPatternOccurrence; k <= maxPatternOccurrence; k++ {

			invalidIDCandidate := generateID(k, sequenceCandidate, powerOfTenL)

			// Optimization: Logic is monotonic. If we exceed endID, we stop.
			if invalidIDCandidate > endIDInt {
				break
			}

			if startIDInt <= invalidIDCandidate && invalidIDCandidate <= endIDInt {
				*sumPointer += invalidIDCandidate
			}
		}
	}
}

// generateID constructs the full ID arithmetically: Seq^k.
// Method: Iterative shifting and addition.
func generateID(k, sequenceCandidate, necessaryPowerOfTen int) int {
	id := sequenceCandidate

	// We start with one sequence, then shift and add k-1 times.
	for i := 1; i < k; i++ {
		id = id * necessaryPowerOfTen
		id += sequenceCandidate
	}
	return id
}
