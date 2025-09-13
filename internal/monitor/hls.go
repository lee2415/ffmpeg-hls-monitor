package monitor

import (
	"monitorMultiview/internal/config"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type HLSPackage struct {
	ChannelID    string
	Path         string
	M3U8Files    []string
	LatestFile   string
	LastUpdate   time.Time
	TotalSize    int64
	SegmentCount int
}

type HLSMonitor struct {
	packages map[string]*HLSPackage
}

func NewHLSMonitor() *HLSMonitor {
	return &HLSMonitor{
		packages: make(map[string]*HLSPackage),
	}
}

func (m *HLSMonitor) GetPackages() []*HLSPackage {
	m.updatePackages()
	
	result := make([]*HLSPackage, 0, len(m.packages))
	for _, pkg := range m.packages {
		result = append(result, pkg)
	}
	return result
}

func (m *HLSMonitor) GetPackageByChannel(channelID string) *HLSPackage {
	m.updatePackages()
	return m.packages[channelID]
}

func (m *HLSMonitor) updatePackages() {
	channels := config.GetChannels()
	for _, ch := range channels {
		pkg := m.scanHLSPackage(ch.ID, ch.Path)
		if pkg != nil {
			m.packages[ch.ID] = pkg
		}
	}
}

func (m *HLSMonitor) scanHLSPackage(channelID, path string) *HLSPackage {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &HLSPackage{
			ChannelID:  channelID,
			Path:       path,
			M3U8Files:  []string{},
			LatestFile: "N/A",
			LastUpdate: time.Now(),
		}
	}
	
	var m3u8Files []string
	var latestFile string
	var latestTime time.Time
	var totalSize int64
	segmentCount := 0
	
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if strings.HasSuffix(info.Name(), ".m3u8") {
			m3u8Files = append(m3u8Files, info.Name())
		}
		
		if strings.HasSuffix(info.Name(), ".ts") || strings.HasSuffix(info.Name(), ".m4s") {
			segmentCount++
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = info.Name()
			}
		}
		
		totalSize += info.Size()
		return nil
	})
	
	if err != nil {
		return nil
	}
	
	if latestFile == "" {
		latestFile = "N/A"
	}
	
	sort.Strings(m3u8Files)
	
	return &HLSPackage{
		ChannelID:    channelID,
		Path:         path,
		M3U8Files:    m3u8Files,
		LatestFile:   latestFile,
		LastUpdate:   time.Now(),
		TotalSize:    totalSize,
		SegmentCount: segmentCount,
	}
}

