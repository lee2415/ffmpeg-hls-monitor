package ui

import (
	"fmt"
	"monitorMultiview/internal/monitor"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DetailViewModel struct {
	channelID      string
	viewport       viewport.Model
	ffmpegMonitor  *monitor.FFmpegMonitor
	hlsMonitor     *monitor.HLSMonitor
	lastUpdate     time.Time
	width          int
	height         int
	ready          bool
}

func NewDetailViewModel(channelID string, ffmpegMonitor *monitor.FFmpegMonitor, hlsMonitor *monitor.HLSMonitor) *DetailViewModel {
	return &DetailViewModel{
		channelID:     channelID,
		ffmpegMonitor: ffmpegMonitor,
		hlsMonitor:    hlsMonitor,
		lastUpdate:    time.Now(),
	}
}

func (m *DetailViewModel) Init() tea.Cmd {
	return tea.Batch(
		m.updateDetailData(),
		tickCmd(),
	)
}

func (m *DetailViewModel) Update(msg tea.Msg) (*DetailViewModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-8)
			m.viewport.YPosition = 3
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 8
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return SwitchToMainMsg{}
			}
		case "up", "down", "pgup", "pgdown":
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tickMsg:
		cmds = append(cmds, m.updateDetailData(), tickCmd())
	}

	return m, tea.Batch(cmds...)
}

func (m *DetailViewModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	title := TitleStyle.Render(fmt.Sprintf("Channel %s Details", strings.ToUpper(m.channelID)))
	
	helpBar := HelpStyle.Render(
		fmt.Sprintf(
			"Updated: %s  [Esc] Back to List  [↑↓] Scroll  [PgUp/PgDn] Page",
			m.lastUpdate.Format("15:04:05"),
		),
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.viewport.View(),
		helpBar,
	)
}

func (m *DetailViewModel) updateDetailData() tea.Cmd {
	return func() tea.Msg {
		var content strings.Builder
		
		// FFmpeg Process Information
		processes := m.ffmpegMonitor.GetProcesses()
		var process *monitor.FFmpegProcess
		for _, proc := range processes {
			if proc.ChannelID == m.channelID {
				process = proc
				break
			}
		}

		content.WriteString(HeaderStyle.Render("FFmpeg Process Information"))
		content.WriteString("\n\n")
		
		if process != nil {
			content.WriteString(fmt.Sprintf("Channel ID: %s\n", process.ChannelID))
			content.WriteString(fmt.Sprintf("Port: %d\n", process.Port))
			content.WriteString(fmt.Sprintf("PID: %d\n", process.PID))
			content.WriteString(fmt.Sprintf("Status: %s\n", GetStatusColor(process.Status).Render(process.Status)))
			content.WriteString(fmt.Sprintf("Command: %s\n", process.Command))
			content.WriteString(fmt.Sprintf("Last Seen: %s\n", process.LastSeen.Format("2006-01-02 15:04:05")))
		} else {
			content.WriteString(StatusStoppedStyle.Render("Process not running"))
		}

		content.WriteString("\n\n")

		// HLS Package Information
		pkg := m.hlsMonitor.GetPackageByChannel(m.channelID)
		
		content.WriteString(HeaderStyle.Render("HLS Package Information"))
		content.WriteString("\n\n")
		
		if pkg != nil {
			content.WriteString(fmt.Sprintf("Path: %s\n", pkg.Path))
			content.WriteString(fmt.Sprintf("Latest File: %s\n", pkg.LatestFile))
			content.WriteString(fmt.Sprintf("Total Segments: %d\n", pkg.SegmentCount))
			content.WriteString(fmt.Sprintf("Total Size: %s\n", monitor.FormatFileSize(pkg.TotalSize)))
			content.WriteString(fmt.Sprintf("Last Update: %s\n", pkg.LastUpdate.Format("2006-01-02 15:04:05")))
			
			content.WriteString(fmt.Sprintf("\nM3U8 Files (%d):\n", len(pkg.M3U8Files)))
			for i, m3u8File := range pkg.M3U8Files {
				content.WriteString(fmt.Sprintf("  %d. %s\n", i+1, m3u8File))
			}

			// Try to parse and display M3U8 content
			if len(pkg.M3U8Files) > 0 {
				m3u8Path := filepath.Join(pkg.Path, pkg.M3U8Files[0])
				m3u8Info, err := monitor.ParseM3U8(m3u8Path)
				if err == nil {
					content.WriteString("\n")
					content.WriteString(HeaderStyle.Render("M3U8 Content Details"))
					content.WriteString("\n\n")
					content.WriteString(fmt.Sprintf("Version: %d\n", m3u8Info.Version))
					content.WriteString(fmt.Sprintf("Target Duration: %d seconds\n", m3u8Info.TargetDuration))
					content.WriteString(fmt.Sprintf("Media Sequence: %d\n", m3u8Info.MediaSequence))
					content.WriteString(fmt.Sprintf("Total Segments in Playlist: %d\n", len(m3u8Info.Segments)))

					content.WriteString("\nLatest Segments:\n")
					segmentCount := len(m3u8Info.Segments)
					start := segmentCount - 10
					if start < 0 {
						start = 0
					}
					
					for i := start; i < segmentCount; i++ {
						segment := m3u8Info.Segments[i]
						marker := " "
						if i == segmentCount-1 {
							marker = "→"
						}
						content.WriteString(fmt.Sprintf("  %s %s (%.1fs)\n", 
							marker, segment.URI, segment.Duration))
					}

					content.WriteString("\n")
					content.WriteString(HeaderStyle.Render("M3U8 File Content"))
					content.WriteString("\n\n")
					
					// Show M3U8 content with syntax highlighting
					lines := strings.Split(m3u8Info.Content, "\n")
					for _, line := range lines {
						if strings.HasPrefix(line, "#") {
							if strings.HasPrefix(line, "#EXTM3U") {
								content.WriteString(lipgloss.NewStyle().Foreground(primaryColor).Render(line))
							} else if strings.HasPrefix(line, "#EXT") {
								content.WriteString(lipgloss.NewStyle().Foreground(secondaryColor).Render(line))
							} else {
								content.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(line))
							}
						} else if line != "" {
							content.WriteString(line)
						}
						content.WriteString("\n")
					}
				} else {
					content.WriteString(fmt.Sprintf("\nError parsing M3U8: %v\n", err))
				}
			}
		} else {
			content.WriteString(StatusStoppedStyle.Render("No HLS package found"))
		}

		m.viewport.SetContent(content.String())
		m.lastUpdate = time.Now()
		
		return nil
	}
}

type SwitchToMainMsg struct{}