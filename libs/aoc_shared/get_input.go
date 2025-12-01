package aocshared

import (
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
