package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"lyric-emb/audio"
)

// 1. Define the Response Structure
type LyricResponse struct {
	ID           int    `json:"id"`
	TrackName    string `json:"trackName"`
	ArtistName   string `json:"artistName"`
	AlbumName    string `json:"albumName"`
	Duration     int    `json:"duration"`
	Instrumental bool   `json:"instrumental"`
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

// 2. Define the Error Structure
type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const baseURL = "https://lrclib.net/api/get"

// GetLyrics fetches lyrics from the API using metadata
func GetLyrics(meta audio.Metadata) (*LyricResponse, error) {
	// 3. Build the Query Parameters
	params := url.Values{}
	params.Add("artist_name", meta.Artist)
	params.Add("track_name", meta.Title)
	params.Add("album_name", meta.Album)
	params.Add("duration", strconv.Itoa(meta.Duration))

	// 4. Construct the Final URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 5. Initialize the HTTP Client with a Timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 6. Make the GET Request
	resp, err := client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	// 7. Handle Non-200 Status Codes
	if resp.StatusCode != http.StatusOK {
		var apiErr ApiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Message)
	}

	// 8. Decode the Successful JSON Response
	var result LyricResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &result, nil
}
