package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const Gigabyte = 1024 * 1024 * 1024
const Megabyte = 1024 * 1024

type CPUPercentage struct {
	Utilization float64
}

type MemoryUsage struct {
	TotalMemory       float32
	FreeMemory        float32
	UsedMemoryPercent float64
}

type NetworkUsage struct {
	Name         string
	ReceivedRate float64
	SentRate     float64
}

type DiskUsage struct {
	Mount   string
	TotalGB float64
	FreeGB  float64
}

type Report struct {
	CPUUsage     []CPUPercentage
	MemoryUsage  []MemoryUsage
	NetworkUsage []NetworkUsage
	DiskUsage    []DiskUsage
}

func (r *Report) Generate() error {
	var err error

	if r.CPUUsage, err = ReportCPUPercentage(); err != nil {
		return err
	}
	if r.MemoryUsage, err = ReportMemoryUsage(); err != nil {
		return err
	}
	if r.NetworkUsage, err = ReportNetworkUsage(); err != nil {
		return err
	}
	if r.DiskUsage, err = ReportDisk(); err != nil {
		return err
	}

	return nil
}

func main() {
	report := &Report{}
	for {
		numCPU := runtime.NumCPU()

		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err == nil {
			containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
			if err == nil {
				for _, container := range containers {
					for _, name := range container.Names {
						fmt.Println(strings.TrimPrefix(name, "/"))
					}
				}
			}
		}

		fmt.Printf("Number of logical CPUs: %d\n", numCPU)

		if err := report.Generate(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error generating report:", err)
			continue
		}
		clearScreen()
		fmt.Printf("Report: %+v\n", report)
		time.Sleep(time.Second)
	}
}

func ReportCPUPercentage() ([]CPUPercentage, error) {
	cpuUsagePercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	cpuUsage := make([]CPUPercentage, len(cpuUsagePercent))
	for i, utilization := range cpuUsagePercent {
		cpuUsage[i] = CPUPercentage{Utilization: utilization}
	}
	return cpuUsage, nil
}

func ReportMemoryUsage() ([]MemoryUsage, error) {
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	totalMemory := float32(virtualMemory.Total) / Gigabyte
	freeMemory := float32(virtualMemory.Free) / Gigabyte
	return []MemoryUsage{{TotalMemory: totalMemory, FreeMemory: freeMemory, UsedMemoryPercent: float64(virtualMemory.UsedPercent)}}, nil
}

var previousNetIO []net.IOCountersStat

func ReportNetworkUsage() ([]NetworkUsage, error) {
	netIO, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	if previousNetIO == nil {
		previousNetIO = netIO
		return nil, nil
	}
	var networkUsage []NetworkUsage
	for i, io := range netIO {
		if i < len(previousNetIO) {
			deltaRecv := io.BytesRecv - previousNetIO[i].BytesRecv
			deltaSent := io.BytesSent - previousNetIO[i].BytesSent
			if deltaRecv > 0 && deltaSent > 0 {
				recvRate := float64(deltaRecv) / Megabyte
				sentRate := float64(deltaSent) / Megabyte
				networkUsage = append(networkUsage, NetworkUsage{Name: io.Name, ReceivedRate: recvRate, SentRate: sentRate})
			}
		}
	}
	previousNetIO = netIO
	return networkUsage, nil
}

func ReportDisk() ([]DiskUsage, error) {
	parts, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}
	var diskUsage []DiskUsage
	for _, part := range parts {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			return nil, err
		}
		totalGB := float64(usage.Total) / Gigabyte
		freeGB := float64(usage.Free) / Gigabyte
		diskUsage = append(diskUsage, DiskUsage{Mount: part.Mountpoint, TotalGB: totalGB, FreeGB: freeGB})
	}
	return diskUsage, nil
}

func clearScreen() {
	var cmd *exec.Cmd
	switch os := runtime.GOOS; os {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error clearing screen:", err)
	}
}
