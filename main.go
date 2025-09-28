package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const statsURL = "http://srv.msk01.gigacorp.local/_stats"

type ServerStats struct {
	LoadAveragePercent    int64
	MemTotalBytes         int64
	MemUsageBytes         int64
	DiskTotalBytes        int64
	DiskUsageBytes        int64
	BandwidthTotalBytesps int64
	BandwidthUsageBytesps int64
}

func checkStats() error {
	resp, err := http.Get(statsURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(string(body), ",")
	if len(parts) != 7 {
		return fmt.Errorf("unexpected response format")
	}

	loadAverage, _ := strconv.ParseInt(parts[0], 10, 64)
	memTotal, _ := strconv.ParseInt(parts[1], 10, 64)
	memUsage, _ := strconv.ParseInt(parts[2], 10, 64)
	diskTotal, _ := strconv.ParseInt(parts[3], 10, 64)
	diskUsage, _ := strconv.ParseInt(parts[4], 10, 64)
	bandwidthTotal, _ := strconv.ParseInt(parts[5], 10, 64)
	bandwidthUsage, _ := strconv.ParseInt(parts[6], 10, 64)

	stats := ServerStats{
		LoadAveragePercent:    loadAverage,
		MemTotalBytes:         memTotal,
		MemUsageBytes:         memUsage,
		DiskTotalBytes:        diskTotal,
		DiskUsageBytes:        diskUsage,
		BandwidthTotalBytesps: bandwidthTotal,
		BandwidthUsageBytesps: bandwidthUsage,
	}

	if stats.LoadAveragePercent > 30 {
		fmt.Printf("Load Average is too high: %d\n", stats.LoadAveragePercent)
	}
	memUsedPercent := stats.MemUsageBytes * 100 / stats.MemTotalBytes
	if memUsedPercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memUsedPercent)
	}
	diskUsedPercent := stats.DiskUsageBytes * 100 / stats.DiskTotalBytes
	diskFreeMB := float64(stats.DiskTotalBytes-stats.DiskUsageBytes) / 1024 / 1024
	if diskUsedPercent > 90 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", int64(diskFreeMB))
	}
	bandwidthUsedPercent := stats.BandwidthUsageBytesps * 100 / stats.BandwidthTotalBytesps
	bandwidthFreeMBps := float64(stats.BandwidthTotalBytesps-stats.BandwidthUsageBytesps) / 1000 / 1000
	if bandwidthUsedPercent > 90 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int64(bandwidthFreeMBps))
	}
	resp.Body.Close()
	return nil
}

func main() {
	errCount := 0

	for {
		err := checkStats()
		if err != nil {
			errCount++
			if errCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
			}
			continue
		} else {
			errCount = 0
		}
		time.Sleep(time.Second)
	}
}
