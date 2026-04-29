package task

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// PingData holds the result of a single ping test to a Cloudflare IP.
type PingData struct {
	IP       *net.IPAddr
	Sent     int
	Received int
	Delay    time.Duration
}

// PingDelaySet is a sortable slice of PingData.
type PingDelaySet []*PingData

func (s PingDelaySet) Len() int      { return len(s) }
func (s PingDelaySet) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s PingDelaySet) Less(i, j int) bool {
	// Sort by loss rate first, then by delay
	iLoss := float32(s[i].Sent-s[i].Received) / float32(s[i].Sent)
	jLoss := float32(s[j].Sent-s[j].Received) / float32(s[j].Sent)
	if iLoss != jLoss {
		return iLoss < jLoss
	}
	return s[i].Delay < s[j].Delay
}

// PingConfig holds configuration for the TCP ping test.
type PingConfig struct {
	Port        int
	PingTimes   int
	MaxRoutines int
	Timeout     time.Duration
}

// DefaultPingConfig returns a PingConfig with sensible defaults.
func DefaultPingConfig() *PingConfig {
	return &PingConfig{
		Port:        443,
		PingTimes:   4,
		MaxRoutines: 1000,
		Timeout:     time.Second,
	}
}

// tcpPing performs a single TCP connection attempt and returns the round-trip duration.
func tcpPing(ip *net.IPAddr, port int, timeout time.Duration) (time.Duration, error) {
	addr := fmt.Sprintf("%s:%d", ip.String(), port)
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return 0, err
	}
	delay := time.Since(start)
	_ = conn.Close()
	return delay, nil
}

// pingIP runs multiple TCP pings against a single IP and returns aggregated PingData.
func pingIP(ip *net.IPAddr, cfg *PingConfig) *PingData {
	data := &PingData{
		IP:   ip,
		Sent: cfg.PingTimes,
	}
	var totalDelay time.Duration
	for i := 0; i < cfg.PingTimes; i++ {
		delay, err := tcpPing(ip, cfg.Port, cfg.Timeout)
		if err != nil {
			continue
		}
		data.Received++
		totalDelay += delay
	}
	if data.Received > 0 {
		data.Delay = totalDelay / time.Duration(data.Received)
	}
	return data
}

// BatchPing concurrently pings a list of IPs and returns sorted results.
func BatchPing(ips []*net.IPAddr, cfg *PingConfig) PingDelaySet {
	if cfg == nil {
		cfg = DefaultPingConfig()
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results PingDelaySet
		sem     = make(chan struct{}, cfg.MaxRoutines)
	)

	for _, ip := range ips {
		wg.Add(1)
		sem <- struct{}{}
		go func(addr *net.IPAddr) {
			defer wg.Done()
			defer func() { <-sem }()
			data := pingIP(addr, cfg)
			// Only include IPs that responded at least once
			if data.Received > 0 {
				mu.Lock()
				results = append(results, data)
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()
	sort.Sort(results)
	return results
}
