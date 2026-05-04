package main

import (
	"bufio"
	"fmt"
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

	// 1. Setup the default "current" context
	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	fmt.Println("Scanning initial directory...")
	currentFiles, err := audio.ScanDirectory(baseDir)
	if err != nil {
		fmt.Printf("Error scanning directory: %v\n", err)
	}

	fmt.Println("\n======================================")
	fmt.Println("WELCOME TO LARRY")
	fmt.Println("======================================")
	fmt.Printf("Loaded %d files from context: '%s'\n", len(currentFiles), baseDir)
	fmt.Printf(" Please Visit `https://github.com/type-abhay/larry#` to know how to use it!")
	printHelp()

	// 2. The Unified Command Loop
	for {
		fmt.Print("\nlarry> ")
		input := getInput(reader)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		arg := strings.Join(parts[1:], " ")
		arg = strings.Trim(arg, "\"")

		switch command {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return

		case "help":
			printHelp()

		case "search":
			if arg == "" {
				fmt.Println(" ! Usage: search <keyword>")
				continue
			}
			term := strings.ToLower(arg)
			count := 0
			for _, f := range currentFiles {
				if strings.Contains(strings.ToLower(filepath.Base(f)), term) {
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

			term := strings.ToLower(arg)
			var matches []string
			for _, f := range currentFiles {
				if strings.Contains(strings.ToLower(filepath.Base(f)), term) {
					matches = append(matches, f)
				}
			}

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

			targetPath := matches[0]

			if command == "info" {
				displayInfo(targetPath)
			} else {
				processSingleFile(targetPath, reader)
			}

		case "bulk":
			if arg == "" {
				fmt.Println(" ! Usage: bulk <folder> (or 'bulk current')")
				continue
			}

			targetDir := arg
			if arg == "current" {
				targetDir = baseDir
			}

			runBulkProcess(targetDir, reader)

		default:
			fmt.Println(" ! Unknown command. Type 'help' to see available commands.")
		}
	}
}

// printHelp displays the available commands
func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  search <keyword> : List all files containing the keyword")
	fmt.Println("  info <filename>  : View full tags for a specific file")
	fmt.Println("  get <filename>   : Fetch API lyrics and embed interactively")
	fmt.Println("  bulk <folder>    : Auto-embed lyrics for a whole folder (use 'bulk current' for default)")
	fmt.Println("  help             : Show this menu")
	fmt.Println("  exit             : Quit the program")
}

// displayInfo handles the purely visual tag output
func displayInfo(path string) {
	meta, err := audio.ReadMetadata(path)
	if err != nil {
		fmt.Printf(" ! Error reading metadata: %v\n", err)
		return
	}
	fmt.Println("\n--- FILE INFO ---")
	fmt.Printf("File:     %s\n", filepath.Base(path))
	fmt.Printf("Title:    %s\n", meta.Title)
	fmt.Printf("Artist:   %s\n", meta.Artist)
	fmt.Printf("Album:    %s\n", meta.Album)
	fmt.Printf("Format:   %s\n", meta.Format)
	fmt.Printf("Duration: %ds\n", meta.Duration)
	fmt.Printf("Lyrics:   %t (Embedded)\n", meta.HasLyrics)
}

// processSingleFile is your existing Interactive flow (omitted some unchanged fmt code for brevity)
func processSingleFile(path string, reader *bufio.Reader) {
	displayInfo(path)
	meta, _ := audio.ReadMetadata(path)

	fmt.Print("\nSearch lyrics via API? (y/n): ")
	if strings.ToLower(getInput(reader)) != "y" {
		return
	}

	lyrics, err := api.GetLyrics(meta)
	if err != nil {
		fmt.Printf("API Error: %v\n", err)
		return
	}

	fmt.Printf("\nAPI Response (Duration Match: %ds):\n", int(lyrics.Duration))
	fmt.Println("1. Embed Synced Lyrics")
	fmt.Println("2. Embed Plain Lyrics")
	fmt.Println("3. Cancel")
	fmt.Print("Choice: ")

	choice := getInput(reader)
	var lyricText string
	var isSynced bool

	switch choice {
	case "1":
		if lyrics.SyncedLyrics == "" {
			fmt.Println(" ! Error: No synced lyrics available.")
			return
		}
		lyricText = lyrics.SyncedLyrics
		isSynced = true
	case "2":
		if lyrics.PlainLyrics == "" {
			fmt.Println(" ! Error: No plain lyrics available.")
			return
		}
		lyricText = lyrics.PlainLyrics
		isSynced = false
	default:
		fmt.Println(" > Operation cancelled.")
		return
	}

	err = audio.EmbedLyrics(path, lyricText, isSynced)
	if err != nil {
		fmt.Printf(" [X] Failed to embed: %v\n", err)
	} else {
		fmt.Println(" [✓] Success! Lyrics embedded.")
	}
}

// runBulkProcess handles the new automated, multi-file flow
func runBulkProcess(folder string, reader *bufio.Reader) {
	fmt.Printf("\nScanning '%s' for bulk processing...\n", folder)
	files, err := audio.ScanDirectory(folder)
	if err != nil {
		fmt.Printf(" ! Error accessing folder: %v\n", err)
		return
	}

	count := len(files)
	if count == 0 {
		fmt.Println(" ! No supported audio files found in that directory.")
		return
	}

	fmt.Printf(" > Found %d files.\n", count)
	if count > 20 {
		fmt.Println(" !!! WARNING !!!")
		fmt.Println(" You are about to hit the API and modify a large number of files.")
		fmt.Println(" Doing this may trigger API rate limits or take a significant amount of time.")
	}

	fmt.Print(" Proceed with bulk embedding? (y/n): ")
	if strings.ToLower(getInput(reader)) != "y" {
		fmt.Println(" > Bulk operation cancelled.")
		return
	}

	var notFound []string
	successCount := 0

	fmt.Println("\nStarting Bulk Process...")
	for i, path := range files {
		filename := filepath.Base(path)
		fmt.Printf("[%d/%d] Processing: %s... ", i+1, count, filename)

		meta, err := audio.ReadMetadata(path)
		if err != nil {
			fmt.Println("Error reading tags (Skipped)")
			notFound = append(notFound, filename)
			continue
		}

		lyrics, err := api.GetLyrics(meta)
		if err != nil {
			fmt.Println("Not found on API")
			notFound = append(notFound, filename)
			continue
		}

		// Auto-Selection Logic: Synced > Plain > Skip
		var lyricText string
		var isSynced bool

		if lyrics.SyncedLyrics != "" {
			lyricText = lyrics.SyncedLyrics
			isSynced = true
		} else if lyrics.PlainLyrics != "" {
			lyricText = lyrics.PlainLyrics
			isSynced = false
		} else {
			fmt.Println("No lyrics data available")
			notFound = append(notFound, filename)
			continue
		}

		// Embed
		err = audio.EmbedLyrics(path, lyricText, isSynced)
		if err != nil {
			fmt.Println("Failed to write to file")
			notFound = append(notFound, filename)
		} else {
			fmt.Println("✓ Success")
			successCount++
		}
	}

	fmt.Println("\n=== BULK RUN RESULTS ===")
	fmt.Printf("Successfully embedded: %d\n", successCount)
	fmt.Printf("Failed / Not Found: %d\n", len(notFound))

	if len(notFound) > 0 {
		fmt.Println("\nMissing tracks:")
		for _, name := range notFound {
			fmt.Printf(" - %s\n", name)
		}
	}
}

// Helper to handle repetitive input cleaning
func getInput(r *bufio.Reader) string {
	input, _ := r.ReadString('\n')
	return strings.TrimSpace(input)
}
