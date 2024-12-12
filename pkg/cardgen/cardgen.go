/*
TODO: Phase 2 - Logo Integration
1. Fix logo image handling - DONE
2. Verify logo format and dimensions
3. Add proper positioning and scaling
4. Consider alternative image libraries if needed
*/

package cardgen

import (
	"fmt"
	"math/rand"
	"time"
	"os"
	"path/filepath"
	"strings"
	"image/png"
	"image/jpeg"
	"image"
	"image/color"
	_ "image/jpeg"
	"github.com/jung-kurt/gofpdf"
	"github.com/google/uuid"
)

// convertJPGToPNG converts a JPG image to PNG format
func convertJPGToPNG(jpgPath string) (string, error) {
	// Open JPG file
	jpgFile, err := os.Open(jpgPath)
	if err != nil {
		return "", fmt.Errorf("failed to open JPG %s: %v", jpgPath, err)
	}
	defer jpgFile.Close()

	// Decode JPG
	img, err := jpeg.Decode(jpgFile)
	if err != nil {
		return "", fmt.Errorf("failed to decode JPG %s: %v", jpgPath, err)
	}

	// Convert to RGBA (8-bit per channel)
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, color.RGBAModel.Convert(img.At(x, y)))
		}
	}

	// Create PNG file in temp directory
	pngPath := filepath.Join(os.TempDir(), strings.TrimSuffix(filepath.Base(jpgPath), ".jpg")+".png")
	pngFile, err := os.Create(pngPath)
	if err != nil {
		return "", fmt.Errorf("failed to create PNG %s: %v", pngPath, err)
	}
	defer pngFile.Close()

	// Encode as PNG with default encoder (8-bit depth)
	encoder := png.Encoder{
		CompressionLevel: png.DefaultCompression,
	}
	if err := encoder.Encode(pngFile, rgba); err != nil {
		return "", fmt.Errorf("failed to encode PNG %s: %v", pngPath, err)
	}

	return pngPath, nil
}

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

// generateID creates a unique card ID
func generateID() string {
	id := uuid.New().String()
	shortID := strings.ReplaceAll(id, "-", "")[:6]
	return fmt.Sprintf("Card No.# %s", shortID)
}

// GenerateCards generates the specified number of unique bingo cards
func (g *Generator) GenerateCards(count int) ([]Card, error) {
	if len(g.images) < 24 {
		return nil, fmt.Errorf("not enough images provided: need at least 24, got %d", len(g.images))
	}

	cards := make([]Card, count)
	for i := 0; i < count; i++ {
		cardID := generateID()

		shuffled := make([]string, len(g.images))
		copy(shuffled, g.images)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		squares := make([]string, 25)
		copy(squares[:12], shuffled[:12])
		squares[12] = filepath.Join("assets", "free_space.jpg")
		copy(squares[13:], shuffled[12:24])

		cards[i] = Card{
			ID:      cardID,
			Squares: squares,
		}
	}

	return cards, nil
}

// SaveToPDF saves the cards to PDF files optimized for Acrobat stamp tool
func (g *Generator) SaveToPDF(cards []Card, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Convert free space to PNG
	freeSpacePath := filepath.Join("assets", "free_space.jpg")
	freeSpacePNG, err := convertJPGToPNG(freeSpacePath)
	if err != nil {
		return fmt.Errorf("failed to convert free space image: %v", err)
	}
	defer os.Remove(freeSpacePNG)

	// Convert logo to PNG
	logoPath := filepath.Join("assets", "logo.jpg")
	logoPNG, err := convertJPGToPNG(logoPath)
	if err != nil {
		return fmt.Errorf("failed to convert logo image: %v", err)
	}
	defer os.Remove(logoPNG)

	// Convert gameplay images to PNG
	imagePNGs := make([]string, len(g.images))
	for i, img := range g.images {
		pngPath, err := convertJPGToPNG(img)
		if err != nil {
			return fmt.Errorf("failed to convert gameplay image %s: %v", img, err)
		}
		imagePNGs[i] = pngPath
		defer os.Remove(pngPath)
	}

	for _, card := range cards {
		// Create PDF
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.SetFont("Arial", "", 12)
		pdf.AddPage()

		// Set up dimensions
		const (
			margin     = 20.0
			cellSize   = 35.0
			gridSize   = 5
			pageWidth  = 210.0
			pageHeight = 297.0
			logoWidth  = 50.0
			logoHeight = 20.0
			idPadding  = 5.0 // 5mm padding for card ID
		)

		// Add card ID at top left
		pdf.SetFont("Arial", "", 12)
		pdf.Text(idPadding, idPadding+5, card.ID) // +5 to account for font height

		// Add logo at the top
		logoX := (pageWidth - logoWidth) / 2
		logoY := margin + 5
		pdf.Image(logoPNG, logoX, logoY, logoWidth, logoHeight, false, "PNG", 0, "")

		// Add title
		pdf.SetFont("Arial", "B", 24)
		title := "SSO&O Holiday Bingo"
		titleWidth := pdf.GetStringWidth(title)
		pdf.Text((pageWidth-titleWidth)/2, margin+30, title)

		// Calculate grid position
		startX := (pageWidth - (cellSize * float64(gridSize))) / 2
		startY := margin + 45

		// Draw grid
		for row := 0; row < gridSize; row++ {
			for col := 0; col < gridSize; col++ {
				x := startX + float64(col)*cellSize
				y := startY + float64(row)*cellSize
				index := row*gridSize + col
				
				// Draw cell border
				pdf.SetFillColor(255, 255, 255) // White background
				pdf.Rect(x, y, cellSize, cellSize, "DF")

				if index == 12 { // Center cell
					pdf.Image(freeSpacePNG, x, y, cellSize, cellSize, false, "PNG", 0, "")
				} else {
					imgIndex := index
					if index > 12 {
						imgIndex--
					}
					pdf.Image(imagePNGs[imgIndex], x, y, cellSize, cellSize, false, "PNG", 0, "")
				}
				
				// Re-draw border to ensure it's visible
				pdf.Rect(x, y, cellSize, cellSize, "D")
			}
		}

		// Calculate position for instructions (below grid with padding)
		instructionsY := startY + (cellSize * float64(gridSize)) + 20 // 20mm padding

		// Add instructions header
		pdf.SetFont("Arial", "B", 12)
		header := "Instructions:"
		headerWidth := pdf.GetStringWidth(header)
		pdf.Text((pageWidth-headerWidth)/2, instructionsY, header)

		// Add step-by-step instructions
		pdf.SetFont("Arial", "", 10)
		instructions := []string{
			"To Mark: Comments > Drawing Markups > Stamps > Select stamp > Click square",
			"To Reset: Tools > Comments > Clear All",
		}

		lineHeight := 5.0 // Space between lines
		for i, line := range instructions {
			lineWidth := pdf.GetStringWidth(line)
			y := instructionsY + float64(i+1)*lineHeight + 5 // +5 for padding below header
			pdf.Text((pageWidth-lineWidth)/2, y, line)
		}

		// Save PDF
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s.pdf", strings.ReplaceAll(card.ID, " ", "_")))
		if err := pdf.OutputFileAndClose(outputPath); err != nil {
			return fmt.Errorf("failed to save PDF: %v", err)
		}
	}

	return nil
}
