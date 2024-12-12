// Package main implements the Holiday Bingo game with concurrent image processing
// Version 1.0 - Implements basic game functionality with optimized image loading
package main

import (
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
	"github.com/ginjaninja78/holidaybingo/pkg/cache"
)

// Global state variables for game management
// TODO: Refactor into proper state management structure in v2
var (
	mainLabel       *widget.Label
	historyShelf    *fyne.Container
	images          []fyne.Resource    // Stores all loaded game images
	currentIndex    int               // Current image index in play
	gameActive      bool              // Controls game state
	newGameButton   *widget.Button
	nextButton      *widget.Button
	bingoButton     *widget.Button
	continueButton  *widget.Button
	endGameButton   *widget.Button
	mainView        *fyne.Container
	buttonContainer *fyne.Container
	historyScroll   *container.Scroll
	imageContainer  *fyne.Container
	imageCache      *cache.ImageCacheManager  // Manages concurrent image loading and caching
	totalImages     int                      // Total number of images to be loaded
	loadedCount     int                      // Number of images currently loaded
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize cache manager
	imageCache = cache.NewImageCacheManager(4) // Use 4 workers

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

	// Create a loading progress bar
	progress := widget.NewProgressBar()
	imageContainer.Objects = []fyne.CanvasObject{
		widget.NewLabel("Loading images..."),
		progress,
	}
	imageContainer.Refresh()

	// Collect image paths
	var imagePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			imagePaths = append(imagePaths, filepath.Join(imgDir, entry.Name()))
		}
	}

	if len(imagePaths) == 0 {
		imageContainer.Objects = []fyne.CanvasObject{
			widget.NewLabel("No images found in directory"),
		}
		return
	}

	// Initialize image slice with capacity
	images = make([]fyne.Resource, 0, len(imagePaths))
	loadedImages := make(map[string]bool)
	
	// Start background loading
	go imageCache.PreloadImages(imagePaths)

	// Start a goroutine to collect results
	resultChan := imageCache.GetResultChannel()
	totalImages = len(imagePaths)
	loadedCount = 0

	go func() {
		for result := range resultChan {
			if result.Err == nil && !loadedImages[result.Path] {
				images = append(images, result.Resource)
				loadedImages[result.Path] = true
				loadedCount++
				progress.SetValue(float64(loadedCount) / float64(totalImages))

				// If this is the first image, start the game
				if len(images) == 1 {
					rand.Seed(time.Now().UnixNano())
					currentIndex = 0
					gameActive = true
					displayNextImage()
				}
			}
		}
	}()

	// Start with any cached images immediately
	for _, path := range imagePaths {
		if resource, exists := imageCache.Get(path); exists && !loadedImages[path] {
			images = append(images, resource)
			loadedImages[path] = true
			loadedCount++
		}
	}

	// Start game if we have any images
	if len(images) > 0 {
		rand.Seed(time.Now().UnixNano())
		currentIndex = 0
		gameActive = true
		displayNextImage()
	}
}

func displayNextImage() {
	if !gameActive {
		return
	}

	// If we've shown all current images but more are still loading, wait for next image
	if currentIndex >= len(images) {
		if loadedCount < totalImages {
			imageContainer.Objects = []fyne.CanvasObject{
				widget.NewLabel("Loading more images..."),
			}
			return
		}
		// Only end if we've shown ALL images
		imageContainer.Objects = []fyne.CanvasObject{
			widget.NewLabel("Game Over - All images shown"),
		}
		gameActive = false
		return
	}

	// Display current image
	img := canvas.NewImageFromResource(images[currentIndex])
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(400, 400))
	imageContainer.Objects = []fyne.CanvasObject{img}

	// Add to history
	historyImg := canvas.NewImageFromResource(images[currentIndex])
	historyImg.SetMinSize(fyne.NewSize(100, 100))
	historyImg.FillMode = canvas.ImageFillContain
	historyShelf.Add(historyImg)
	historyScroll.Refresh()

	currentIndex++
	mainView.Refresh()
}
