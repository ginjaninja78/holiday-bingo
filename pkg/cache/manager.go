package cache

import (
	"container/list"
	"fyne.io/fyne/v2"
	"sync"
	"time"
)

const (
	defaultMaxWorkers = 8  // Increased worker count
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
		maxWorkers = defaultMaxWorkers
	}

	manager := &ImageCacheManager{
		cache:      make(map[string]*list.Element),
		lruList:    list.New(),
		loadQueue:  make(chan string, maxWorkers*4),  // Increased buffer
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
		if _, exists := m.Get(path); !exists {
			if resource, err := m.loadAndOptimizeImage(path); err == nil {
				m.resultChan <- &ImageResult{Path: path, Resource: resource}
				m.Set(path, resource)
			} else {
				m.resultChan <- &ImageResult{Path: path, Err: err}
			}
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

	if element, exists := m.cache[path]; exists {
		m.lruList.MoveToFront(element)
		cachedImg := element.Value.(*CachedImage)
		cachedImg.resource = resource
		cachedImg.lastUsed = time.Now()
		return
	}

	if m.lruList.Len() >= m.maxSize {
		oldest := m.lruList.Back()
		if oldest != nil {
			m.lruList.Remove(oldest)
			delete(m.cache, oldest.Value.(*CachedImage).resource.Name())
		}
	}

	cachedImg := &CachedImage{
		resource: resource,
		lastUsed: time.Now(),
	}
	element := m.lruList.PushFront(cachedImg)
	m.cache[path] = element
}

func (m *ImageCacheManager) PreloadImages(paths []string) {
	m.wg.Add(len(paths))
	for _, path := range paths {
		select {
		case m.loadQueue <- path:
			// Successfully queued
		default:
			// Queue is full, start a new worker
			go func(p string) {
				m.loadQueue <- p
			}(path)
		}
	}
}

func (m *ImageCacheManager) WaitForLoad() {
	m.wg.Wait()
}

func (m *ImageCacheManager) GetResultChannel() <-chan *ImageResult {
	return m.resultChan
}
