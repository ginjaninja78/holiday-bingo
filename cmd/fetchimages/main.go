package main

import (
	"fmt"
	"log"
	"time"
	"holidaybingo/pkg/unsplash"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// We need at least 24 images for the bingo cards
	requiredImages := 24
	fetchedImages := 0

	fmt.Printf("Fetching %d images from Unsplash...\n", requiredImages)

	for fetchedImages < requiredImages {
		if !unsplash.CanMakeRequest() {
			fmt.Println("Reached API request limit. Waiting for 1 hour...")
			time.Sleep(time.Hour)
			continue
		}

		// Fetch a photo from Unsplash
		photoData, err := unsplash.GetPhoto()
		if err != nil {
			fmt.Printf("Error fetching photo: %v\n", err)
			continue
		}

		// Save the photo
		if err := unsplash.SavePhoto(photoData); err != nil {
			fmt.Printf("Error saving photo: %v\n", err)
			continue
		}

		fetchedImages++
		fmt.Printf("Fetched and saved image %d of %d\n", fetchedImages, requiredImages)

		// Sleep for a bit to avoid hitting rate limits
		time.Sleep(time.Second)
	}

	fmt.Println("Successfully fetched all required images!")
}
