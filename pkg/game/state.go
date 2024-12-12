// Package game handles the core game logic for Holiday Bingo
package game

import (
	"fyne.io/fyne/v2"
	"math/rand"
	"sync"
	"time"
)

// GameState manages the state of the game
type GameState struct {
	mu            sync.RWMutex
	images        []fyne.Resource
	currentIndex  int
	isActive      bool
	imagesLoaded  int
	totalImages   int
	usedIndices   map[int]bool
	rng           *rand.Rand
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	return &GameState{
		images:      make([]fyne.Resource, 0),
		usedIndices: make(map[int]bool),
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Reset resets the game state
func (g *GameState) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.images = make([]fyne.Resource, 0)
	g.currentIndex = -1  // Start at -1 so first NextImage() sets to 0
	g.isActive = false
	g.imagesLoaded = 0
	g.totalImages = 0
	g.usedIndices = make(map[int]bool)
	g.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// AddImage adds an image to the game state and shuffles the deck
func (g *GameState) AddImage(image fyne.Resource) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.images = append(g.images, image)
	g.imagesLoaded++
	
	// Shuffle the entire deck
	for i := len(g.images) - 1; i > 0; i-- {
		j := g.rng.Intn(i + 1)
		g.images[i], g.images[j] = g.images[j], g.images[i]
	}
}

// GetLoadingProgress returns the loading progress as a float between 0 and 1
func (g *GameState) GetLoadingProgress() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.totalImages == 0 {
		return 0
	}
	return float64(g.imagesLoaded) / float64(g.totalImages)
}

// SetTotalImages sets the total number of images to load
func (g *GameState) SetTotalImages(total int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.totalImages = total
}

// GetImageCount returns the number of loaded images
func (g *GameState) GetImageCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.images)
}

// GetCurrentImage returns the current image
func (g *GameState) GetCurrentImage() fyne.Resource {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.currentIndex >= 0 && g.currentIndex < len(g.images) {
		return g.images[g.currentIndex]
	}
	return nil
}

// NextImage moves to and returns the next image
func (g *GameState) NextImage() fyne.Resource {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.isActive || len(g.images) == 0 {
		return nil
	}

	// If this is the first image, start at index 0
	if g.currentIndex == -1 {
		g.currentIndex = 0
		return g.images[0]
	}

	// Mark current index as used
	g.usedIndices[g.currentIndex] = true

	// Find next unused index
	nextIndex := (g.currentIndex + 1) % len(g.images)
	originalNext := nextIndex

	// Keep looking until we find an unused index or we've checked all indices
	for g.usedIndices[nextIndex] {
		nextIndex = (nextIndex + 1) % len(g.images)
		if nextIndex == originalNext {
			// We've checked all indices and they're all used
			g.isActive = false
			return nil
		}
	}

	g.currentIndex = nextIndex
	return g.images[nextIndex]
}

// IsActive returns whether the game is active
func (g *GameState) IsActive() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isActive
}

// SetActive sets the game's active state
func (g *GameState) SetActive(active bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.isActive = active
}
