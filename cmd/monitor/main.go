package main

import (
	"fmt"
	"log"
	"monitorMultiview/internal/config"
	"monitorMultiview/internal/monitor"
	"monitorMultiview/internal/ui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	currentView   string // "main" or "detail"
	mainView      *ui.MainViewModel
	detailView    *ui.DetailViewModel
	ffmpegMonitor *monitor.FFmpegMonitor
	hlsMonitor    *monitor.HLSMonitor
}

func (m Model) Init() tea.Cmd {
	if m.currentView == "main" {
		return m.mainView.Init()
	}
	return m.detailView.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case ui.SwitchToDetailMsg:
		m.currentView = "detail"
		m.detailView = ui.NewDetailViewModel(msg.ChannelID, m.ffmpegMonitor, m.hlsMonitor)
		return m, m.detailView.Init()

	case ui.SwitchToMainMsg:
		m.currentView = "main"
		return m, m.mainView.Init()
	}

	var cmd tea.Cmd
	if m.currentView == "main" {
		m.mainView, cmd = m.mainView.Update(msg)
	} else {
		m.detailView, cmd = m.detailView.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.currentView == "main" {
		return m.mainView.View()
	}
	return m.detailView.View()
}

func main() {
	// Initialize configuration with command line arguments
	config.InitConfig()

	// Initialize monitors
	ffmpegMonitor := monitor.NewFFmpegMonitor()
	hlsMonitor := monitor.NewHLSMonitor()

	// Initialize main view
	mainView := ui.NewMainViewModel()

	// Create model
	model := Model{
		currentView:   "main",
		mainView:      mainView,
		ffmpegMonitor: ffmpegMonitor,
		hlsMonitor:    hlsMonitor,
	}

	// Create program with full screen mode
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

func init() {
	// Configure logging to file instead of stdout to avoid interfering with TUI
	logFile, err := os.OpenFile("monitor.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
}