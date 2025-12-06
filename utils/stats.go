package utils

import (
	"os"
	"path/filepath"
	"time"
)

var StartTime time.Time

func init() {
	StartTime = time.Now()
}

type DirStats struct {
	TotalSize int64 `json:"total_size_bytes"`
	FileCount int   `json:"file_count"`
}

type SystemStats struct {
	Uptime        string   `json:"uptime"`
	UptimeSeconds float64  `json:"uptime_seconds"`
	AvatarStats   DirStats `json:"avatar_stats"`
	ImageStats    DirStats `json:"image_stats"`
	CacheStats    DirStats `json:"cache_stats"`
	TotalUsage    int64    `json:"total_usage_bytes"`
}

func GetDirStats(path string) DirStats {
	var size int64
	var count int

	// Walk through directory
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	return DirStats{
		TotalSize: size,
		FileCount: count,
	}
}

func GetSystemStats(cfgAvatar, cfgImage string) SystemStats {
	avatarStats := GetDirStats(cfgAvatar)
	imageStats := GetDirStats(cfgImage)
	cacheStats := GetDirStats(filepath.Join("data", "cache"))

	totalUsage := avatarStats.TotalSize + imageStats.TotalSize + cacheStats.TotalSize

	return SystemStats{
		Uptime:        time.Since(StartTime).String(),
		UptimeSeconds: time.Since(StartTime).Seconds(),
		AvatarStats:   avatarStats,
		ImageStats:    imageStats,
		CacheStats:    cacheStats,
		TotalUsage:    totalUsage,
	}
}
