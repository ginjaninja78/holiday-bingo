package cardgen

import (
	"fmt"
	"math/rand"
	"time"
	"os"
	"path/filepath"
	"github.com/jung-kurt/gofpdf"
	"image"
	_ "image/jpeg"
	_ "image/png"
)

// Card represents a bingo card with its properties
type Card struct {
	ID      string
	Squares []string
}

// Generator handles the card generation process
type Generator struct {
	templatePath string
	images       []string
}

// NewGenerator creates a new card generator
func NewGenerator(templatePath string) *Generator {
	return &Generator{
		templatePath: templatePath,
		images:      make([]string, 0),
	}
}

// SetImages sets the available images for card generation
func (g *Generator) SetImages(images []string) {
	g.images = images
}

// generateID creates a unique card ID in format XX123
func generateID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const numbers = "0123456789"

	rand.Seed(time.Now().UnixNano())

	id := make([]byte, 5)
	id[0] = letters[rand.Intn(len(letters))]
	id[1] = letters[rand.Intn(len(letters))]
	id[2] = numbers[rand.Intn(len(numbers))]
	id[3] = numbers[rand.Intn(len(numbers))]
	id[4] = numbers[rand.Intn(len(numbers))]

	return string(id)
}

// GenerateCards generates the specified number of unique bingo cards
func (g *Generator) GenerateCards(count int) ([]Card, error) {
	if len(g.images) < 24 { // Need at least 24 images (5x5 grid minus center)
		return nil, fmt.Errorf("not enough images provided: need at least 24, got %d", len(g.images))
	}

	cards := make([]Card, count)
	for i := 0; i < count; i++ {
		// Generate unique ID
		cardID := generateID()

		// Shuffle images for this card
		shuffled := make([]string, len(g.images))
		copy(shuffled, g.images)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		// Take first 24 images for the card
		squares := make([]string, 25)
		copy(squares[:12], shuffled[:12])
		squares[12] = "FREE" // Center square is FREE
		copy(squares[13:], shuffled[12:24])

		cards[i] = Card{
			ID:      cardID,
			Squares: squares,
		}
	}

	return cards, nil
}

// SaveToPDF saves the cards to PDF files with clickable squares
func (g *Generator) SaveToPDF(cards []Card, outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	for _, card := range cards {
		// Create new PDF
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetAutoPageBreak(false, 0)
		pdf.AddPage()

		// Set up card layout
		const (
			margin     = 10.0  // mm
			cellSize   = 35.0  // mm
			gridSize   = 5
			imageSize  = 30.0  // mm
			pageWidth  = 210.0 // A4 width in mm
			pageHeight = 297.0 // A4 height in mm
		)

		// Calculate starting position to center the grid
		startX := (pageWidth - (cellSize * float64(gridSize))) / 2
		startY := margin + 20 // Leave space for title

		// Add title
		pdf.SetFont("Arial", "B", 24)
		pdf.Text((pageWidth-pdf.GetStringWidth("Holiday Bingo"))/2, margin+10, "Holiday Bingo")

		// Add card ID
		pdf.SetFont("Arial", "", 12)
		pdf.Text(margin, pageHeight-margin, fmt.Sprintf("Card ID: %s", card.ID))

		// Draw grid and add images
		for row := 0; row < gridSize; row++ {
			for col := 0; col < gridSize; col++ {
				x := startX + float64(col)*cellSize
				y := startY + float64(row)*cellSize
				index := row*gridSize + col

				// Draw cell border
				pdf.Rect(x, y, cellSize, cellSize, "D")

				if index == 12 { // Center FREE space
					pdf.SetFont("Arial", "B", 16)
					pdf.Text(x+(cellSize-pdf.GetStringWidth("FREE"))/2, y+cellSize/2, "FREE")
					
					// Add checkbox
					pdf.SetFont("ZapfDingbats", "", 12)
					pdf.Text(x+2, y+5, "☐")
				} else {
					// Add image
					imgPath := card.Squares[index]
					if imgFile, err := os.Open(imgPath); err == nil {
						if img, _, err := image.DecodeConfig(imgFile); err == nil {
							imgFile.Close()
							
							// Calculate scaling to fit in cell while maintaining aspect ratio
							scale := imageSize / float64(img.Width)
							if float64(img.Height)*scale > imageSize {
								scale = imageSize / float64(img.Height)
							}
							
							imgWidth := float64(img.Width) * scale
							imgHeight := float64(img.Height) * scale
							
							// Center image in cell
							imgX := x + (cellSize-imgWidth)/2
							imgY := y + (cellSize-imgHeight)/2
							
							pdf.Image(imgPath, imgX, imgY, imgWidth, imgHeight, false, "", 0, "")
						}
					}
					
					// Add checkbox
					pdf.SetFont("ZapfDingbats", "", 12)
					pdf.Text(x+2, y+5, "☐")
				}
			}
		}

		// Add instructions
		pdf.SetFont("Arial", "", 10)
		pdf.Text(margin, margin, "Click or mark the boxes to shade squares as they are called.")

		// Save PDF
		outputPath := filepath.Join(outputDir, fmt.Sprintf("HolidayBingo_%s.pdf", card.ID))
		if err := pdf.OutputFileAndClose(outputPath); err != nil {
			return fmt.Errorf("failed to save PDF: %v", err)
		}
	}

	return nil
}
