package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HLS HLSConfig `yaml:"hls"`
	FFmpeg FFmpegConfig `yaml:"ffmpeg"`
	Channels ChannelsConfig `yaml:"channels"`
	UI UIConfig `yaml:"ui"`
	Logging LoggingConfig `yaml:"logging"`
	App AppConfig `yaml:"app"`
}

type HLSConfig struct {
	BasePath string `yaml:"base_path"`
	ChannelDirPattern string `yaml:"channel_dir_pattern"`
}

type FFmpegConfig struct {
	StartPort int `yaml:"start_port"`
	PortIncrement int `yaml:"port_increment"`
}

type ChannelsConfig struct {
	Count int `yaml:"count"`
	IDFormat string `yaml:"id_format"`
	NameFormat string `yaml:"name_format"`
}

type UIConfig struct {
	RefreshInterval int `yaml:"refresh_interval"`
	Fullscreen bool `yaml:"fullscreen"`
	Theme string `yaml:"theme"`
}

type LoggingConfig struct {
	File string `yaml:"file"`
	Level string `yaml:"level"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Version string `yaml:"version"`
	Description string `yaml:"description"`
}

type Channel struct {
	ID     string
	Name   string
	Port   int
	Path   string
}

var GlobalConfig = Config{
	HLS: HLSConfig{
		BasePath: "/output",
		ChannelDirPattern: "channel%02d",
	},
	FFmpeg: FFmpegConfig{
		StartPort: 8001,
		PortIncrement: 1,
	},
	Channels: ChannelsConfig{
		Count: 24,
		IDFormat: "ch%02d",
		NameFormat: "Channel %02d",
	},
	UI: UIConfig{
		RefreshInterval: 1,
		Fullscreen: true,
		Theme: "dark",
	},
	Logging: LoggingConfig{
		File: "monitor.log",
		Level: "info",
	},
	App: AppConfig{
		Name: "MultiView Monitor",
		Version: "1.0.0",
		Description: "Real-time FFmpeg and HLS monitoring tool",
	},
}

func InitConfig() {
	var configFile string
	var hlsPath string
	var channelCount int
	var startPort int
	var generateConfig bool

	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.StringVar(&configFile, "f", "", "Path to configuration file (short)")
	flag.BoolVar(&generateConfig, "generate-config", false, "Generate default configuration file")
	flag.StringVar(&hlsPath, "hls-path", "", "Base path for HLS package directories")
	flag.StringVar(&hlsPath, "p", "", "Base path for HLS package directories (short)")
	flag.IntVar(&channelCount, "channels", 0, "Number of channels to monitor")
	flag.IntVar(&channelCount, "c", 0, "Number of channels to monitor (short)")
	flag.IntVar(&startPort, "start-port", 0, "Starting port number for FFmpeg processes")
	flag.IntVar(&startPort, "s", 0, "Starting port number for FFmpeg processes (short)")

	help := flag.Bool("help", false, "Show help message")
	flag.BoolVar(help, "h", false, "Show help message (short)")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if generateConfig {
		if err := GenerateConfigFile("multiview-monitor.yaml"); err != nil {
			fmt.Printf("Error generating config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Generated default configuration file: multiview-monitor.yaml")
		fmt.Println("Edit the file and run the program with: ./multiview-monitor -f multiview-monitor.yaml")
		os.Exit(0)
	}

	// Load configuration file if specified
	if configFile != "" {
		if err := LoadConfigFile(configFile); err != nil {
			fmt.Printf("Error loading config file %s: %v\n", configFile, err)
			os.Exit(1)
		}
	} else {
		// Try to load from default locations
		defaultPaths := []string{
			"multiview-monitor.yaml",
			"configs/multiview-monitor.yaml",
			filepath.Join(os.Getenv("HOME"), ".multiview-monitor.yaml"),
		}
		
		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("Loading configuration from: %s\n", path)
				if err := LoadConfigFile(path); err != nil {
					fmt.Printf("Warning: Error loading config file %s: %v\n", path, err)
				} else {
					break
				}
			}
		}
	}

	// Command line arguments override config file settings
	if hlsPath != "" {
		GlobalConfig.HLS.BasePath = hlsPath
	}
	if channelCount > 0 {
		GlobalConfig.Channels.Count = channelCount
	}
	if startPort > 0 {
		GlobalConfig.FFmpeg.StartPort = startPort
	}

	// Validate configuration
	if err := ValidateConfig(); err != nil {
		fmt.Printf("Configuration validation error: %v\n", err)
		os.Exit(1)
	}

	printCurrentConfig()
}

func LoadConfigFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Merge with global config (only non-zero values)
	if config.HLS.BasePath != "" {
		GlobalConfig.HLS.BasePath = config.HLS.BasePath
	}
	if config.HLS.ChannelDirPattern != "" {
		GlobalConfig.HLS.ChannelDirPattern = config.HLS.ChannelDirPattern
	}
	if config.FFmpeg.StartPort > 0 {
		GlobalConfig.FFmpeg.StartPort = config.FFmpeg.StartPort
	}
	if config.FFmpeg.PortIncrement > 0 {
		GlobalConfig.FFmpeg.PortIncrement = config.FFmpeg.PortIncrement
	}
	if config.Channels.Count > 0 {
		GlobalConfig.Channels.Count = config.Channels.Count
	}
	if config.Channels.IDFormat != "" {
		GlobalConfig.Channels.IDFormat = config.Channels.IDFormat
	}
	if config.Channels.NameFormat != "" {
		GlobalConfig.Channels.NameFormat = config.Channels.NameFormat
	}
	if config.UI.RefreshInterval > 0 {
		GlobalConfig.UI.RefreshInterval = config.UI.RefreshInterval
	}
	if config.UI.Theme != "" {
		GlobalConfig.UI.Theme = config.UI.Theme
	}
	if config.Logging.File != "" {
		GlobalConfig.Logging.File = config.Logging.File
	}
	if config.Logging.Level != "" {
		GlobalConfig.Logging.Level = config.Logging.Level
	}
	if config.App.Name != "" {
		GlobalConfig.App.Name = config.App.Name
	}
	if config.App.Version != "" {
		GlobalConfig.App.Version = config.App.Version
	}
	if config.App.Description != "" {
		GlobalConfig.App.Description = config.App.Description
	}

	return nil
}

func GenerateConfigFile(filename string) error {
	data, err := yaml.Marshal(&GlobalConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	header := `# MultiView Monitor Configuration File
# This file contains settings for the MultiView Monitor application
# 
# Priority: Command line arguments > Config file > Default values
#
# Examples:
#   ./multiview-monitor -f multiview-monitor.yaml
#   ./multiview-monitor -f custom.yaml -c 12 -p /custom/path
#
# Generate new config: ./multiview-monitor --generate-config

`

	fullData := []byte(header + string(data))
	
	if err := os.WriteFile(filename, fullData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ValidateConfig() error {
	if GlobalConfig.Channels.Count <= 0 || GlobalConfig.Channels.Count > 999 {
		return fmt.Errorf("invalid channel count: %d (must be 1-999)", GlobalConfig.Channels.Count)
	}
	if GlobalConfig.FFmpeg.StartPort <= 0 || GlobalConfig.FFmpeg.StartPort > 65535 {
		return fmt.Errorf("invalid start port: %d (must be 1-65535)", GlobalConfig.FFmpeg.StartPort)
	}
	if GlobalConfig.HLS.BasePath == "" {
		return fmt.Errorf("HLS base path cannot be empty")
	}
	if GlobalConfig.UI.RefreshInterval <= 0 {
		return fmt.Errorf("refresh interval must be positive: %d", GlobalConfig.UI.RefreshInterval)
	}
	
	return nil
}

func printHelp() {
	fmt.Printf("%s - %s\n", GlobalConfig.App.Name, GlobalConfig.App.Description)
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Printf("  %s [options]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  -f, --config string        Path to configuration file")
	fmt.Println("  --generate-config          Generate default configuration file")
	fmt.Println("")
	fmt.Println("Options (override config file):")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Printf("  %s --generate-config\n", os.Args[0])
	fmt.Printf("  %s -f myconfig.yaml\n", os.Args[0])
	fmt.Printf("  %s -f myconfig.yaml -c 12 -p /custom/path\n", os.Args[0])
	fmt.Printf("  %s -p /data/hls -c 12 -s 9001\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Config file locations (checked in order):")
	fmt.Println("  1. ./multiview-monitor.yaml")
	fmt.Println("  2. ./configs/multiview-monitor.yaml")
	fmt.Println("  3. ~/.multiview-monitor.yaml")
	fmt.Println("")
	fmt.Println("Keyboard shortcuts:")
	fmt.Println("  Tab       - Switch between FFmpeg and HLS panels")
	fmt.Println("  ↑/↓       - Navigate channels")
	fmt.Println("  Enter     - View channel details")
	fmt.Println("  Esc       - Return to main view")
	fmt.Println("  q         - Quit")
}

func printCurrentConfig() {
	fmt.Printf("Starting %s v%s with:\n", GlobalConfig.App.Name, GlobalConfig.App.Version)
	fmt.Printf("  HLS Base Path: %s\n", GlobalConfig.HLS.BasePath)
	fmt.Printf("  Channels: %d\n", GlobalConfig.Channels.Count)
	fmt.Printf("  Start Port: %d\n", GlobalConfig.FFmpeg.StartPort)
	fmt.Printf("  Refresh Interval: %ds\n", GlobalConfig.UI.RefreshInterval)
	fmt.Println("")
}

func GetChannels() []Channel {
	channels := make([]Channel, GlobalConfig.Channels.Count)
	for i := 0; i < GlobalConfig.Channels.Count; i++ {
		channelNum := i + 1
		channels[i] = Channel{
			ID:   fmt.Sprintf(GlobalConfig.Channels.IDFormat, channelNum),
			Name: fmt.Sprintf(GlobalConfig.Channels.NameFormat, channelNum),
			Port: GlobalConfig.FFmpeg.StartPort + (i * GlobalConfig.FFmpeg.PortIncrement),
			Path: filepath.Join(GlobalConfig.HLS.BasePath, fmt.Sprintf(GlobalConfig.HLS.ChannelDirPattern, channelNum)),
		}
	}
	return channels
}

func GetChannelByID(id string) *Channel {
	channels := GetChannels()
	for _, ch := range channels {
		if ch.ID == id {
			return &ch
		}
	}
	return nil
}