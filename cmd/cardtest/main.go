package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"holidaybingo/pkg/cardgen"
)

func main() {
	// Initialize card generator with template
	templatePath := filepath.Join("pkg", "cardgen", "templates", "card_template.html.html")
	generator := cardgen.NewGenerator(templatePath)

	// Get list of actual images from img directory
	imgDir := filepath.Join("img")
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Fatalf("Failed to read image directory: %v", err)
	}

	var images []string
	for _, entry := range entries {
		if !entry.IsDir() {
			// Only include image files
			if ext := filepath.Ext(entry.Name()); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				images = append(images, filepath.Join(imgDir, entry.Name()))
			}
		}
	}

	if len(images) < 24 {
		log.Fatalf("Not enough images in img directory. Need at least 24, found %d", len(images))
	}

	generator.SetImages(images)

	// Generate 3 test cards
	cards, err := generator.GenerateCards(3)
	if err != nil {
		log.Fatalf("Failed to generate cards: %v", err)
	}

	// Save cards to PDF files in the cards directory
	if err := generator.SaveToPDF(cards, "cards"); err != nil {
		log.Fatalf("Failed to save cards: %v", err)
	}

	fmt.Printf("Successfully generated %d cards and saved to cards directory\n", len(cards))
	fmt.Println("\nCard IDs:")
	for i, card := range cards {
		fmt.Printf("Card %d: %s\n", i+1, card.ID)
	}
}
