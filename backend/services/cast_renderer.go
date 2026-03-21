package services

// Device discovery: AirPlay (mDNS) + DLNA (SSDP) + direct port probing.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/huin/goupnp/dcps/av1"
)

// DLNARenderer represents a discovered cast device on the LAN.
// Protocol is "dlna" or "airplay".
type DLNARenderer struct {
	Name     string `json:"name"`
	Location string `json:"location"` // DLNA: device description URL; AirPlay: "airplay://ip:port"
	Protocol string `json:"protocol"` // "dlna" | "airplay"
}

// DiscoverRenderers scans the LAN for cast devices:
//   - AirPlay devices via mDNS (_airplay._tcp), fast (~4s)
//   - DLNA MediaRenderer devices via SSDP/UPnP
//   - Fallback: direct port probing on IPs found via AirPlay (catches devices
//     like JMGO that don't respond to SSDP multicast but run DLNA on a fixed port)
func (s *CastService) DiscoverRenderers() ([]DLNARenderer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	seen := make(map[string]bool)
	var renderers []DLNARenderer
	var airplayIPs []string

	// --- AirPlay via mDNS (faster, works with JMGO/Android projectors) ---
	resolver, err := zeroconf.NewResolver(nil)
	if err == nil {
		entries := make(chan *zeroconf.ServiceEntry)
		go resolver.Browse(ctx, "_airplay._tcp", "local.", entries) //nolint:errcheck
		for entry := range entries {
			if len(entry.AddrIPv4) == 0 {
				continue
			}
			ip := entry.AddrIPv4[0].String()
			loc := fmt.Sprintf("airplay://%s:%d", ip, entry.Port)
			if seen[loc] {
				continue
			}
			seen[loc] = true
			// Collect unique IPs for DLNA fallback probing below.
			ipSeen := false
			for _, knownIP := range airplayIPs {
				if knownIP == ip {
					ipSeen = true
					break
				}
			}
			if !ipSeen {
				airplayIPs = append(airplayIPs, ip)
			}
			name := entry.Instance
			if name == "" {
				name = entry.HostName
			}
			renderers = append(renderers, DLNARenderer{
				Name:     name,
				Location: loc,
				Protocol: "airplay",
			})
		}
	}

	// --- DLNA/UPnP via SSDP (traditional smart TVs) ---
	clients, _, _ := av1.NewAVTransport1ClientsCtx(ctx)
	for _, c := range clients {
		loc := c.Location.String()
		if seen[loc] {
			continue
		}
		seen[loc] = true
		name := c.RootDevice.Device.FriendlyName
		if name == "" {
			name = c.RootDevice.Device.ModelName
		}
		if name == "" {
			name = loc
		}
		renderers = append(renderers, DLNARenderer{Name: name, Location: loc, Protocol: "dlna"})
	}

	// --- DLNA fallback: probe known ports on IPs discovered via AirPlay ---
	// Some devices (e.g. JMGO projectors) run a DLNA MediaRenderer on a fixed
	// port but do NOT respond to SSDP multicast, so SSDP/UPnP discovery misses
	// them. By probing their IPs directly we reliably surface the DLNA renderer.
	dlnaRenderers := probeDLNAOnIPs(airplayIPs, seen)
	renderers = append(renderers, dlnaRenderers...)

	return renderers, nil
}

// probeDLNAOnIPs concurrently probes common DLNA MediaRenderer ports on each
// IP address. Returns any renderers found that haven't already been seen.
func probeDLNAOnIPs(ips []string, seen map[string]bool) []DLNARenderer {
	// Common ports used by DLNA MediaRenderers, especially on Chinese devices.
	probePorts := []int{49152, 49153, 49154, 49155, 8200, 52235, 7676, 55000, 60000, 8888, 5000}
	probePaths := []string{"/description.xml", "/rootDesc.xml", "/upnp/desc.xml", "/DeviceDescription.xml"}

	type result struct {
		r DLNARenderer
	}
	ch := make(chan result, len(ips)*len(probePorts))
	var wg sync.WaitGroup

	client := &http.Client{Timeout: 800 * time.Millisecond}

	for _, ip := range ips {
		for _, port := range probePorts {
			wg.Add(1)
			go func(ip string, port int) {
				defer wg.Done()
				for _, path := range probePaths {
					loc := fmt.Sprintf("http://%s:%d%s", ip, port, path)
					resp, err := client.Get(loc)
					if err != nil {
						continue
					}
					body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
					resp.Body.Close()
					if resp.StatusCode != 200 {
						continue
					}
					bodyStr := string(body)
					// Must advertise as a MediaRenderer (not a MediaServer).
					if !strings.Contains(bodyStr, "MediaRenderer") {
						continue
					}
					// Extract friendlyName.
					name := ""
					if i := strings.Index(bodyStr, "<friendlyName>"); i >= 0 {
						s := bodyStr[i+len("<friendlyName>"):]
						if j := strings.Index(s, "</friendlyName>"); j >= 0 {
							name = s[:j]
						}
					}
					if name == "" {
						name = fmt.Sprintf("DLNA @ %s:%d", ip, port)
					}
					ch <- result{DLNARenderer{Name: name, Location: loc, Protocol: "dlna"}}
					return // Found on this port; stop trying other paths/ports for this ip:port.
				}
			}(ip, port)
		}
	}

	wg.Wait()
	close(ch)

	var found []DLNARenderer
	for r := range ch {
		loc := r.r.Location
		if seen[loc] {
			continue
		}
		seen[loc] = true
		found = append(found, r.r)
	}
	return found
}
