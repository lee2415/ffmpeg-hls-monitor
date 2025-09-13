package ui

import (
	"fmt"
	"monitorMultiview/internal/config"
	"monitorMultiview/internal/monitor"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainViewModel struct {
	ffmpegTable    table.Model
	hlsTable       table.Model
	selectedPanel  int // 0 = ffmpeg, 1 = hls
	selectedRow    int
	ffmpegMonitor  *monitor.FFmpegMonitor
	hlsMonitor     *monitor.HLSMonitor
	lastUpdate     time.Time
	width          int
	height         int
}

type tickMsg time.Time

func NewMainViewModel() *MainViewModel {
	// Create FFmpeg table with dynamic sizing
	ffmpegColumns := []table.Column{
		{Title: "Ch", Width: 5},
		{Title: "Port", Width: 8},
		{Title: "PID", Width: 8},
		{Title: "Status", Width: 8},
		{Title: "Command", Width: 40},
	}

	ffmpegTable := table.New(
		table.WithColumns(ffmpegColumns),
		table.WithFocused(true),
		table.WithHeight(25),
	)

	// Create HLS table with dynamic sizing
	hlsColumns := []table.Column{
		{Title: "Ch", Width: 5},
		{Title: "Path", Width: 25},
		{Title: "Latest File", Width: 18},
		{Title: "M3U8", Width: 12},
		{Title: "Segs", Width: 6},
		{Title: "Size", Width: 10},
	}

	hlsTable := table.New(
		table.WithColumns(hlsColumns),
		table.WithHeight(25),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	ffmpegTable.SetStyles(s)
	hlsTable.SetStyles(s)

	return &MainViewModel{
		ffmpegTable:   ffmpegTable,
		hlsTable:      hlsTable,
		selectedPanel: 0,
		ffmpegMonitor: monitor.NewFFmpegMonitor(),
		hlsMonitor:    monitor.NewHLSMonitor(),
		lastUpdate:    time.Now(),
	}
}

func (m *MainViewModel) Init() tea.Cmd {
	return tea.Batch(
		m.updateData(),
		tickCmd(),
	)
}

func (m *MainViewModel) Update(msg tea.Msg) (*MainViewModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSizes()

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.selectedPanel == 0 {
				m.selectedPanel = 1
				m.ffmpegTable.Blur()
				m.hlsTable.Focus()
			} else {
				m.selectedPanel = 0
				m.hlsTable.Blur()
				m.ffmpegTable.Focus()
			}

		case "enter":
			if m.selectedPanel == 1 { // HLS table
				selectedRow := m.hlsTable.Cursor()
				if selectedRow < len(m.hlsTable.Rows()) {
					channelID := m.hlsTable.Rows()[selectedRow][0] // First column is channel ID
					return m, func() tea.Msg {
						return SwitchToDetailMsg{ChannelID: channelID}
					}
				}
			}

		case "up", "down":
			if m.selectedPanel == 0 {
				m.ffmpegTable, cmd = m.ffmpegTable.Update(msg)
			} else {
				m.hlsTable, cmd = m.hlsTable.Update(msg)
			}
			cmds = append(cmds, cmd)
		}

	case tickMsg:
		cmds = append(cmds, m.updateData(), tickCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m *MainViewModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := TitleStyle.
		Width(m.width).
		Render(fmt.Sprintf("MultiView Monitor - %s (%d channels)", config.GlobalConfig.HLS.BasePath, config.GlobalConfig.Channels.Count))
	
	// Calculate widths for two-column layout that fills the screen
	leftWidth := (m.width - 6) / 2  // Account for borders and padding
	rightWidth := m.width - leftWidth - 6

	// Update table widths dynamically
	m.updateTableWidths(leftWidth, rightWidth)

	// FFmpeg panel with full height
	ffmpegTitle := HeaderStyle.Render(fmt.Sprintf("FFmpeg Processes (%d)", config.GlobalConfig.Channels.Count))
	ffmpegPanel := BaseStyle.Copy().
		Width(leftWidth).
		Height(m.height - 6).  // Fill available height
		Render(ffmpegTitle + "\n" + m.ffmpegTable.View())

	// HLS panel with full height  
	hlsTitle := HeaderStyle.Render(fmt.Sprintf("HLS Packages (%d)", config.GlobalConfig.Channels.Count))
	hlsPanel := BaseStyle.Copy().
		Width(rightWidth).
		Height(m.height - 6).  // Fill available height
		Render(hlsTitle + "\n" + m.hlsTable.View())

	// Status bar spanning full width
	runningCount := len(m.ffmpegMonitor.GetProcesses())
	totalPackages := len(m.hlsMonitor.GetPackages())
	statusBar := HelpStyle.
		Width(m.width).
		Render(fmt.Sprintf(
			"Status: %d/%d Running  Packages: %d/%d  Updated: %s  [Tab] Switch  [↑↓] Select  [Enter] Details  [q] Quit",
			runningCount, config.GlobalConfig.Channels.Count, 
			totalPackages, config.GlobalConfig.Channels.Count, 
			m.lastUpdate.Format("15:04:05"),
		))

	// Layout filling the entire screen
	content := lipgloss.JoinHorizontal(lipgloss.Top, ffmpegPanel, hlsPanel)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		statusBar,
	)
}

func (m *MainViewModel) updateData() tea.Cmd {
	return func() tea.Msg {
		// Update FFmpeg processes
		processes := m.ffmpegMonitor.GetProcesses()
		ffmpegRows := make([]table.Row, 0, 24)

		// Create a map for quick lookup
		processMap := make(map[string]*monitor.FFmpegProcess)
		for _, proc := range processes {
			processMap[proc.ChannelID] = proc
		}

		// Generate rows for all configured channels
		channels := config.GetChannels()
		for _, ch := range channels {
			if proc, exists := processMap[ch.ID]; exists {
				ffmpegRows = append(ffmpegRows, table.Row{
					ch.ID,
					fmt.Sprintf(":%d", proc.Port),
					fmt.Sprintf("%d", proc.PID),
					proc.Status,
					TruncateText(proc.Command, 40),
				})
			} else {
				ffmpegRows = append(ffmpegRows, table.Row{
					ch.ID,
					fmt.Sprintf(":%d", ch.Port),
					"-",
					"STOP",
					"Not running",
				})
			}
		}

		m.ffmpegTable.SetRows(ffmpegRows)

		// Update HLS packages
		packages := m.hlsMonitor.GetPackages()
		hlsRows := make([]table.Row, 0, 24)

		packageMap := make(map[string]*monitor.HLSPackage)
		for _, pkg := range packages {
			packageMap[pkg.ChannelID] = pkg
		}

		for _, ch := range channels {
			if pkg, exists := packageMap[ch.ID]; exists {
				m3u8Count := fmt.Sprintf("%d files", len(pkg.M3U8Files))
				if len(pkg.M3U8Files) == 1 {
					m3u8Count = pkg.M3U8Files[0]
					if len(m3u8Count) > 12 {
						m3u8Count = m3u8Count[:9] + "..."
					}
				}

				hlsRows = append(hlsRows, table.Row{
					ch.ID,
					TruncateText(filepath.Base(pkg.Path), 25),
					TruncateText(pkg.LatestFile, 18),
					m3u8Count,
					fmt.Sprintf("%d", pkg.SegmentCount),
					monitor.FormatFileSize(pkg.TotalSize),
				})
			} else {
				hlsRows = append(hlsRows, table.Row{
					ch.ID,
					TruncateText(filepath.Base(ch.Path), 25),
					"N/A",
					"No files",
					"0",
					"0 B",
				})
			}
		}

		m.hlsTable.SetRows(hlsRows)
		m.lastUpdate = time.Now()
		
		return nil
	}
}

func (m *MainViewModel) updateTableSizes() {
	if m.width > 0 {
		tableHeight := m.height - 8 // Reserve space for title and status
		if tableHeight > config.GlobalConfig.Channels.Count+2 {
			tableHeight = config.GlobalConfig.Channels.Count + 2 // Max channels + header
		}
		if tableHeight < 10 {
			tableHeight = 10 // Minimum height
		}

		// Create new tables with updated sizing
		m.recreateTablesWithSize(tableHeight)
	}
}

func (m *MainViewModel) recreateTablesWithSize(tableHeight int) {
	// Recreate FFmpeg table with new height
	leftWidth := (m.width - 6) / 2
	rightWidth := m.width - leftWidth - 6
	
	ffmpegColumns := []table.Column{
		{Title: "Ch", Width: 5},
		{Title: "Port", Width: 8},
		{Title: "PID", Width: 8},
		{Title: "Status", Width: 8},
		{Title: "Command", Width: max(leftWidth-35, 20)},
	}

	oldFocused := m.ffmpegTable.Focused()
	oldCursor := m.ffmpegTable.Cursor()
	
	m.ffmpegTable = table.New(
		table.WithColumns(ffmpegColumns),
		table.WithHeight(tableHeight),
	)
	
	if oldFocused {
		m.ffmpegTable.Focus()
	}
	m.ffmpegTable.SetCursor(oldCursor)

	// Recreate HLS table with new height
	hlsColumns := []table.Column{
		{Title: "Ch", Width: 5},
		{Title: "Path", Width: max(rightWidth-60, 15)},
		{Title: "Latest File", Width: 18},
		{Title: "M3U8", Width: 12},
		{Title: "Segs", Width: 6},
		{Title: "Size", Width: 10},
	}

	oldHLSCursor := m.hlsTable.Cursor()
	
	m.hlsTable = table.New(
		table.WithColumns(hlsColumns),
		table.WithHeight(tableHeight),
	)
	
	if !oldFocused {
		m.hlsTable.Focus()
	}
	m.hlsTable.SetCursor(oldHLSCursor)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	m.ffmpegTable.SetStyles(s)
	m.hlsTable.SetStyles(s)
}

func (m *MainViewModel) updateTableWidths(leftWidth, rightWidth int) {
	// Tables are recreated with proper sizing during window resize
	m.recreateTablesWithSize(m.height - 8)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type SwitchToDetailMsg struct {
	ChannelID string
}