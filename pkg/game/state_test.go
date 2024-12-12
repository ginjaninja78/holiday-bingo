package game

import (
	"testing"

	"github.com/ginjaninja78/holidaybingo/pkg/cache"
)

type mockResource struct {
	name string
}

func (m mockResource) Name() string {
	return m.name
}

func (m mockResource) Content() []byte {
	return []byte{}
}

func TestGameState(t *testing.T) {
	cacheManager := cache.NewImageCacheManager(4)
	state := NewGameState(cacheManager)

	t.Run("Initial State", func(t *testing.T) {
		if state.IsActive() {
			t.Error("New game state should not be active")
		}
		if state.GetImageCount() != 0 {
			t.Error("New game state should have no images")
		}
	})

	t.Run("Add Image", func(t *testing.T) {
		resource := mockResource{name: "test1.jpg"}
		state.AddImage(resource)
		if state.GetImageCount() != 1 {
			t.Error("Image count should be 1 after adding an image")
		}
		// Test duplicate image
		state.AddImage(resource)
		if state.GetImageCount() != 1 {
			t.Error("Duplicate image should not be added")
		}
	})

	t.Run("Game Active State", func(t *testing.T) {
		state.SetActive(true)
		if !state.IsActive() {
			t.Error("Game should be active after SetActive(true)")
		}
		state.SetActive(false)
		if state.IsActive() {
			t.Error("Game should not be active after SetActive(false)")
		}
	})

	t.Run("Next Image", func(t *testing.T) {
		state.Reset()
		state.SetActive(true)
		
		// Add test images
		state.AddImage(mockResource{name: "test1.jpg"})
		state.AddImage(mockResource{name: "test2.jpg"})
		
		first := state.GetCurrentImage()
		if first == nil {
			t.Error("Should have current image")
		}
		
		next := state.NextImage()
		if next == nil {
			t.Error("Should have next image")
		}
		if next.Name() == first.Name() {
			t.Error("Next image should be different from first")
		}
	})

	t.Run("Loading Progress", func(t *testing.T) {
		state.Reset()
		state.SetTotalImages(4)
		
		if progress := state.GetLoadingProgress(); progress != 0 {
			t.Errorf("Initial progress should be 0, got %f", progress)
		}
		
		state.AddImage(mockResource{name: "test1.jpg"})
		state.AddImage(mockResource{name: "test2.jpg"})
		
		expected := 0.5 // 2/4
		if progress := state.GetLoadingProgress(); progress != expected {
			t.Errorf("Progress should be %f, got %f", expected, progress)
		}
	})
}
