package aocshared

import (
	"fmt"
	"os"
)

func GetInput(year, day int) string {
	content, err := os.ReadFile("input.txt")
	if err != nil {
		panic(err)
	}

	return string(content)
}

func GetTestInput(year, day int) string {
	content, err := os.ReadFile("test_input.txt")
	if err != nil {
		panic(err)
	}

	return string(content)
}

func GetTestInputN(year, day, n int) string {
	content, err := os.ReadFile(fmt.Sprintf("test_input_%d.txt", n))
	if err != nil {
		panic(err)
	}

	return string(content)
}
