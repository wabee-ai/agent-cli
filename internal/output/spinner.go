package output

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/wabee-ai/wabee-cli/internal/config"
)

// Spinner frames (CharSet 43 style - moving arrows)
var spinnerFrames = []string{
	"[    >>>>    ]",
	"[     >>>>   ]",
	"[      >>>>  ]",
	"[       >>>> ]",
	"[        >>>>]",
	"[        >>>>]",
	"[>       >>> ]",
	"[>>       >> ]",
	"[>>>       > ]",
	"[>>>>        ]",
	"[>>>>        ]",
	"[ >>>>       ]",
	"[  >>>>      ]",
	"[   >>>>     ]",
}

// Spinner provides an animated progress indicator
type Spinner struct {
	message string
	style   lipgloss.Style
	active  bool
	stopCh  chan struct{}
	mu      sync.Mutex
	index   int
}

// NewSpinner creates and starts a new animated spinner with the given message
func NewSpinner(message string) *Spinner {
	style := lipgloss.NewStyle()
	if config.UseColor() {
		style = style.
			Foreground(lipgloss.Color("#d2ff00")).
			Background(lipgloss.Color("#333333")).
			Bold(true).
			Padding(0, 1)
	}

	s := &Spinner{
		message: message,
		style:   style,
		active:  true,
		stopCh:  make(chan struct{}),
		index:   0,
	}

	go s.run()
	return s
}

// run animates the spinner in a background goroutine
func (s *Spinner) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			if s.active {
				s.render()
				s.index = (s.index + 1) % len(spinnerFrames)
			}
			s.mu.Unlock()
		}
	}
}

// render draws the current spinner frame
func (s *Spinner) render() {
	frame := spinnerFrames[s.index]
	output := frame + " " + s.message
	if config.UseColor() {
		output = s.style.Render(output)
	}
	fmt.Print("\r\033[K" + output)
}

// UpdateMessage changes the spinner message without stopping the animation
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Stop stops the spinner and clears the line
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active {
		s.active = false
		close(s.stopCh)
		fmt.Print("\r\033[K")
	}
}

// IsActive returns whether the spinner is currently running
func (s *Spinner) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}
