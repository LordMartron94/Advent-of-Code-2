package main

import (
	aocshared "aoc_shared"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const test = false

// For this one I have failed. It was completely out of my knowledge.
// I have used LLMs to solve the entire problem and have done the research and learning afterward.

// The part 2 solution I got from my competitor, in Python.

func main() {
	day10 := Day10{}

	aocshared.DebugAndLogTasks(
		"2025 day 10",
		aocshared.Task{Name: "Gather Input", Run: day10.GatherInput},
		aocshared.Task{Name: "Parse Input", Run: day10.ParseInput},
		aocshared.Task{Name: "Solve", Run: day10.Solve},
		aocshared.Task{Name: "Print Results", Run: day10.PrintResults},
	)
}

type Day10 struct {
	input            string
	machines         []Machine
	fewestPressesPt1 int
	fewestPressesPt2 int64
}

type Machine struct {
	IndicatorLightDiagram string
	ButtonWiring          []string
	JoltageRequirements   string
}

func (d *Day10) GatherInput() {
	if test {
		d.input = aocshared.GetTestInput(2025, 10)
	} else {
		d.input = aocshared.GetInput(2025, 10)
	}
}

var MacroRegex = regexp.MustCompile(`^\[([.#]+)\]\s+((?:\(\d+(?:,\d+)*\)\s*)+)\s+\{(\d+(?:,\d+)*)\}$`)
var MicroRegex = regexp.MustCompile(`\(([^)]+)\)`)

func extractMajorComponents(line string) (indicator string, schematicsRaw string, joltageRaw string, err error) {
	matches := MacroRegex.FindStringSubmatch(line)
	if len(matches) != 4 {
		return "", "", "", fmt.Errorf("line does not match expected format: %s", line)
	}
	return matches[1], matches[2], matches[3], nil
}

func tokenizeSchematics(rawString string) []string {
	matches := MicroRegex.FindAllStringSubmatch(rawString, -1)
	var tokens []string
	for _, match := range matches {
		if len(match) > 1 {
			tokens = append(tokens, match[1])
		}
	}
	return tokens
}

func (d *Day10) ParseInput() {
	lines := strings.Split(strings.TrimSpace(d.input), "\n")
	machines := make([]Machine, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indicator, schematicsRaw, joltageRaw, err := extractMajorComponents(line)
		if err != nil {
			fmt.Println("Error parsing line:", line, err)
			continue
		}
		schematicsList := tokenizeSchematics(schematicsRaw)
		machines = append(machines, Machine{
			IndicatorLightDiagram: indicator,
			ButtonWiring:          schematicsList,
			JoltageRequirements:   joltageRaw,
		})
	}
	d.machines = machines
}

func (d *Day10) Solve() {
	d.fewestPressesPt1 = 0
	d.fewestPressesPt2 = 0

	for _, machine := range d.machines {
		d.fewestPressesPt1 += solveMachine(machine)
	}
}

// --- PART 1 Logic ---

type bitset uint64

func solveMachine(machine Machine) int {
	requiredState := parseDiagram(machine.IndicatorLightDiagram)
	buttonsParsed := make([]bitset, len(machine.ButtonWiring))

	for i, wiring := range machine.ButtonWiring {
		numbers := strings.Split(wiring, ",")
		parsed := make([]int, len(numbers))
		for j, number := range numbers {
			v, _ := strconv.Atoi(number)
			parsed[j] = v
		}
		buttonsParsed[i] = parseButton(parsed)
	}

	matrix := buildMatrix(machine, buttonsParsed)
	rref, exists := reduceToRREF(matrix, requiredState, len(buttonsParsed))

	if !exists {
		return 0
	}

	nullSpaceBasis := findNullSpaceBasis(rref, len(buttonsParsed))
	return findMinWeightSolution(rref, nullSpaceBasis, len(buttonsParsed))
}

func buildMatrix(machine Machine, buttonsParsed []bitset) []bitset {
	numLights := len(machine.IndicatorLightDiagram)
	matrix := make([]bitset, numLights)
	for buttonIndex, btnMask := range buttonsParsed {
		for lightIndex := 0; lightIndex < numLights; lightIndex++ {
			if (btnMask>>lightIndex)&1 == 1 {
				matrix[lightIndex] |= (1 << buttonIndex)
			}
		}
	}
	return matrix
}

func reduceToRREF(A_matrix []bitset, target bitset, numButtons int) ([]bitset, bool) {
	N := len(A_matrix)
	M := numButtons
	AugCol := M
	augmented := make([]bitset, N)
	for i := 0; i < N; i++ {
		b_i := (target >> i) & 1
		augmented[i] = A_matrix[i] | (b_i << AugCol)
	}

	pivotRow := 0
	pivotCol := 0
	pivotColsMap := make(map[int]int)

	for pivotRow < N && pivotCol < M {
		i := pivotRow
		for i < N && (augmented[i]>>pivotCol)&1 == 0 {
			i++
		}
		if i < N {
			augmented[pivotRow], augmented[i] = augmented[i], augmented[pivotRow]
			pivotColsMap[pivotRow] = pivotCol
			for j := pivotRow + 1; j < N; j++ {
				if (augmented[j]>>pivotCol)&1 == 1 {
					augmented[j] ^= augmented[pivotRow]
				}
			}
			pivotRow++
			pivotCol++
		} else {
			pivotCol++
		}
	}

	for i := pivotRow; i < N; i++ {
		if (augmented[i]>>AugCol)&1 == 1 {
			return nil, false
		}
	}

	for r := pivotRow - 1; r >= 0; r-- {
		pivotCol, ok := pivotColsMap[r]
		if !ok {
			continue
		}
		for i := 0; i < r; i++ {
			if (augmented[i]>>pivotCol)&1 == 1 {
				augmented[i] ^= augmented[r]
			}
		}
	}
	return augmented, true
}

func findNullSpaceBasis(augmentedRREF []bitset, numButtons int) []bitset {
	N := len(augmentedRREF)
	M := numButtons
	pivotCols := make(map[int]struct{})
	var freeCols []int
	pivotRow := 0
	for col := 0; col < M; col++ {
		if pivotRow < N && (augmentedRREF[pivotRow]>>col)&1 == 1 {
			pivotCols[col] = struct{}{}
			pivotRow++
		} else {
			freeCols = append(freeCols, col)
		}
	}

	k := len(freeCols)
	nullSpaceBasis := make([]bitset, k)
	for basisIdx, freeCol := range freeCols {
		var k_vector bitset = 0
		k_vector |= 1 << freeCol
		for r := 0; r < N; r++ {
			pivotCol := -1
			for c := 0; c < M; c++ {
				if (augmentedRREF[r]>>c)&1 == 1 {
					pivotCol = c
					break
				}
			}
			if pivotCol != -1 {
				var dependentValue bitset = 0
				for _, freeIdx := range freeCols {
					coefficient := (augmentedRREF[r] >> freeIdx) & 1
					x_free_val := (k_vector >> freeIdx) & 1
					if coefficient == 1 && x_free_val == 1 {
						dependentValue ^= 1
					}
				}
				if dependentValue == 1 {
					k_vector |= 1 << pivotCol
				}
			}
		}
		nullSpaceBasis[basisIdx] = k_vector
	}
	return nullSpaceBasis
}

func findMinWeightSolution(augmentedRREF []bitset, nullSpaceBasis []bitset, numButtons int) int {
	N := len(augmentedRREF)
	M := numButtons
	AugCol := M
	k := len(nullSpaceBasis)
	var xp bitset = 0
	pivotRow := 0
	for col := 0; col < M; col++ {
		if pivotRow < N && (augmentedRREF[pivotRow]>>col)&1 == 1 {
			if (augmentedRREF[pivotRow]>>AugCol)&1 == 1 {
				xp |= 1 << col
			}
			pivotRow++
		}
	}
	minWeight := -1
	for i := 0; i < 1<<k; i++ {
		var current_homogeneous_solution bitset = 0
		for l := 0; l < k; l++ {
			if (i>>l)&1 == 1 {
				current_homogeneous_solution ^= nullSpaceBasis[l]
			}
		}
		total_solution := xp ^ current_homogeneous_solution
		weight := countSetBits(total_solution)
		if minWeight == -1 || weight < minWeight {
			minWeight = weight
		}
	}
	return minWeight
}

func countSetBits(b bitset) int {
	count := 0
	for i := 0; i < 64; i++ {
		if (b>>i)&1 == 1 {
			count++
		}
	}
	return count
}

func parseDiagram(diagram string) bitset {
	var mask bitset = 0
	for i, c := range diagram {
		if c == '#' {
			mask |= 1 << i
		}
	}
	return mask
}

func parseButton(indices []int) bitset {
	var mask bitset = 0
	for _, idx := range indices {
		mask |= 1 << idx
	}
	return mask
}

// --- Fraction Implementation ---

type Fraction struct {
	num int64
	den int64
}

func Gcd(a, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func (f *Fraction) Simplify() {
	if f.den < 0 {
		f.num = -f.num
		f.den = -f.den
	}
	if f.num == 0 {
		f.den = 1
		return
	}
	g := Gcd(f.num, f.den)
	f.num /= g
	f.den /= g
}

func NewFraction(n, d int64) Fraction {
	if d == 0 {
		panic("Division by zero in NewFraction")
	}
	f := Fraction{num: n, den: d}
	f.Simplify()
	return f
}

func (f Fraction) Add(other Fraction) Fraction {
	n1 := big.NewInt(f.num)
	n2 := big.NewInt(other.num)
	d1 := big.NewInt(f.den)
	d2 := big.NewInt(other.den)

	term1 := new(big.Int).Mul(n1, d2)
	term2 := new(big.Int).Mul(n2, d1)
	numRes := new(big.Int).Add(term1, term2)
	denRes := new(big.Int).Mul(d1, d2)

	return bigIntToFraction(numRes, denRes)
}

func (f Fraction) Sub(other Fraction) Fraction {
	n1 := big.NewInt(f.num)
	n2 := big.NewInt(other.num)
	d1 := big.NewInt(f.den)
	d2 := big.NewInt(other.den)

	term1 := new(big.Int).Mul(n1, d2)
	term2 := new(big.Int).Mul(n2, d1)
	numRes := new(big.Int).Sub(term1, term2)
	denRes := new(big.Int).Mul(d1, d2)

	return bigIntToFraction(numRes, denRes)
}

func (f Fraction) Mul(other Fraction) Fraction {
	n1 := big.NewInt(f.num)
	n2 := big.NewInt(other.num)
	d1 := big.NewInt(f.den)
	d2 := big.NewInt(other.den)

	numRes := new(big.Int).Mul(n1, n2)
	denRes := new(big.Int).Mul(d1, d2)

	return bigIntToFraction(numRes, denRes)
}

func (f Fraction) Div(other Fraction) Fraction {
	if other.num == 0 {
		panic("Division by zero fraction")
	}
	n1 := big.NewInt(f.num)
	n2 := big.NewInt(other.num)
	d1 := big.NewInt(f.den)
	d2 := big.NewInt(other.den)

	numRes := new(big.Int).Mul(n1, d2)
	denRes := new(big.Int).Mul(d1, n2)

	return bigIntToFraction(numRes, denRes)
}

func bigIntToFraction(n, d *big.Int) Fraction {
	if d.Sign() == 0 {
		panic("BigInt division by zero")
	}
	g := new(big.Int).GCD(nil, nil, new(big.Int).Abs(n), new(big.Int).Abs(d))
	if g.Sign() != 0 {
		n.Div(n, g)
		d.Div(d, g)
	}
	if !n.IsInt64() || !d.IsInt64() {
		panic("Fraction overflow")
	}
	return NewFraction(n.Int64(), d.Int64())
}

func (f Fraction) IsZero() bool {
	return f.num == 0
}
func (f Fraction) IsNegative() bool {
	return f.num < 0
}

func (d *Day10) PrintResults() {
	fmt.Printf("Solution Pt1: %d\n", d.fewestPressesPt1)
	fmt.Printf("Solution Pt2: %d\n", d.fewestPressesPt2)
}
