package logrotate

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	Monthly  RotationDuration = 'M'
	Weekly   RotationDuration = 'W'
	Daily    RotationDuration = 'D'
	Hourly   RotationDuration = 'h'
	Minutely RotationDuration = 'm'
	Yearly   RotationDuration = 'Y'
)

type RotationDuration byte

func (this RotationDuration) Valid() bool {
	switch this {
	case 0, Monthly, Weekly, Daily, Hourly, Minutely, Yearly:
		return true
	}
	return false
}

type Options struct {
	HistoryPath  string // path inside root
	MaxSize      int64 // -1 disables limitter
	Duration     RotationDuration
	FileMode     os.FileMode
	DirMode      os.FileMode
	HistoryDir   string
	HistoryCount int
}

type Config struct {
	MaxSize      string `yaml:"max_size" mapstructure:"max_size"`
	Duration     string
	FileMode     os.FileMode `yaml:"file_mode" mapstructure:"file_mode"`
	DirMode      os.FileMode `yaml:"dir_mode" mapstructure:"dir_mode"`
	HistoryDir   string      `yaml:"history_dir" mapstructure:"history_dir"`
	HistoryPath  string      `yaml:"history_path" mapstructure:"history_path"`
	HistoryCount int         `yaml:"history_count" mapstructure:"history_count"`
}

func (this Config) Yaml() string {
	return fmt.Sprintf(`max_size: %s
duration: %s
file_mode: 0%o
dir_mode: 0%o
history_dir: %q
history_path: %q
history_count: %d`,
		this.MaxSize,
		this.Duration,
		this.FileMode,
		this.DirMode,
		this.HistoryDir,
		this.HistoryPath,
		this.HistoryCount)
}

func (this Config) Options() (opt Options, err error) {
	var d RotationDuration
	if this.Duration != "" {
		d = RotationDuration(this.Duration[0])
		if !d.Valid() {
			err = fmt.Errorf("bad duration %q", string(d))
			return
		}
	}

	var s int

	if this.MaxSize != "" {
		var (
			m       float64
			c       = strings.ToUpper(string(this.MaxSize[len(this.MaxSize)-1]))[0]
			maxSize = this.MaxSize[0 : len(this.MaxSize)-1]
		)
		switch c {
		case 'K':
			m = 3
		case 'M':
			m = 6
		case 'G':
			m = 9
		case 'T':
			m = 12
		default:
			maxSize += string(c)
		}
		if s, err = strconv.Atoi(maxSize); err != nil {
			return
		}
		s *= int(math.Pow(10, m))
	}

	return Options{
		HistoryDir:   this.HistoryDir,
		HistoryPath:  this.HistoryPath,
		MaxSize:      int64(s),
		HistoryCount: this.HistoryCount,
		Duration:     d,
		FileMode:     this.FileMode,
		DirMode:      this.DirMode,
	}, nil
}
