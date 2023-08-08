package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var querySearch string

type model struct {
	isLoading bool
	items     []ServiceResponse
	progress  progress.Model
	percent   float64
	mu        *sync.Mutex
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tickMsg:
		m.mu.Lock()
		m.percent += 0.01
		m.mu.Unlock()

		if m.percent >= 1.0 {
			m.isLoading = false
			return m, nil
		}

		return m, tickCmd()

	}

	var cmd tea.Cmd
	return m, cmd
}

var styleTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).MarginBottom(1).MarginLeft(2).MarginTop(1).Bold(true).Render
var styleOk = lipgloss.NewStyle().Foreground(lipgloss.Color("#0ac782")).MarginBottom(1).MarginLeft(1).Render
var styleError = lipgloss.NewStyle().Foreground(lipgloss.Color("#cf5159")).MarginBottom(1).MarginLeft(1).Render

const (
	padding  = 2
	maxWidth = 80
)

func (m *model) View() string {

	if m.isLoading {
		pad := strings.Repeat(" ", padding)
		return " title " +
			pad + m.progress.ViewAs(m.percent) + "\n\n"
	}

	var s strings.Builder
	s.WriteString(styleTitle("Results for name " + querySearch))
	for _, item := range m.items {
		if item.exists {
			s.WriteString(styleError(fmt.Sprintf("\n \u2717 - %s - already exists - %s", item.name, item.url)))
		} else {
			s.WriteString(styleOk(fmt.Sprintf("\n \u2713 - %s - is available!", item.name)))
		}
	}

	return s.String()
}

func (m *model) getServicesInfo(querySearch string) {
	wg := sync.WaitGroup{}
	services := []NameChecker{Github{}, GoPkg{}, Homebrew{}}
	wg.Add(len(services))

	for _, s := range services {
		go func(s NameChecker) {
			defer wg.Done()
			r, err := s.Check(querySearch)

			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			m.mu.Lock()
			m.percent += float64(100/len(services)) / 100
			m.items = append(m.items, r)
			m.mu.Unlock()
		}(s)
	}

	wg.Wait()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: no package name")
		return
	}

	querySearch = os.Args[1]

	progress := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	m := model{
		progress:  progress,
		isLoading: true,
		mu:        &sync.Mutex{},
	}

	go m.getServicesInfo(querySearch)

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
