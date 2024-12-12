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
	imageCache      *cache.ImageCacheManager
	totalImages     int
	loadedCount     int
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize cache manager
	imageCache = cache.NewImageCacheManager(4)

	myApp := app.NewWithID("com.example.holidaybingo")
	myApp.Settings().SetTheme(theme.LightTheme())
	myWindow := myApp.NewWindow("SSO&O Holiday BINGO!")

	initializeGame()

	mainLabel = widget.NewLabel("")
	mainLabel.Alignment = fyne.TextAlignCenter
	mainLabel.TextStyle = fyne.TextStyle{Bold: true}

	nextButton = widget.NewButton("Next", func() {
		if nextButton.Text == "Continue" {
			gameActive = true
			nextButton.SetText("Next")
			bingoButton.SetText("Bingo!")
			log.Println("Game continued")
			return
		}
		
		if !gameActive {
			log.Println("Game not active, ignoring Next click")
			return
		}
		displayNextImage()
		log.Println("Next clicked")
	})

	bingoButton = widget.NewButton("Bingo!", func() {
		if bingoButton.Text == "End Game" {
			log.Println("Ending game and resetting state")
			gameActive = false
			currentIndex = 0
			loadedCount = 0
			totalImages = 0
			images = make([]fyne.Resource, 0)
			historyShelf.Objects = []fyne.CanvasObject{}
			historyScroll.Refresh()
			imageContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Click New Game to start!")}
			bingoButton.SetText("Bingo!")
			nextButton.SetText("Next")
			mainView.Refresh()
			log.Println("Game ended and reset")
			return
		}
		
		gameActive = false
		nextButton.SetText("Continue")
		bingoButton.SetText("End Game")
		log.Println("Bingo clicked")
	})

	// Create New Game button separately to reference it
	newGameButton = widget.NewButton("New Game", func() {
		log.Println("Starting new game")
		startNewGame()
		mainLabel.SetText("Let's Play!")
	})

	sidebar := container.NewVBox(
		widget.NewLabel("SSO&O"),
		newGameButton,  // Use the stored button
		widget.NewButton("Generate Cards", func() {
			log.Println("Generate Cards clicked")
		}),
		widget.NewButton("Verify Bingo", func() {
			log.Println("Verify Bingo clicked")
		}),
		widget.NewButton("Scoreboard", func() {
			log.Println("Scoreboard clicked")
		}),
		widget.NewButton("Next Round", func() {
			log.Println("Next Round clicked")
		}),
		widget.NewButton("Config", func() {
			log.Println("Config clicked")
		}),
		widget.NewButton("Exit", func() {
			myApp.Quit()
		}),
		layout.NewSpacer(),
	)

	historyLabel := widget.NewLabel("History")
	historyLabel.Alignment = fyne.TextAlignCenter

	historyShelf = container.NewHBox()
	historyScroll = container.NewHScroll(historyShelf)

	historyContainer := container.NewVBox(
		historyLabel,
		container.NewPadded(historyScroll),
	)

	historyContainer.Resize(fyne.NewSize(800, 150))

	imageContainer = container.NewCenter(widget.NewLabel("Click New Game to start!"))

	buttonBox := container.NewHBox(
		layout.NewSpacer(),
		nextButton,
		layout.NewSpacer(),
		bingoButton,
		layout.NewSpacer(),
	)

	rightSide := container.NewBorder(
		historyContainer,
		buttonBox,
		nil,
		nil,
		imageContainer,
	)

	content := container.NewHSplit(
		sidebar,
		rightSide,
	)
	content.SetOffset(0.2)

	mainView = container.NewMax(content)
	myWindow.SetContent(mainView)
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.ShowAndRun()
}

func initializeGame() {
	log.Println("Initializing game state")
	images = make([]fyne.Resource, 0)
	currentIndex = 0
	loadedCount = 0
	totalImages = 0
	gameActive = false
	// Create a new image cache manager
	imageCache = cache.NewImageCacheManager(4)
	log.Println("Game initialized")
}

func startNewGame() {
	log.Println("Starting new game setup")
	initializeGame()
	
	// Clear history
	if historyShelf != nil {
		historyShelf.Objects = []fyne.CanvasObject{}
		historyScroll.Refresh()
	}

	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
		return
	}
	exeDir := filepath.Dir(exePath)
	imgDir := filepath.Join(exeDir, "img")
	
	log.Printf("Looking for images in: %s", imgDir)
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Printf("Failed to read image directory: %v", err)
		// Try current directory as fallback
		imgDir = "img"
		entries, err = os.ReadDir(imgDir)
		if err != nil {
			log.Printf("Failed to read fallback image directory: %v", err)
			return
		}
	}

	var imagePaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := filepath.Ext(entry.Name())
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				fullPath := filepath.Join(imgDir, entry.Name())
				imagePaths = append(imagePaths, fullPath)
				log.Printf("Found image: %s", fullPath)
			}
		}
	}

	if len(imagePaths) == 0 {
		log.Println("No images found in directory")
		imageContainer.Objects = []fyne.CanvasObject{
			widget.NewLabel("No images found in directory"),
		}
		return
	}

	log.Printf("Found %d images to load", len(imagePaths))
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
			if result.Err != nil {
				log.Printf("Error loading image %s: %v", result.Path, result.Err)
				continue
			}
			
			if !loadedImages[result.Path] {
				log.Printf("Successfully loaded image: %s", result.Path)
				images = append(images, result.Resource)
				loadedImages[result.Path] = true
				loadedCount++

				if loadedCount == totalImages {
					log.Printf("All %d images loaded, shuffling", totalImages)
					source := rand.NewSource(time.Now().UnixNano())
					r := rand.New(source)

					r.Shuffle(len(images), func(i, j int) {
						images[i], images[j] = images[j], images[i]
					})

					if !gameActive {
						log.Println("Starting game with loaded images")
						currentIndex = 0
						gameActive = true
						displayNextImage()
					}
				}
			}
		}
	}()

	// Start with any cached images immediately
	for _, path := range imagePaths {
		if resource, exists := imageCache.Get(path); exists && !loadedImages[path] {
			log.Printf("Found cached image: %s", path)
			images = append(images, resource)
			loadedImages[path] = true
			loadedCount++
		}
	}
}

func displayNextImage() {
	if !gameActive {
		log.Println("Game not active, cannot display next image")
		return
	}

	if currentIndex >= len(images) {
		if loadedCount < totalImages {
			log.Printf("Waiting for more images to load (%d/%d loaded)", loadedCount, totalImages)
			imageContainer.Objects = []fyne.CanvasObject{
				widget.NewLabel("Loading more images..."),
			}
			return
		}
		log.Println("Game Over - All images shown")
		imageContainer.Objects = []fyne.CanvasObject{
			widget.NewLabel("Game Over - All images shown"),
		}
		gameActive = false
		return
	}

	currentImg := images[currentIndex]
	log.Printf("Displaying image %d/%d: %s", currentIndex+1, len(images), currentImg.Name())

	img := canvas.NewImageFromResource(currentImg)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(400, 400))
	imageContainer.Objects = []fyne.CanvasObject{img}

	if currentIndex > 0 {
		historyImg := canvas.NewImageFromResource(images[currentIndex-1])
		historyImg.SetMinSize(fyne.NewSize(100, 100))
		historyImg.FillMode = canvas.ImageFillContain
		historyShelf.Add(historyImg)
		historyScroll.Refresh()
		log.Printf("Added image %d to history", currentIndex)
	}

	currentIndex++
	mainView.Refresh()
}
