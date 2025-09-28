package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const statsUrl = "http://srv.msk01.gigacorp.local/_stats"

type ServerStats struct {
	LoadAveragePercent    int
	MemFreeBytes          int
	MemUsageBytes         int
	DiskTotalBytes        int
	DiskUsageBytes        int
	BandwidthTotalBytesps int
	BandwidthUsageBytesps int
}

func fetchServerStats() (ServerStats, error) {
	resp, err := http.Get(statsUrl)
	if err != nil {
		return ServerStats{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return ServerStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var body []byte
	resp.Body.Read(body)

	parts := strings.Split(string(body), ",")
	if len(parts) != 6 {
		panic("Unexpected number of stats")
	}

	loadAverage, _ := strconv.Atoi(parts[0])
	memFree, _ := strconv.Atoi(parts[1])
	memUsage, _ := strconv.Atoi(parts[2])
	diskTotal, _ := strconv.Atoi(parts[3])
	diskUsage, _ := strconv.Atoi(parts[4])
	bandwidthTotal, _ := strconv.Atoi(parts[5])
	bandwidthUsage, _ := strconv.Atoi(parts[6])

	return ServerStats{
		LoadAveragePercent:    loadAverage,
		MemFreeBytes:          memFree,
		MemUsageBytes:         memUsage,
		DiskTotalBytes:        diskTotal,
		DiskUsageBytes:        diskUsage,
		BandwidthTotalBytesps: bandwidthTotal,
		BandwidthUsageBytesps: bandwidthUsage,
	}, nil
}

func checkStats(stats ServerStats) {
	if stats.LoadAveragePercent > 30 {
		fmt.Printf("Load Average is too high: %d\n", stats.LoadAveragePercent)
	}
	memUsedPercent := (stats.MemUsageBytes * 100) / (stats.MemFreeBytes + stats.MemUsageBytes)
	if memUsedPercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memUsedPercent)
	}
	diskUsedPercent := (stats.DiskUsageBytes * 100) / stats.DiskTotalBytes
	diskFreeMB := (stats.DiskTotalBytes - stats.DiskUsageBytes) / 1024 / 1024
	if diskUsedPercent > 90 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMB)
	}
	bandwidthUsedPercent := (stats.BandwidthUsageBytesps * 100) / stats.BandwidthTotalBytesps
	bandwidthFreeMBps := (stats.BandwidthTotalBytesps - stats.BandwidthUsageBytesps) / 1024 / 1024
	if bandwidthUsedPercent > 90 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available", bandwidthFreeMBps)
	}
}

func main() {
	errCount := 0

	for {
		stats, err := fetchServerStats()
		if err != nil {
			errCount++
			if errCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			continue
		} else {
			errCount = 0
		}
		checkStats(stats)
	}
}
