package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
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

func fetchServerStats() (ServerStats, error) {
	resp, err := http.Get(statsURL)
	if err != nil {
		return ServerStats{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return ServerStats{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ServerStats{}, err
	}

	parts := strings.Split(string(body), ",")
	if len(parts) != 7 {
		return ServerStats{}, fmt.Errorf("unexpected response format")
	}

	loadAverage, _ := strconv.ParseInt(parts[0], 10, 64)
	memTotal, _ := strconv.ParseInt(parts[1], 10, 64)
	memUsage, _ := strconv.ParseInt(parts[2], 10, 64)
	diskTotal, _ := strconv.ParseInt(parts[3], 10, 64)
	diskUsage, _ := strconv.ParseInt(parts[4], 10, 64)
	bandwidthTotal, _ := strconv.ParseInt(parts[5], 10, 64)
	bandwidthUsage, _ := strconv.ParseInt(parts[6], 10, 64)

	return ServerStats{
		LoadAveragePercent:    loadAverage,
		MemTotalBytes:         memTotal,
		MemUsageBytes:         memUsage,
		DiskTotalBytes:        diskTotal,
		DiskUsageBytes:        diskUsage,
		BandwidthTotalBytesps: bandwidthTotal,
		BandwidthUsageBytesps: bandwidthUsage,
	}, nil
}

func checkStats(stats ServerStats, w io.Writer) {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	if stats.LoadAveragePercent > 30 {
		fmt.Fprintf(bw, "Load Average is too high: %d\n", stats.LoadAveragePercent)
	}
	memUsedPercent := stats.MemUsageBytes * 100 / stats.MemTotalBytes
	if memUsedPercent > 80 {
		fmt.Fprintf(bw, "Memory usage too high: %d%%\n", memUsedPercent)
	}
	diskUsedPercent := stats.DiskUsageBytes * 100 / stats.DiskTotalBytes
	diskFreeMB := float64(stats.DiskTotalBytes-stats.DiskUsageBytes) / 1024 / 1024
	if diskUsedPercent > 90 {
		fmt.Fprintf(bw, "Free disk space is too low: %d Mb left\n", int64(diskFreeMB))
	}
	bandwidthUsedPercent := stats.BandwidthUsageBytesps * 100 / stats.BandwidthTotalBytesps
	bandwidthFreeMBps := float64(stats.BandwidthTotalBytesps-stats.BandwidthUsageBytesps) / 1000 / 1000
	if bandwidthUsedPercent > 90 {
		fmt.Fprintf(bw, "Network bandwidth usage high: %d Mbit/s available\n", int64(bandwidthFreeMBps))
	}
	time.Sleep(20 * time.Millisecond)
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
		checkStats(stats, os.Stdout)
	}
}
