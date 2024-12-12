package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ginjaninja78/holidaybingo/pkg/cache"
	"math/rand"
	"time"
)

// UIManager handles all UI components and their interactions
type UIManager struct {
	window        fyne.Window
	mainImage     *canvas.Image
	historyShelf  *fyne.Container
	historyScroll *container.Scroll
	loading       *LoadingState
	loadingLabel  *widget.Label
	imageCache    *cache.ImageCacheManager
	imageList     []string
	currentIdx    int
	isLoading     bool
	currentImage  fyne.Resource
}

// NewUIManager creates a new UI manager
func NewUIManager(window fyne.Window) *UIManager {
	loadingLabel := widget.NewLabel("")
	loadingLabel.Hide()

	manager := &UIManager{
		window:        window,
		mainImage:     canvas.NewImageFromFile(""),
		historyShelf:  container.NewHBox(),
		historyScroll: container.NewHScroll(nil), // Will set shelf later
		loadingLabel:  loadingLabel,
		loading:       NewLoadingState(loadingLabel),
		imageCache:    cache.NewImageCacheManager(cache.DefaultMaxWorkers),
		isLoading:     false,
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

	// Add loading label to main content
	loadingContainer := container.NewHBox(m.loadingLabel)
	mainContainer := container.NewBorder(
		loadingContainer,
		nil,
		nil,
		nil,
		content,
	)

	// Set window content
	m.window.SetContent(container.NewMax(mainContainer))
}

// startNewGame handles starting a new game
func (m *UIManager) startNewGame() {
	if m.isLoading {
		return
	}

	m.isLoading = true
	m.historyShelf.Objects = nil
	m.currentIdx = -1
	m.currentImage = nil

	m.loadingLabel.Show()
	m.loading.Start()

	go func() {
		defer func() {
			m.isLoading = false
			m.loading.Stop()
		}()

		// Get all images from the test directory
		m.imageList = []string{
			"/Users/christophermoore/CascadeProjects/HolidayBingo2/CascadeProjects/windsurf-project/holiday-bingo/test/image1.jpg",
			"/Users/christophermoore/CascadeProjects/HolidayBingo2/CascadeProjects/windsurf-project/holiday-bingo/test/image2.jpg",
		}

		// Randomize the image list
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(m.imageList), func(i, j int) {
			m.imageList[i], m.imageList[j] = m.imageList[j], m.imageList[i]
		})

		// Preload all images
		m.imageCache.PreloadImages(m.imageList)

		// Monitor loading progress
		resultChan := m.imageCache.GetResultChannel()
		for i := 0; i < len(m.imageList); i++ {
			select {
			case result := <-resultChan:
				if result.Err != nil {
					continue
				}
				// If this is the first image, display it
				if i == 0 {
					m.currentIdx = 0
					m.currentImage = result.Resource
					m.mainImage.Resource = result.Resource
					m.mainImage.Refresh()
				}
			}
		}
	}()
}

// displayNextImage handles the next button click
func (m *UIManager) displayNextImage() {
	if m.isLoading || len(m.imageList) == 0 || m.currentIdx >= len(m.imageList)-1 {
		return
	}

	// Add current image to history before changing
	if m.currentImage != nil {
		historyImg := canvas.NewImageFromResource(m.currentImage)
		historyImg.SetMinSize(fyne.NewSize(100, 100))
		historyImg.FillMode = canvas.ImageFillOriginal
		m.historyShelf.Add(historyImg)
		m.historyScroll.Refresh()
	}

	m.currentIdx++
	nextPath := m.imageList[m.currentIdx]

	// Try to get from cache first
	if resource, ok := m.imageCache.Get(nextPath); ok {
		m.updateMainDisplay(resource)
		return
	}

	// If not in cache, queue it and wait for result
	m.isLoading = true
	m.loading.Start()
	
	go func() {
		defer func() {
			m.isLoading = false
			m.loading.Stop()
		}()

		m.imageCache.QueueImage(nextPath, nil)
		select {
		case result := <-m.imageCache.GetResultChannel():
			if result.Path == nextPath && result.Err == nil {
				m.updateMainDisplay(result.Resource)
			}
		case <-time.After(3 * time.Second):
			// Timeout after 3 seconds
			return
		}
	}()
}

func (m *UIManager) updateMainDisplay(resource fyne.Resource) {
	m.currentImage = resource
	m.mainImage.Resource = resource
	m.mainImage.Refresh()
}
