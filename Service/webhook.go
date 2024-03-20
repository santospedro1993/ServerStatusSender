package main

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Stats struct {
	// Add any other stats you are interested in here.
	MemoryUsage uint64
	CpuUsage    float64
}

// ContainerStats structure holds container and its stats
type ContainerStats struct {
	ContainerName string `json:"containerName"`
	Stats         *Stats `json:"stats"`
}

func getStats(cli *client.Client, container *types.Container) *Stats {
	stats, _ := cli.ContainerStats(context.Background(), container.ID, false)

	data := &types.StatsJSON{}
	json.NewDecoder(stats.Body).Decode(&data)
	defer stats.Body.Close()

	return &Stats{
		MemoryUsage: data.MemoryStats.Usage,
		CpuUsage:    calculateCPUPercent(data),
	}
}

func calculateCPUPercent(stats *types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	cpuPercent := (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	return cpuPercent
}

func sendStatsToWebhook(webhookURL string, stats []ContainerStats) {
	data, err := json.Marshal(stats)
	if err != nil {
		log.Fatalf("Failed to marshal stats: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Failed to send POST request: %v", err)
		return
	}
	defer resp.Body.Close()
}

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
		os.Exit(1)
	}

	webhookURL := "<your_webhook_url_here>"

	for {
		containers, _ := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		containerStats := make([]ContainerStats, 0, len(containers))
		for _, container := range containers {
			containerStats = append(containerStats, ContainerStats{
				ContainerName: container.Names[0],
				Stats:         getStats(cli, &container),
			})
		}
		sendStatsToWebhook(webhookURL, containerStats)

		time.Sleep(1 * time.Minute)
	}
}
