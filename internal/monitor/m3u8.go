package monitor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type M3U8Info struct {
	Version        int
	TargetDuration int
	MediaSequence  int
	Segments       []SegmentInfo
	Content        string
}

type SegmentInfo struct {
	Duration float64
	URI      string
}

func ParseM3U8(filePath string) (*M3U8Info, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &M3U8Info{
		Segments: make([]SegmentInfo, 0),
	}

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	var currentDuration float64

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		content.WriteString(line + "\n")

		if strings.HasPrefix(line, "#EXT-X-VERSION:") {
			version, _ := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-VERSION:"))
			info.Version = version
		} else if strings.HasPrefix(line, "#EXT-X-TARGETDURATION:") {
			duration, _ := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-TARGETDURATION:"))
			info.TargetDuration = duration
		} else if strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:") {
			sequence, _ := strconv.Atoi(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"))
			info.MediaSequence = sequence
		} else if strings.HasPrefix(line, "#EXTINF:") {
			durationStr := strings.TrimPrefix(line, "#EXTINF:")
			durationStr = strings.TrimSuffix(durationStr, ",")
			if idx := strings.Index(durationStr, ","); idx > 0 {
				durationStr = durationStr[:idx]
			}
			duration, _ := strconv.ParseFloat(durationStr, 64)
			currentDuration = duration
		} else if !strings.HasPrefix(line, "#") && line != "" {
			info.Segments = append(info.Segments, SegmentInfo{
				Duration: currentDuration,
				URI:      line,
			})
			currentDuration = 0
		}
	}

	info.Content = content.String()
	return info, scanner.Err()
}

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}