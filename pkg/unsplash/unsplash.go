package unsplash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var requestCount int = 0
const maxRequests int = 10

// CanMakeRequest checks if the request limit has been reached
func CanMakeRequest() bool {
    return requestCount < maxRequests
}

// IncrementRequestCount increments the request counter
func IncrementRequestCount() {
    requestCount++
}

// Struct to parse the JSON response from Unsplash
type UnsplashResponse struct {
	Urls struct {
		Full string `json:"full"`
	} `json:"urls"`
}

// GetPhoto fetches a photo from Unsplash
func GetPhoto() ([]byte, error) {
	if !CanMakeRequest() {
		return nil, fmt.Errorf("Request limit reached")
	}

	apiKey := os.Getenv("UNSPLASH_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("Unsplash API key not set")
	}

	// TODO: Refine query keywords for better matching of winter holiday themes and clip art
	// Further refined query to focus on winter holiday themes and clip art
	query := url.QueryEscape("Christmas tree clip art,Hanukkah menorah illustration,Kwanzaa candles art,holiday decorations graphic")
	url := fmt.Sprintf("https://api.unsplash.com/photos/random?query=%s&client_id=%s&orientation=squarish", query, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Log HTTP status and headers
	log.Printf("HTTP Status: %s", resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			log.Printf("Header: %s: %s", key, value)
		}
	}

	// Read and log the response body for debugging
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("Response Body: %s", respBody)

	// Parse the JSON response
	var unsplashResp UnsplashResponse
	if err := json.Unmarshal(respBody, &unsplashResp); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %v", err)
	}

	// Fetch the actual image
	imageResp, err := http.Get(unsplashResp.Urls.Full)
	if err != nil {
		return nil, err
	}
	defer imageResp.Body.Close()

	// Read the image data
	imageData, err := ioutil.ReadAll(imageResp.Body)
	if err != nil {
		return nil, err
	}

	IncrementRequestCount()
	return imageData, nil
}

// SavePhoto saves the photo to the img directory with a sequential name
func SavePhoto(photoData []byte) error {
	imgDir := "img"
	// Ensure the img directory exists
	if _, err := os.Stat(imgDir); os.IsNotExist(err) {
		os.Mkdir(imgDir, os.ModePerm)
	}

	// Find the next available image filename
	var imgPath string
	for i := 1; ; i++ {
		imgPath = fmt.Sprintf("%s/img%d.jpg", imgDir, i)
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			break
		}
	}

	// Save the image
	err := os.WriteFile(imgPath, photoData, 0644)
	if err != nil {
		return err
	}
	return nil
}
