# Larry: CLI Based Lyric Embedder
<table>
  <tr>
    <td><img src="icon.ico" width="120" alt="App icon"/></td>
    <td><h1>Larry</h1></td>
  </tr>
</table>

[![Get it on GitHub](https://img.shields.io/badge/Get%20it%20on-GitHub-blue?style=for-the-badge&logo=github)](https://github.com/type-abhay/larry/releases)

A command-line utility built in Go for managing audio metadata. **Lyric Embedder** allows you to search, view, and embed both synced (LRC) and plain text lyrics directly into your music files. It supports **MP3**, **FLAC**, and **M4A** formats using native bit-stream manipulation.

---

## 🚀 How to Use
- Download the .exe from the [releases](https://github.com/type-abhay/larry/releases), **PASTE IT IN YOUR MUSIC FOLDER**.
- It is **IMPORTANT** that the .exe is in your music folder(where you have the files you want to embed lyrics in).
- When you launch the program, it scans your current directory and opens an interactive command loop. Simply type a command and press **Enter**.

### Usage of Commands
| Command | Description | Example |
| :--- | :--- | :--- |
| `search <term>` | Finds all files in the current context containing the keyword. | `search linkin` |
| `info <file>` | Displays full metadata (Title, Artist, Format, Duration, etc.). | `info "in the end"` |
| `get <file>` | Fetches lyrics from the API and lets you choose what to embed. | `get "numb"` |
| `bulk <folder>` | Automatically embeds lyrics for every file in a folder. | `bulk "C:/MyMusic"` |
| `bulk current` | Runs the automated bulk process on your starting directory. | `bulk current` |
| `help` | Displays the command reference guide. | `help` |
| `exit` | Safely closes the application. | `exit` |

### Calling the API for embedding a singular file
1. Type `get <song name>`.
2. The app reads your local tags and queries the LRCLIB API.
3. If a match is found, you choose between **Synced** (time-stamped) or **Plain** lyrics.
4. The app performs "audio surgery" to save the tags without re-encoding the audio.

### The Bulk Flow
The bulk command is designed for speed. It prioritizes **Synced** lyrics; if those aren't found, it falls back to **Plain** text. If the API has no record of the song, it skips the file and provides a summary report at the end.

---

## 🛠️ Building from Source

If you want to compile the project yourself, follow these steps:

### Prerequisites
* **Go 1.20+** installed.
* **rsrc** tool for Windows resource embedding:
  ```powershell
  go install github.com/akavel/rsrc@latest
  ```

### Build Steps
1. **Generate the Windows Resource File:**
   Ensure `icon.ico` and `main.manifest` are in the root directory, then run:
   ```powershell
   rsrc -manifest main.manifest -ico icon.ico -o rsrc.syso
   ```
2. **Download Dependencies:**
   ```powershell
   go mod tidy
   ```
3. **Compile the Executable:**
   ```powershell
   go build -o lyric-tool.exe
   ```

---

## Caution

* **Corrupt Files:** If the program encounters an extremely corrupted MP3 header, it may throw a "Critical Error." Please ensure your files play correctly in a standard media player.
* **API Limits:** When using `bulk` mode on more than 100 files, you may encounter temporary rate limitingThis `README.md` is designed to look professional while clearly explaining the transition from a standard "wizard" flow to your new, powerful command-based interface.
***
## 🆘 Caution
* **API Limits:** When using `bulk` mode on more than 100 files, you may encounter temporary rate limiting from the API provider. 
* **Issues:** Found a bug? Open an issue on the GitHub repository or contact the maintainer.
##  Support the Project
Just give it a ⭐ or share if you found it useful :)
##  Credits
Uses [LRCLIB](https://lrclib.net/) API to fetch lyrics. These guys the goats for providing lyrics in the first place :)

---
*Created with ❤️ & Personal Need 👀*