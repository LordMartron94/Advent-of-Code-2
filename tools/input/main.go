package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const configPath = "config.toml"

type Settings struct {
	Key string `toml:"key"`
}

func main() {
	settings := loadConfig()
	year, day := getUserInput()

	url := fmt.Sprintf("https://adventofcode.com/%d/day/%d/input", year, day)

	fmt.Printf("Fetching input for Year: %d, Day: %d...\n", year, day)
	data, err := fetchInput(url, settings.Key)
	if err != nil {
		log.Fatalf("Failed to fetch input: %v", err)
	}

	if err := saveInput(year, day, data); err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}

	fmt.Println("Done.")
}

// loadConfig isolates the configuration logic.
func loadConfig() Settings {
	var settings Settings
	if _, err := toml.DecodeFile(configPath, &settings); err != nil {
		// Panic here is acceptable as the app cannot start without config
		panic(fmt.Errorf("failed to load config: %w", err))
	}
	return settings
}

// getUserInput handles the CLI interaction.
func getUserInput() (int, int) {
	var year, day int

	fmt.Print("Enter Year (e.g., 2023): ")
	_, err := fmt.Scan(&year)
	if err != nil {
		log.Fatal("Invalid year input")
	}

	fmt.Print("Enter Day (1-25): ")
	_, err = fmt.Scan(&day)
	if err != nil {
		log.Fatal("Invalid day input")
	}

	return year, day
}

// saveInput handles directory creation and file writing.
func saveInput(year, day int, data []byte) error {
	dirPath := filepath.Join("src", fmt.Sprintf("%d", year), fmt.Sprintf("day%d", day))

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	filePath := filepath.Join(dirPath, "input.txt")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Saved input to: %s\n", filePath)
	return nil
}

// fetchInput remains focused solely on the HTTP transport.
func fetchInput(url, sessionID string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: sessionID,
	})

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
