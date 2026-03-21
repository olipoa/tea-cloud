package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/av1"
)

// browseMDNS 通过 mDNS/Bonjour 广播查找指定服务类型的设备（电视、投影仪等均走此协议）
func browseMDNS(label, svcType string, timeout time.Duration, results chan<- string) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Printf("[%s] resolver error: %v", label, err)
		return
	}
	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	go func() {
		if err := resolver.Browse(ctx, svcType, "local.", entries); err != nil {
			log.Printf("[%s] browse error: %v", label, err)
		}
	}()

	for entry := range entries {
		var ips []string
		for _, ip := range entry.AddrIPv4 {
			ips = append(ips, ip.String())
		}
		txt := strings.Join(entry.Text, " | ")
		line := fmt.Sprintf("[%s] %s  IPs=%v  Port=%d  Host=%s  txt=[%s]",
			label, entry.Instance, ips, entry.Port, entry.HostName, txt)
		results <- line
	}
}

// ssdpScan 用 UPnP/SSDP 广播（传统方式）
func ssdpScan(label, target string, timeout time.Duration, results chan<- string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	devs, err := goupnp.DiscoverDevicesCtx(ctx, target)
	if err != nil {
		log.Printf("[SSDP/%s] %v", label, err)
		return
	}
	seen := map[string]bool{}
	for _, d := range devs {
		key := d.Location.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		if d.Err != nil {
			continue
		}
		line := fmt.Sprintf("[SSDP/%s] %s  model=%s  loc=%s",
			label, d.Root.Device.FriendlyName, d.Root.Device.ModelName, d.Location)
		results <- line
	}
}

// probeAVTransport 用 goupnp av1 直接扫 AVTransport 服务
func probeAVTransport(timeout time.Duration, results chan<- string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	clients, _, err := av1.NewAVTransport1ClientsCtx(ctx)
	if err != nil {
		log.Printf("[AVTransport] %v", err)
		return
	}
	for _, c := range clients {
		results <- fmt.Sprintf("[AVTransport] %s @ %s", c.RootDevice.Device.FriendlyName, c.Location)
	}
}

// probeDescXML 对已知 IP（来自 ARP 表）直接 HTTP 探测常见 UPnP 描述路径
func probeDescXML(knownIPs []string, results chan<- string) {
	ports := []int{
		49152, 49153, 49154, 49155,
		7676, 8200, 52235, 1400, 55000, 60000, 8888,
		7000,  // AirPlay
		8008, 8009, // Google Cast
		5000, 5001, // 部分国产设备
	}
	paths := []string{
		"/description.xml", "/rootDesc.xml", "/upnp/desc.xml",
		"/upnpdevicedesc.xml", "/DeviceDescription.xml",
		"/dlna/device.xml", "/device_description.xml",
		"/", // AirPlay 根路径
	}
	httpClient := &http.Client{Timeout: 600 * time.Millisecond}
	var wg sync.WaitGroup
	for _, ip := range knownIPs {
		for _, port := range ports {
			for _, p := range paths {
				rawURL := fmt.Sprintf("http://%s:%d%s", ip, port, p)
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					resp, err := httpClient.Get(u)
					if err != nil {
						return
					}
					defer resp.Body.Close()
					if resp.StatusCode != 200 {
						return
					}
					ct := resp.Header.Get("Content-Type")
					body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
					bodyStr := string(body)
					isXML := strings.Contains(ct, "xml") ||
						strings.Contains(bodyStr, "urn:schemas-upnp-org") ||
						strings.Contains(bodyStr, "deviceType") ||
						strings.Contains(bodyStr, "<root") ||
						strings.Contains(bodyStr, "JMGO") ||
						strings.Contains(bodyStr, "AirPlay") ||
						strings.Contains(bodyStr, "Chromecast")
					// AirPlay 根路径返回特定 header
					if !isXML {
						if resp.Header.Get("X-Apple-Device-ID") != "" ||
							resp.Header.Get("CSeq") != "" {
							isXML = true
						}
					}
					if !isXML {
						return
					}
					end := 300
					if len(bodyStr) < end {
						end = len(bodyStr)
					}
					results <- fmt.Sprintf("[HTTP-PROBE] %s  ct=%s  body_start=[%s]",
						u, ct, strings.ReplaceAll(bodyStr[:end], "\n", " "))
				}(rawURL)
			}
		}
	}
	wg.Wait()
}

func main() {
	fmt.Println("=== Tea-Cloud 投屏设备发现工具 ===")
	fmt.Println("策略1: mDNS/Bonjour (AirPlay/Cast/DLNA)")
	fmt.Println("策略2: SSDP/UPnP 广播")
	fmt.Println("策略3: 对已知 ARP IP 直接 HTTP 探测")
	fmt.Println()

	results := make(chan string, 256)

	// 已知 LAN 设备 IP（可通过命令行参数传入，格式: diagcast 192.168.1.10 192.168.1.20）
	knownIPs := os.Args[1:]

	var wg sync.WaitGroup

	// mDNS 查找（Android TV / AirPlay / Google Cast / DLNA）
	mdnsServices := []struct{ label, svcType string }{
		{"AirPlay", "_airplay._tcp"},
		{"AirPlay-Audio(RAOP)", "_raop._tcp"},
		{"GoogleCast", "_googlecast._tcp"},
		{"DLNA-mDNS", "_dlna._udp"},
		{"UPnP-TCP", "_upnp._tcp"},
		{"HTTP", "_http._tcp"},
		{"MediaRenderer", "_media-renderer._tcp"},
	}
	for _, svc := range mdnsServices {
		wg.Add(1)
		go func(l, t string) {
			defer wg.Done()
			browseMDNS(l, t, 5*time.Second, results)
		}(svc.label, svc.svcType)
	}

	// mDNS hostname 直接解析
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, h := range []string{"JMGO_ThNv", "JMGO-ThNv", "jmgo", "JMGO"} {
			addrs, err := net.LookupHost(h + ".local")
			if err == nil {
				results <- fmt.Sprintf("[mDNS-Lookup] %s.local -> %v", h, addrs)
			}
		}
	}()

	// SSDP 广播
	wg.Add(1)
	go func() {
		defer wg.Done()
		ssdpScan("all", "ssdp:all", 5*time.Second, results)
		ssdpScan("Renderer", "urn:schemas-upnp-org:device:MediaRenderer:1", 5*time.Second, results)
	}()

	// AVTransport 客户端
	wg.Add(1)
	go func() {
		defer wg.Done()
		probeAVTransport(6*time.Second, results)
	}()

	// HTTP 直探已知 IP（含 AirPlay/Cast 特征头检测）
	wg.Add(1)
	go func() {
		defer wg.Done()
		probeDescXML(knownIPs, results)
	}()

	// 收集并打印结果
	go func() {
		wg.Wait()
		close(results)
	}()

	found := 0
	seenLines := map[string]bool{}
	for line := range results {
		if seenLines[line] {
			continue
		}
		seenLines[line] = true
		fmt.Println(line)
		found++
	}

	fmt.Printf("\n=== 完成，共发现 %d 个设备/服务 ===\n", found)
}
