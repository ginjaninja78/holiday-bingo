package cache

import (
	"container/list"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

const (
	DefaultMaxWorkers = 8   // Increased worker count
	defaultCacheSize  = 200 // Increased cache size
)

type CachedImage struct {
	resource fyne.Resource
	lastUsed time.Time
	size     int64
}

type ImageCacheManager struct {
	cache      map[string]*list.Element
	lruList    *list.List
	cacheMux   sync.RWMutex
	loadQueue  chan string
	resultChan chan *ImageResult
	maxWorkers int
	maxSize    int
	wg         sync.WaitGroup
}

type ImageResult struct {
	Path     string
	Resource fyne.Resource
	Err      error
}

func NewImageCacheManager(maxWorkers int) *ImageCacheManager {
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxWorkers
	}

	manager := &ImageCacheManager{
		cache:      make(map[string]*list.Element),
		lruList:    list.New(),
		loadQueue:  make(chan string, maxWorkers*4), // Increased buffer
		resultChan: make(chan *ImageResult, maxWorkers*4),
		maxWorkers: maxWorkers,
		maxSize:    defaultCacheSize,
	}

	manager.startWorkers()
	return manager
}

func (m *ImageCacheManager) startWorkers() {
	for i := 0; i < m.maxWorkers; i++ {
		go m.worker()
	}
}

func (m *ImageCacheManager) worker() {
	for path := range m.loadQueue {
		// Skip if already cached
		if _, exists := m.Get(path); exists {
			m.wg.Done()
			continue
		}

		// Load and optimize the image
		resource, err := m.loadAndOptimizeImage(path)

		// Always send result, even on error
		result := &ImageResult{
			Path:     path,
			Resource: resource,
			Err:      err,
		}
		m.resultChan <- result

		// Only cache if successful
		if err == nil {
			m.Set(path, resource)
		}
		
		m.wg.Done()
	}
}

func (m *ImageCacheManager) Get(path string) (fyne.Resource, bool) {
	m.cacheMux.RLock()
	defer m.cacheMux.RUnlock()

	if element, exists := m.cache[path]; exists {
		m.lruList.MoveToFront(element)
		cachedImg := element.Value.(*CachedImage)
		cachedImg.lastUsed = time.Now()
		return cachedImg.resource, true
	}
	return nil, false
}

func (m *ImageCacheManager) Set(path string, resource fyne.Resource) {
	m.cacheMux.Lock()
	defer m.cacheMux.Unlock()

	// If path already exists, update it
	if element, exists := m.cache[path]; exists {
		m.lruList.MoveToFront(element)
		cachedImg := element.Value.(*CachedImage)
		cachedImg.resource = resource
		cachedImg.lastUsed = time.Now()
		return
	}

	// Create new cache entry
	cachedImg := &CachedImage{
		resource: resource,
		lastUsed: time.Now(),
	}

	// Add to front of LRU list and cache
	element := m.lruList.PushFront(cachedImg)
	m.cache[path] = element

	// Remove oldest if we're over capacity
	for m.lruList.Len() > m.maxSize {
		oldest := m.lruList.Back()
		if oldest != nil {
			m.lruList.Remove(oldest)
			// Find and remove from cache map
			for path, element := range m.cache {
				if element == oldest {
					delete(m.cache, path)
					break
				}
			}
		}
	}
}

func (m *ImageCacheManager) PreloadImages(paths []string) {
	for _, path := range paths {
		if _, exists := m.Get(path); exists {
			continue // Skip if already cached
		}
		m.wg.Add(1)
		m.loadQueue <- path
	}
}

func (m *ImageCacheManager) WaitForLoad() {
	m.wg.Wait()
}

func (m *ImageCacheManager) GetResultChannel() <-chan *ImageResult {
	return m.resultChan
}
