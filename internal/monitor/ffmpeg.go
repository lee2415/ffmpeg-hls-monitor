package monitor

import (
	"fmt"
	"monitorMultiview/internal/config"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type FFmpegProcess struct {
	ChannelID string
	Port      int
	PID       int
	Status    string
	Command   string
	LastSeen  time.Time
}

type FFmpegMonitor struct {
	processes map[string]*FFmpegProcess
}

func NewFFmpegMonitor() *FFmpegMonitor {
	return &FFmpegMonitor{
		processes: make(map[string]*FFmpegProcess),
	}
}

func (m *FFmpegMonitor) GetProcesses() []*FFmpegProcess {
	m.updateProcesses()
	
	result := make([]*FFmpegProcess, 0, len(m.processes))
	for _, proc := range m.processes {
		result = append(result, proc)
	}
	return result
}

func (m *FFmpegMonitor) updateProcesses() {
	processes := m.scanFFmpegProcesses()
	processMap := make(map[string]*FFmpegProcess)
	
	for _, proc := range processes {
		processMap[proc.ChannelID] = proc
	}
	
	m.processes = processMap
}

func (m *FFmpegMonitor) scanFFmpegProcesses() []*FFmpegProcess {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return []*FFmpegProcess{}
	}
	
	lines := strings.Split(string(output), "\n")
	var processes []*FFmpegProcess
	
	for _, line := range lines {
		if !strings.Contains(line, "ffmpeg") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}
		
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		
		cmdLine := strings.Join(fields[10:], " ")
		if len(cmdLine) > 50 {
			cmdLine = cmdLine[:47] + "..."
		}
		
		port := m.extractPortFromCommand(cmdLine)
		channelID := m.getChannelIDFromPort(port)
		
		if channelID == "" {
			continue
		}
		
		processes = append(processes, &FFmpegProcess{
			ChannelID: channelID,
			Port:      port,
			PID:       pid,
			Status:    "RUN",
			Command:   cmdLine,
			LastSeen:  time.Now(),
		})
	}
	
	return processes
}

func (m *FFmpegMonitor) extractPortFromCommand(cmdLine string) int {
	channels := config.GetChannels()
	for _, ch := range channels {
		if strings.Contains(cmdLine, fmt.Sprintf(":%d", ch.Port)) {
			return ch.Port
		}
	}
	return 0
}

func (m *FFmpegMonitor) getChannelIDFromPort(port int) string {
	channels := config.GetChannels()
	for _, ch := range channels {
		if ch.Port == port {
			return ch.ID
		}
	}
	return ""
}