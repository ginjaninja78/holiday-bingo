package ui

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ginjaninja78/holidaybingo/pkg/cache"
	"log"
)

// UIManager handles all UI components and their interactions
type UIManager struct {
	window        fyne.Window
	mainImage     *canvas.Image
	historyShelf  *fyne.Container
	historyScroll *container.Scroll
	imageCache    *cache.ImageCacheManager
	imageList     []string
	currentIdx    int
	currentImage  fyne.Resource
	usedIndices   map[int]bool
	rng           *rand.Rand
}

// NewUIManager creates a new UI manager
func NewUIManager(window fyne.Window) *UIManager {
	manager := &UIManager{
		window:        window,
		mainImage:     canvas.NewImageFromFile(""),
		historyShelf:  container.NewHBox(),
		historyScroll: container.NewHScroll(nil), // Will set shelf later
		imageCache:    cache.NewImageCacheManager(cache.DefaultMaxWorkers),
		usedIndices:   make(map[int]bool),
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	manager.historyScroll.Content = manager.historyShelf
	manager.initialize()
	return manager
}

// initialize sets up all UI components
func (m *UIManager) initialize() {
	m.mainImage.FillMode = canvas.ImageFillOriginal
	m.mainImage.SetMinSize(fyne.NewSize(400, 400))

	// Create sidebar
	sidebar := container.NewVBox(
		widget.NewLabel("SSO&O"),
		widget.NewButton("New Game", m.startNewGame),
		widget.NewButton("Generate Cards", nil),
		widget.NewButton("Verify Bingo", nil),
		widget.NewButton("Scoreboard", nil),
		widget.NewButton("Next Round", nil),
		widget.NewButton("Config", nil),
		widget.NewButton("Exit", m.window.Close),
		layout.NewSpacer(),
	)

	// Create history container
	historyContainer := container.NewVBox(
		widget.NewLabel("History"),
		container.NewPadded(m.historyScroll),
	)
	historyContainer.Resize(fyne.NewSize(800, 150))

	// Create main content area
	mainContent := container.NewMax(m.mainImage)

	// Create button box
	buttonBox := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButton("Next", m.displayNextImage),
		layout.NewSpacer(),
		widget.NewButton("Bingo!", nil),
		layout.NewSpacer(),
	)

	// Create right side content
	rightSide := container.NewBorder(
		historyContainer,
		buttonBox,
		nil,
		nil,
		mainContent,
	)

	// Create main split
	content := container.NewHSplit(sidebar, rightSide)
	content.SetOffset(0.2)

	// Set window content
	m.window.SetContent(container.NewMax(content))
}

// startNewGame handles starting a new game
func (m *UIManager) startNewGame() {
	// Clear existing state
	m.historyShelf.Objects = nil
	m.historyScroll.Refresh()
	m.currentIdx = -1
	m.currentImage = nil
	m.mainImage.Resource = nil
	m.mainImage.Refresh()
	m.usedIndices = make(map[int]bool)
	m.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all images from the img directory
	imgDir := "img"
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		log.Printf("Error reading image directory: %v", err)
		return
	}

	// Clear existing image list
	m.imageList = nil

	// Add all image files to the list
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				m.imageList = append(m.imageList, filepath.Join(imgDir, entry.Name()))
			}
		}
	}

	if len(m.imageList) == 0 {
		log.Printf("No images found in directory")
		return
	}

	log.Printf("Found %d images", len(m.imageList))

	// Start preloading images
	m.imageCache.PreloadImages(m.imageList)

	// Display first random image
	if len(m.imageList) > 0 {
		m.displayNextImage()
	}
}

func (m *UIManager) getRandomUnusedIndex() int {
	if len(m.imageList) == 0 {
		return -1
	}

	// If all indices are used, return -1
	if len(m.usedIndices) >= len(m.imageList) {
		return -1
	}

	// Keep trying until we find an unused index
	for {
		idx := m.rng.Intn(len(m.imageList))
		if !m.usedIndices[idx] {
			return idx
		}
	}
}

func (m *UIManager) displayNextImage() {
	if len(m.imageList) == 0 {
		log.Printf("No images in list")
		return
	}

	// Get a random unused index
	nextIdx := m.getRandomUnusedIndex()
	if nextIdx == -1 {
		log.Printf("No more unused images")
		return
	}

	nextPath := m.imageList[nextIdx]
	log.Printf("Loading next image: %s", nextPath)

	// Try to get from cache first
	if resource, ok := m.imageCache.Get(nextPath); ok {
		log.Printf("Found image in cache: %s", nextPath)
		
		// First move current image to history if it exists
		if m.currentImage != nil {
			m.addToHistory(m.currentImage)
		}

		// Then update main display with new image
		m.currentIdx = nextIdx
		m.usedIndices[nextIdx] = true
		m.updateMainDisplay(resource)
		return
	}

	log.Printf("Image not in cache, queueing: %s", nextPath)
	// If not in cache, preload it
	m.imageCache.PreloadImages([]string{nextPath})
	
	// Wait for the image in a goroutine
	go func() {
		resultChan := m.imageCache.GetResultChannel()
		for result := range resultChan {
			if result.Path == nextPath {
				if result.Err != nil {
					log.Printf("Error loading image %s: %v", nextPath, result.Err)
					return
				}
				log.Printf("Loaded image from queue: %s", nextPath)
				
				// First move current image to history if it exists
				if m.currentImage != nil {
					m.addToHistory(m.currentImage)
				}

				// Then update main display with new image
				m.currentIdx = nextIdx
				m.usedIndices[nextIdx] = true
				m.updateMainDisplay(result.Resource)
				return
			}
		}
	}()
}

func (m *UIManager) updateMainDisplay(resource fyne.Resource) {
	if resource == nil {
		log.Printf("Attempted to display nil resource")
		return
	}
	m.currentImage = resource
	m.mainImage.Resource = resource
	m.mainImage.Refresh()
	log.Printf("Updated display with image: %s", resource.Name())
}

func (m *UIManager) addToHistory(resource fyne.Resource) {
	if resource != nil {
		historyImg := canvas.NewImageFromResource(resource)
		historyImg.SetMinSize(fyne.NewSize(100, 100))
		historyImg.FillMode = canvas.ImageFillOriginal
		m.historyShelf.Add(historyImg)
		m.historyScroll.Refresh()
	}
}
