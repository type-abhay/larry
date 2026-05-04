package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"lyric-emb/api"
	"lyric-emb/audio"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("\n!!! CRITICAL ERROR !!!\n")
			fmt.Printf("The program encountered a corrupt file or a library crash.\n")
			fmt.Printf("Reason: %v\n", r)
			fmt.Println("Press Enter to exit...")
			bufio.NewReader(os.Stdin).ReadString('\n')
		}
	}()
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
	fmt.Println("\nSelective Embed Mode")
	fmt.Println("Available commands:")
	fmt.Println("  search <keyword> : List all files containing the keyword")
	fmt.Println("  info <filename>  : View full tags for a specific file")
	fmt.Println("  get <filename>   : Fetch API lyrics and embed them")
	fmt.Println("  exit (or back)   : Return to the main menu")
	fmt.Println("Note: You don't need the full name! 'info baldur' will find 'Baldur's Gate.mp3'")

	for {
		fmt.Print("\nlarry> ")
		input := getInput(reader)
		if input == "" {
			continue
		}

		// Split the input into the command and the arguments
		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		// Rejoin the rest of the string for the search term and remove quotes
		arg := strings.Join(parts[1:], " ")
		arg = strings.Trim(arg, "\"")

		switch command {
		case "back", "exit":
			return

		case "search":
			if arg == "" {
				fmt.Println(" ! Usage: search <keyword>")
				continue
			}
			term := strings.ToLower(arg)
			count := 0
			for _, f := range files {
				base := strings.ToLower(filepath.Base(f))
				if strings.Contains(base, term) {
					fmt.Printf(" - %s\n", filepath.Base(f))
					count++
				}
			}
			fmt.Printf(" > Found %d matching files.\n", count)

		case "info", "get":
			if arg == "" {
				fmt.Printf(" ! Usage: %s <filename>\n", command)
				continue
			}

			// Find matching files based on the argument
			term := strings.ToLower(arg)
			var matches []string
			for _, f := range files {
				base := strings.ToLower(filepath.Base(f))
				if strings.Contains(base, term) {
					matches = append(matches, f)
				}
			}

			// Handle the match results
			if len(matches) == 0 {
				fmt.Println(" ! No files found matching that name.")
				continue
			} else if len(matches) > 1 {
				fmt.Println(" ! Multiple files found. Please be more specific:")
				for _, m := range matches {
					fmt.Printf("   - %s\n", filepath.Base(m))
				}
				continue
			}

			// Exactly one match found!
			targetPath := matches[0]

			if command == "info" {
				meta, err := audio.ReadMetadata(targetPath)
				if err != nil {
					fmt.Printf(" ! Error reading metadata: %v\n", err)
					continue
				}
				fmt.Println("\n--- META DATA INFO ---")
				fmt.Printf("File:     %s\n", filepath.Base(targetPath))
				fmt.Printf("Title:    %s\n", meta.Title)
				fmt.Printf("Artist:   %s\n", meta.Artist)
				fmt.Printf("Album:    %s\n", meta.Album)
				fmt.Printf("Format:   %s\n", meta.Format)
				fmt.Printf("Duration: %ds\n", meta.Duration)
				fmt.Printf("Lyrics:   %t (Embedded)\n", meta.HasLyrics)
			} else {
				// Trigger the API and Embedding flow
				processSingleFile(targetPath, reader)
			}

		default:
			fmt.Println(" ! Unknown command. Try: search, info, get, exit")
		}
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
	// ... (Keep the metadata display and API call logic from before) ...

	fmt.Printf("\nAPI Response (Duration Match: %ds):\n", int(lyrics.Duration))
	fmt.Println("1. Embed Synced Lyrics")
	fmt.Println("2. Embed Plain Lyrics")
	fmt.Println("3. Cancel")
	fmt.Print("Choice: ")

	// We capture the input here
	choice := getInput(reader)

	var lyricText string
	var isSynced bool

	// 1. Using the 'choice' variable
	switch choice {
	case "1":
		if lyrics.SyncedLyrics == "" {
			fmt.Println(" ! Error: The API did not return synced lyrics for this track.")
			return
		}
		lyricText = lyrics.SyncedLyrics
		isSynced = true
	case "2":
		if lyrics.PlainLyrics == "" {
			fmt.Println(" ! Error: The API did not return plain lyrics for this track.")
			return
		}
		lyricText = lyrics.PlainLyrics
		isSynced = false
	case "3":
		fmt.Println(" > Operation cancelled.")
		return
	default:
		fmt.Println(" ! Invalid selection. Returning to list.")
		return
	}

	// 2. The Final Payoff: Embedding the data
	fmt.Printf(" > Embedding %s lyrics into file...\n", map[bool]string{true: "synced", false: "plain"}[isSynced])

	err = audio.EmbedLyrics(path, lyricText, isSynced)
	if err != nil {
		fmt.Printf(" [X] Failed to embed: %v\n", err)
	} else {
		fmt.Println(" [✓] Success! Lyrics are now part of the file.")
	}
}

// Helper to handle repetitive input cleaning
func getInput(r *bufio.Reader) string {
	input, _ := r.ReadString('\n')
	return strings.TrimSpace(input)
}

func runBulkProcess(files []string, reader *bufio.Reader) {
	// Original loop logic goes here
}
