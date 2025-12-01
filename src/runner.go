package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) > 2 {
		year, _ := strconv.Atoi(os.Args[1])
		day, _ := strconv.Atoi(os.Args[2])
		if err := runSolution(year, day); err != nil {
			log.Fatalf("Execution failed: %v", err)
		}
	} else {
		year, day := getUserInput()
		if err := runSolution(year, day); err != nil {
			log.Fatalf("Execution failed: %v", err)
		}
	}
}

func runSolution(year, day int) error {
	targetDir := filepath.Join("src", fmt.Sprintf("%d", year), fmt.Sprintf("day%d", day))

	scriptPath := filepath.Join(targetDir, "main.go")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("solution file not found at: %s", scriptPath)
	}

	fmt.Printf("--- Running Year %d Day %d ---\n", year, day)

	cmd := exec.Command("go", "run", "main.go")

	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func getUserInput() (int, int) {
	var year, day int

	fmt.Print("Run Year: ")
	if _, err := fmt.Scan(&year); err != nil {
		log.Fatal("Invalid year input")
	}

	fmt.Print("Run Day: ")
	if _, err := fmt.Scan(&day); err != nil {
		log.Fatal("Invalid day input")
	}

	return year, day
}
