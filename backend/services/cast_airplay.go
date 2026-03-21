package services

// AirPlay 1 media casting implementation.

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// newSessionID generates a random RFC-4122-like UUID for each AirPlay session.
// Using a fixed ID causes some devices to treat the request as a duplicate and ignore it.
func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// castAirPlay implements AirPlay 1 video playback.
// Protocol flow:
//  1. POST /reverse with Upgrade: PTTH/1.0  – opens the reverse event channel
//     (device uses this to send progress/status events back to sender)
//     Without this, strict AirPlay devices (e.g. JMGO, vv=2) report rate=1.0
//     but never actually start fetching the media URL.
//  2. POST /play with Content-Location header – starts URL-based playback
//  3. Poll GET /playback-info for diagnostic logging
func castAirPlay(location, mediaURL string) error {
	host := strings.TrimPrefix(location, "airplay://")
	sessionID := newSessionID()

	// 1. Establish reverse event channel (PTTH/1.0).
	//    We open a TCP connection, send the HTTP upgrade request, and keep the
	//    connection alive in a goroutine for up to 60 seconds so the device
	//    can use it to send back events while media is loading.
	reverseErr := make(chan error, 1)
	reverseReady := make(chan struct{})
	go func() {
		conn, err := net.DialTimeout("tcp", host, 5*time.Second)
		if err != nil {
			log.Printf("[AirPlay] reverse dial failed: %v", err)
			reverseErr <- err
			return
		}
		defer conn.Close()

		reverseReq := fmt.Sprintf(
			"POST /reverse HTTP/1.1\r\n"+
				"Host: %s\r\n"+
				"Upgrade: PTTH/1.0\r\n"+
				"Connection: Upgrade\r\n"+
				"X-Apple-Purpose: event\r\n"+
				"User-Agent: MediaControl/1.0\r\n"+
				"X-Apple-Session-ID: %s\r\n"+
				"Content-Length: 0\r\n"+
				"\r\n",
			host, sessionID)
		if _, err := conn.Write([]byte(reverseReq)); err != nil {
			log.Printf("[AirPlay] reverse write failed: %v", err)
			reverseErr <- err
			return
		}

		// Read the 101 Switching Protocols response.
		buf := make([]byte, 256)
		conn.SetReadDeadline(time.Now().Add(3 * time.Second)) //nolint:errcheck
		n, _ := conn.Read(buf)
		log.Printf("[AirPlay] reverse response: %q", string(buf[:n]))

		// Signal that the reverse channel is ready.
		close(reverseReady)

		// Keep the connection alive so the device can push events back.
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
		io.Copy(io.Discard, conn)                              //nolint:errcheck
	}()

	// Wait for the reverse channel to be established (or fail), with a short timeout.
	select {
	case err := <-reverseErr:
		log.Printf("[AirPlay] reverse channel failed (proceeding anyway): %v", err)
	case <-reverseReady:
		log.Printf("[AirPlay] reverse channel established, sessionID=%s", sessionID)
	case <-time.After(4 * time.Second):
		log.Printf("[AirPlay] reverse channel timeout (proceeding anyway)")
	}

	// 2. Send POST /play.
	httpClient := &http.Client{Timeout: 8 * time.Second}
	body := fmt.Sprintf("Content-Location: %s\r\nStart-Position: 0.000000\r\n", mediaURL)
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/play", host), bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("airplay build request: %w", err)
	}
	req.Header.Set("Content-Type", "text/parameters")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	req.Header.Set("User-Agent", "MediaControl/1.0")
	req.Header.Set("X-Apple-Session-ID", sessionID)

	log.Printf("[AirPlay] POST http://%s/play  session=%s  media=%s", host, sessionID, mediaURL)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("airplay POST /play: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	log.Printf("[AirPlay] /play response: %d  body: %q", resp.StatusCode, b)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("airplay /play returned %d: %s", resp.StatusCode, b)
	}

	// 3. Poll /playback-info after 3 seconds to diagnose device state.
	go func() {
		time.Sleep(3 * time.Second)
		infoReq, err := http.NewRequest("GET", fmt.Sprintf("http://%s/playback-info", host), nil)
		if err != nil {
			return
		}
		infoReq.Header.Set("X-Apple-Session-ID", sessionID)
		infoResp, err := httpClient.Do(infoReq)
		if err != nil {
			log.Printf("[AirPlay] /playback-info error: %v", err)
			return
		}
		defer infoResp.Body.Close()
		infoBody, _ := io.ReadAll(io.LimitReader(infoResp.Body, 1024))
		log.Printf("[AirPlay] /playback-info (3s): %d  body: %q", infoResp.StatusCode, infoBody)
	}()

	return nil
}
