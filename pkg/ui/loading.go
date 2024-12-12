package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2/widget"
)

// LoadingState manages the loading animation
type LoadingState struct {
	dots   int
	ticker *time.Ticker
	label  *widget.Label
	done   chan struct{}
}

// NewLoadingState creates a new loading state
func NewLoadingState(label *widget.Label) *LoadingState {
	return &LoadingState{
		label: label,
		done:  make(chan struct{}),
	}
}

// Start begins the loading animation
func (l *LoadingState) Start() {
	if l.ticker != nil {
		return // Already running
	}

	l.label.Show()
	l.label.SetText("Loading...")
	l.ticker = time.NewTicker(250 * time.Millisecond)
	timeout := time.NewTimer(3 * time.Second)

	go func() {
		defer func() {
			if l.ticker != nil {
				l.ticker.Stop()
				l.ticker = nil
			}
			l.label.Hide()
		}()

		for {
			select {
			case <-l.ticker.C:
				l.updateDots()
			case <-timeout.C:
				return
			case <-l.done:
				return
			}
		}
	}()
}

// updateDots updates the loading animation
func (l *LoadingState) updateDots() {
	l.dots = (l.dots + 1) % 4
	dots := ""
	for i := 0; i < l.dots; i++ {
		dots += "."
	}
	l.label.SetText(fmt.Sprintf("Loading%s", dots))
}

// Stop stops the loading animation
func (l *LoadingState) Stop() {
	select {
	case <-l.done:
		return
	default:
		close(l.done)
	}
}

// IsLoading returns whether the loading animation is active
func (l *LoadingState) IsLoading() bool {
	select {
	case <-l.done:
		return false
	default:
		return l.ticker != nil
	}
}
