package main

import (
	"fmt"
	"log"
	"holidaybingo/pkg/unsplash"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Fetch a photo from Unsplash
	photoData, err := unsplash.GetPhoto()
	if err != nil {
		log.Fatalf("Error fetching photo: %v", err)
	}

	// Save the photo
	err = unsplash.SavePhoto(photoData)
	if err != nil {
		log.Fatalf("Error saving photo: %v", err)
	}

	fmt.Println("Image successfully fetched and saved!")
}
