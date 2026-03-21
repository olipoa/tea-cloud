package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

// PeerInfo represents a discovered tea-cloud peer on the LAN.
type PeerInfo struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	AddrV4   string `json:"addrV4"`
	URL      string `json:"url"`
}

const serviceType = "_tea-cloud._tcp"
const serviceDomain = "local."

// DiscoveryService handles both advertising this node and discovering peers.
type DiscoveryService struct {
	NodeName string
	Port     int
	server   *zeroconf.Server
}

func NewDiscoveryService(nodeName string, port int) *DiscoveryService {
	return &DiscoveryService{NodeName: nodeName, Port: port}
}

// Register starts advertising this node via mDNS.
func (d *DiscoveryService) Register() error {
	server, err := zeroconf.Register(d.NodeName, serviceType, serviceDomain, d.Port, []string{"version=1"}, nil)
	if err != nil {
		return fmt.Errorf("mDNS register: %w", err)
	}
	d.server = server
	log.Printf("[mDNS] Registered as %q on port %d", d.NodeName, d.Port)
	return nil
}

// Stop shuts down the mDNS advertisement.
func (d *DiscoveryService) Stop() {
	if d.server != nil {
		d.server.Shutdown()
	}
}

// Discover scans the LAN for other tea-cloud peers and returns them.
// Waits up to 2 seconds for responses.
func (d *DiscoveryService) Discover() ([]PeerInfo, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := resolver.Browse(ctx, serviceType, serviceDomain, entries); err != nil {
		return nil, fmt.Errorf("mDNS browse: %w", err)
	}

	var peers []PeerInfo
	done := ctx.Done()
	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				return peers, nil
			}
			if entry == nil {
				continue
			}
			peer := PeerInfo{
				Name: entry.ServiceRecord.Instance,
				Host: entry.HostName,
				Port: entry.Port,
			}
			if len(entry.AddrIPv4) > 0 {
				peer.AddrV4 = entry.AddrIPv4[0].String()
				peer.URL = fmt.Sprintf("http://%s:%d", peer.AddrV4, peer.Port)
			} else {
				peer.URL = fmt.Sprintf("http://%s:%d", entry.HostName, peer.Port)
			}
			peers = append(peers, peer)
		case <-done:
			return peers, nil
		}
	}
}

// GetBestLocalIP returns the local IPv4 address that the OS would use to reach the
// host in location. It performs a UDP "dial" (no packets sent) so the kernel
// routing table picks the correct outgoing interface automatically.
// Falls back to GetLocalIPs()[0] if the target cannot be resolved.
func GetBestLocalIP(location string) string {
	// Normalise: "airplay://host:port" → "http://host:port" so url.Parse works.
	normalized := strings.ReplaceAll(location, "airplay://", "http://")
	if u, err := url.Parse(normalized); err == nil && u.Hostname() != "" {
		target := u.Hostname() + ":1"
		conn, err := net.Dial("udp", target)
		if err == nil {
			defer conn.Close()
			if udpAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
				return udpAddr.IP.String()
			}
		}
	}
	// Fallback: first non-loopback IPv4 (original behaviour)
	if ips := GetLocalIPs(); len(ips) > 0 {
		return ips[0]
	}
	return "127.0.0.1"
}

// GetLocalIPs returns non-loopback IPv4 addresses of this machine.
func GetLocalIPs() []string {
	var result []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return result
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				result = append(result, ip4.String())
			}
		}
	}
	return result
}
