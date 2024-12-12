package main

import (
	"log"
	"path/filepath"
	"github.com/ginjaninja78/holidaybingo/pkg/cardgen"
)

func main() {
	// Create generator with template
	generator := cardgen.NewGenerator("pkg/cardgen/templates/card_template_working.html")

	// Get list of image files
	images, err := filepath.Glob("img/*.jpg")
	if err != nil {
		log.Fatalf("Failed to get images: %v", err)
	}

	// Set images for the generator
	generator.SetImages(images)

	// Generate single test card
	cards, err := generator.GenerateCards(1)
	if err != nil {
		log.Fatalf("Failed to generate cards: %v", err)
	}

	// Save to PDF
	if err := generator.SaveToPDF(cards, "cards"); err != nil {
		log.Fatalf("Failed to save PDFs: %v", err)
	}

	log.Println("Test card generated successfully")
}
