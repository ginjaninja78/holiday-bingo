package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
)

var (
	mainLabel       *widget.Label
	historyShelf    *fyne.Container
	images          []fyne.Resource
	currentIndex    int
	gameActive      bool
	newGameButton   *widget.Button
	nextButton      *widget.Button
	bingoButton     *widget.Button
	continueButton  *widget.Button
	endGameButton   *widget.Button
	mainView        *fyne.Container
	buttonContainer *fyne.Container
	historyScroll   *container.Scroll
	imageContainer  *fyne.Container
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	myApp := app.NewWithID("com.example.holidaybingo")
	myApp.Settings().SetTheme(theme.LightTheme())
	myWindow := myApp.NewWindow("SSO&O Holiday BINGO!")

	// Initialize game state
	initializeGame()

	// Main display
	mainLabel = widget.NewLabel("") // Initialize mainLabel first
	mainLabel.Alignment = fyne.TextAlignCenter
	mainLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create buttons
	nextButton = widget.NewButton("Next", func() {
		if nextButton.Text == "Continue" {
			// Resume the game
			gameActive = true
			nextButton.SetText("Next")
			bingoButton.SetText("Bingo!")  // Reset Bingo button text too
			log.Println("Game continued")
			return
		}
		
		if !gameActive {
			return
		}
		displayNextImage()
		log.Println("Next clicked")
	})

	bingoButton = widget.NewButton("Bingo!", func() {
		if bingoButton.Text == "End Game" {
			// End the game
			gameActive = false
			currentIndex = 0
			historyShelf.Objects = []fyne.CanvasObject{}
			historyScroll.Refresh()
			imageContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Click New Game to start!")}
			bingoButton.SetText("Bingo!")
			nextButton.SetText("Next")
			mainView.Refresh()
			return
		}
		
		// Pause the game for Bingo verification
		gameActive = false
		nextButton.SetText("Continue")
		bingoButton.SetText("End Game")
		log.Println("Bingo clicked")
	})

	// Left Sidebar
	sidebar := container.NewVBox(
		widget.NewLabel("SSO&O"),
		widget.NewButton("New Game", func() {
			startNewGame()
			mainLabel.SetText("Let's Play!")
		}),
		widget.NewButton("Generate Cards", func() {
			// TODO: Implement Generate Cards functionality
			log.Println("Generate Cards clicked")
		}),
		widget.NewButton("Verify Bingo", func() {
			// TODO: Implement Verify Bingo functionality
			log.Println("Verify Bingo clicked")
		}),
		widget.NewButton("Scoreboard", func() {
			// TODO: Implement Scoreboard functionality
			log.Println("Scoreboard clicked")
		}),
		widget.NewButton("Next Round", func() {
			// TODO: Implement Next Round functionality
			log.Println("Next Round clicked")
		}),
		widget.NewButton("Config", func() {
			// TODO: Implement Config functionality
			log.Println("Config clicked")
		}),
		widget.NewButton("Exit", func() {
			myApp.Quit()
		}),
		layout.NewSpacer(),
	)

	// History section at top of right side
	historyLabel := widget.NewLabel("History")
	historyLabel.Alignment = fyne.TextAlignCenter

	historyShelf = container.NewHBox()
	historyScroll = container.NewHScroll(historyShelf)

	historyContainer := container.NewVBox(
		historyLabel,
		container.NewPadded(historyScroll),
	)

	// Set a fixed height for the history section
	historyContainer.Resize(fyne.NewSize(800, 150))

	// Main display area
	imageContainer = container.NewCenter(widget.NewLabel("Click New Game to start!"))

	// Buttons below main display
	buttonBox := container.NewHBox(
		layout.NewSpacer(),
		nextButton,
		layout.NewSpacer(),
		bingoButton,
		layout.NewSpacer(),
	)

	// Right side layout (history at top, main content below)
	rightSide := container.NewBorder(
		historyContainer, // top
		buttonBox,        // bottom
		nil,              // left
		nil,              // right
		imageContainer,   // center
	)

	// Combine left sidebar with right side content
	content := container.NewHSplit(
		sidebar,
		rightSide,
	)
	content.SetOffset(0.2) // Sidebar takes 20% of horizontal space

	mainView = container.NewMax(content)
	myWindow.SetContent(mainView)
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.ShowAndRun()
}

func initializeGame() {
	// Initialize resources and game state
	images = make([]fyne.Resource, 0)
	currentIndex = 0
	gameActive = false
	log.Println("Game initialized")
}

func startNewGame() {
	// Clear history
	historyShelf.Objects = []fyne.CanvasObject{}
	historyScroll.Refresh()

	// Load images from the img directory
	imgDir := "img"
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Printf("Failed to read image directory: %v", err)
		return
	}

	// Reset images slice
	images = make([]fyne.Resource, 0)

	// Load each image file
	for _, entry := range entries {
		if !entry.IsDir() {
			// Only include image files
			if ext := filepath.Ext(entry.Name()); ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				imgPath := filepath.Join(imgDir, entry.Name())
				
				// Optimize and load the image
				imgData, err := optimizeImage(imgPath)
				if err != nil {
					log.Printf("Failed to optimize image %s: %v", imgPath, err)
					continue
				}

				// Create a static resource from the optimized image data
				imgRes := fyne.NewStaticResource(entry.Name(), imgData)
				images = append(images, imgRes)
			}
		}
	}

	if len(images) == 0 {
		log.Println("No images found in img directory")
		return
	}

	// Fisher-Yates shuffle
	rand.Seed(time.Now().UnixNano())
	for i := len(images) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		images[i], images[j] = images[j], images[i]
	}
	
	gameActive = true
	currentIndex = 0
	displayNextImage()
	log.Println("New game started with shuffled images")
}

func optimizeImage(imgPath string) ([]byte, error) {
	// Open the image file
	file, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Calculate new size (max dimension of 800 pixels while maintaining aspect ratio)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	var newWidth, newHeight uint
	if width > height {
		newWidth = 800
		newHeight = uint(float64(height) * (800.0 / float64(width)))
	} else {
		newHeight = 800
		newWidth = uint(float64(width) * (800.0 / float64(height)))
	}

	// Resize the image
	resized := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// Encode the resized image
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, resized)
	default:
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func displayNextImage() {
	if !gameActive {
		log.Println("Game not active")
		return
	}

	// If we've reached the end, cycle back to the beginning
	if currentIndex >= len(images) {
		currentIndex = 0
	}

	// Add current image to history before moving to next (only if not first display)
	if currentIndex > 0 {
		previousImage := canvas.NewImageFromResource(images[currentIndex-1])
		previousImage.SetMinSize(fyne.NewSize(100, 100))
		previousImage.FillMode = canvas.ImageFillContain

		historyFrame := container.NewMax(previousImage)
		historyFrame.Resize(fyne.NewSize(100, 100))

		paddedFrame := container.NewPadded(historyFrame)
		historyShelf.Add(paddedFrame)
		historyScroll.Refresh()
	}

	// Update the main image
	image := canvas.NewImageFromResource(images[currentIndex])
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(500, 500))
	imageContainer.Objects = []fyne.CanvasObject{image}

	currentIndex++
	mainView.Refresh()
	log.Printf("Displayed image %d", currentIndex)
}
