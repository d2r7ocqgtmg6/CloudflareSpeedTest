package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Ptechgithub/CloudflareSpeedTest/task"
	"github.com/Ptechgithub/CloudflareSpeedTest/utils"
)

var (
	// Version information
	version = "2.2.5"

	// Command line flags
	ip4File  string
	ip6File  string
	outFile  string
	ipText   string

	pingCount    int
	maxDelay     int
	minDelay     int
	maxLossRate  float64
	disableDownload bool

	downloadTime    int
	testCount       int
	downloadSpeed   float64
	uploadSpeed     float64

	printVersion bool
)

func init() {
	// IP source flags
	flag.StringVar(&ip4File, "f", "ip.txt", "IPv4 data file path (supports CIDR format, e.g., 1.0.0.0/8)")
	flag.StringVar(&ip6File, "f6", "ipv6.txt", "IPv6 data file path (supports CIDR format)")
	flag.StringVar(&ipText, "ip", "", "Specify an IP range to test (comma separated, overrides file)")

	// Latency test flags
	flag.IntVar(&pingCount, "t", 4, "Latency test count for each IP")
	flag.IntVar(&maxDelay, "tl", 300, "Maximum average latency (ms), IPs above this are filtered")
	flag.IntVar(&minDelay, "tll", 0, "Minimum average latency (ms), IPs below this are filtered")
	flag.Float64Var(&maxLossRate, "tlr", 1.00, "Maximum packet loss rate (0.00~1.00), IPs above this are filtered")

	// Download test flags
	flag.BoolVar(&disableDownload, "dd", false, "Disable download speed test, output results sorted by latency")
	flag.IntVar(&downloadTime, "dt", 10, "Download test duration (seconds)")
	flag.IntVar(&testCount, "dn", 10, "Number of IPs to download test")
	flag.Float64Var(&downloadSpeed, "sl", 0, "Minimum download speed (MB/s), stop testing once reached")

	// Output flags
	flag.StringVar(&outFile, "o", "result.csv", "Output file path for results (empty string to disable)")
	flag.IntVar(&utils.PrintNum, "p", 10, "Number of results to display")

	// Misc flags
	flag.BoolVar(&printVersion, "v", false, "Print version and exit")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CloudflareSpeedTest v%s\n", version)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if printVersion {
		fmt.Printf("CloudflareSpeedTest v%s\n", version)
		os.Exit(0)
	}

	// Print banner
	fmt.Printf("CloudflareSpeedTest v%s\n\n", version)

	// Build configuration from flags
	cfg := &task.Config{
		IPv4File:        ip4File,
		IPv6File:        ip6File,
		IPText:          ipText,
		OutFile:         outFile,
		PingCount:       pingCount,
		MaxDelay:        time.Duration(maxDelay) * time.Millisecond,
		MinDelay:        time.Duration(minDelay) * time.Millisecond,
		MaxLossRate:     float32(maxLossRate),
		DisableDownload: disableDownload,
		DownloadTime:    time.Duration(downloadTime) * time.Second,
		TestCount:       testCount,
		MinSpeed:        float32(downloadSpeed),
	}

	// Run the speed test
	if err := task.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
