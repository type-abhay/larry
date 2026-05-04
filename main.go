package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"lyric-emb/api"
	"lyric-emb/audio"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	targetDir := "."
	if len(os.Args) > 1 {
		targetDir = os.Args[1]
	}

	// 1. Initial Scan
	files, err := audio.ScanDirectory(targetDir)
	if err != nil {
		log.Fatalf("Critical Error: %v", err)
	}

	for {
		fmt.Println("\n--- MAIN MENU ---")
		fmt.Println("1. Bulk Process (All files)")
		fmt.Println("2. Selective Process (Choose a file)")
		fmt.Println("3. Exit")
		fmt.Print("Selection: ")

		choice := getInput(reader)

		switch choice {
		case "1":
			runBulkProcess(files, reader)
		case "2":
			runSelectiveProcess(files, reader)
		case "3":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

// runSelectiveProcess handles the one-by-one logic you requested
func runSelectiveProcess(files []string, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- AVAILABLE TRACKS ---")
		// We read basic info for the list
		for i, path := range files {
			meta, _ := audio.ReadMetadata(path)
			fmt.Printf("[%d] %s (%ds)\n", i+1, meta.Title, meta.Duration)
		}
		fmt.Println("[0] Back to Main Menu")
		fmt.Print("\nSelect track number: ")

		input := getInput(reader)
		if input == "0" {
			return
		}

		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(files) {
			fmt.Println("Invalid track number.")
			continue
		}

		// Process the specific chosen file
		processSingleFile(files[idx-1], reader)
	}
}

func processSingleFile(path string, reader *bufio.Reader) {
	// 1. Display Full Tags
	meta, err := audio.ReadMetadata(path)
	if err != nil {
		fmt.Printf("Error reading tags: %v\n", err)
		return
	}

	fmt.Println("\n--- CURRENT METADATA ---")
	fmt.Printf("Title:  %s\n", meta.Title)
	fmt.Printf("Artist: %s\n", meta.Artist)
	fmt.Printf("Album:  %s\n", meta.Album)
	fmt.Printf("Year:   %s\n", meta.Year)
	fmt.Printf("Genre:  %s\n", meta.Genre)
	fmt.Printf("Lyrics: %t (Found embedded)\n", meta.HasLyrics)

	// 2. Confirmation
	fmt.Print("\nSearch lyrics via API? (y/n): ")
	if strings.ToLower(getInput(reader)) != "y" {
		return
	}

	// 3. API Call & Display Options
	lyrics, err := api.GetLyrics(meta)
	if err != nil {
		fmt.Printf("API Error: %v\n", err)
		return
	}

	fmt.Printf("\nAPI Response (Duration Match: %ds):\n", lyrics.Duration)
	fmt.Println("1. Embed Synced Lyrics")
	fmt.Println("2. Embed Plain Lyrics")
	fmt.Println("3. Cancel")
	fmt.Print("Choice: ")

	choice := getInput(reader)
	// ... logic to call audio.EmbedLyrics based on choice ...
	// (Similar to previous implementation, but for this single path)
}

// Helper to handle repetitive input cleaning
func getInput(r *bufio.Reader) string {
	input, _ := r.ReadString('\n')
	return strings.TrimSpace(input)
}

func runBulkProcess(files []string, reader *bufio.Reader) {
	// Original loop logic goes here
}
