package services

// DLNA/UPnP AVTransport media casting implementation.

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/huin/goupnp/dcps/av1"
)

// castDLNA sends Stop + SetAVTransportURI + Play over UPnP SOAP.
// Stop is required by most DLNA renderers before a new URI can be accepted;
// skipping it causes SetAVTransportURI to be silently ignored on many devices.
func castDLNA(location, mediaURL, title string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	log.Printf("[DLNA] connecting to renderer: %s", location)
	clients, err := av1.NewAVTransport1ClientsByURLCtx(ctx, mustParseURL(location))
	if err != nil {
		return fmt.Errorf("connect to renderer: %w", err)
	}
	if len(clients) == 0 {
		return fmt.Errorf("no AVTransport service at %s", location)
	}
	client := clients[0]

	// 1. Stop current playback so the renderer accepts a new URI.
	//    Most commercial apps (Youku, iQiyi, etc.) do this step; skipping it
	//    causes many smart TVs (TCL, Hisense, Mi TV…) to silently ignore the next SetAVTransportURI.
	if stopErr := client.StopCtx(ctx, 0); stopErr != nil {
		log.Printf("[DLNA] Stop (pre-cast) failed (non-fatal): %v", stopErr)
	} else {
		log.Printf("[DLNA] Stop OK")
	}

	// 2. Set new media URI.
	metadata := buildDIDLMetadata(title, mediaURL)
	log.Printf("[DLNA] SetAVTransportURI  media=%s", mediaURL)
	if err := client.SetAVTransportURICtx(ctx, 0, mediaURL, metadata); err != nil {
		return fmt.Errorf("SetAVTransportURI: %w", err)
	}

	// 3. Start playback.
	log.Printf("[DLNA] Play")
	if err := client.PlayCtx(ctx, 0, "1"); err != nil {
		return fmt.Errorf("Play: %w", err)
	}
	log.Printf("[DLNA] Play command sent successfully")
	return nil
}

// mimeClassForURL returns the DLNA upnp:class and MIME type inferred from the
// file extension in the URL. Used to build accurate DIDL-Lite metadata.
func mimeClassForURL(rawURL string) (upnpClass, mimeType string) {
	p := strings.ToLower(rawURL)
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	switch {
	case strings.HasSuffix(p, ".mp4") || strings.HasSuffix(p, ".m4v"):
		return "object.item.videoItem", "video/mp4"
	case strings.HasSuffix(p, ".mkv"):
		return "object.item.videoItem", "video/x-matroska"
	case strings.HasSuffix(p, ".avi"):
		return "object.item.videoItem", "video/x-msvideo"
	case strings.HasSuffix(p, ".mov"):
		return "object.item.videoItem", "video/quicktime"
	case strings.HasSuffix(p, ".wmv"):
		return "object.item.videoItem", "video/x-ms-wmv"
	case strings.HasSuffix(p, ".ts") || strings.HasSuffix(p, ".m2ts"):
		return "object.item.videoItem", "video/MP2T"
	case strings.HasSuffix(p, ".flv"):
		return "object.item.videoItem", "video/x-flv"
	case strings.HasSuffix(p, ".webm"):
		return "object.item.videoItem", "video/webm"
	case strings.HasSuffix(p, ".mp3"):
		return "object.item.audioItem.musicTrack", "audio/mpeg"
	case strings.HasSuffix(p, ".flac"):
		return "object.item.audioItem.musicTrack", "audio/flac"
	case strings.HasSuffix(p, ".ogg") || strings.HasSuffix(p, ".oga"):
		return "object.item.audioItem.musicTrack", "audio/ogg"
	case strings.HasSuffix(p, ".m4a") || strings.HasSuffix(p, ".aac"):
		return "object.item.audioItem.musicTrack", "audio/mp4"
	case strings.HasSuffix(p, ".wav"):
		return "object.item.audioItem.musicTrack", "audio/wav"
	case strings.HasSuffix(p, ".jpg") || strings.HasSuffix(p, ".jpeg"):
		return "object.item.imageItem.photo", "image/jpeg"
	case strings.HasSuffix(p, ".png"):
		return "object.item.imageItem.photo", "image/png"
	case strings.HasSuffix(p, ".gif"):
		return "object.item.imageItem.photo", "image/gif"
	case strings.HasSuffix(p, ".webp"):
		return "object.item.imageItem.photo", "image/webp"
	default:
		return "object.item", "*"
	}
}

// buildDIDLMetadata constructs a DIDL-Lite XML metadata string with accurate
// upnp:class and protocolInfo so strict DLNA renderers (Samsung, LG TVs, etc.)
// accept and play the media.
func buildDIDLMetadata(title, mediaURL string) string {
	if title == "" {
		return ""
	}
	upnpClass, mime := mimeClassForURL(mediaURL)
	protocolInfo := "http-get:*:" + mime + ":*"
	return fmt.Sprintf(
		`<DIDL-Lite xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" `+
			`xmlns:dc="http://purl.org/dc/elements/1.1/" `+
			`xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/">`+
			`<item id="0" parentID="-1" restricted="1">`+
			`<dc:title>%s</dc:title>`+
			`<upnp:class>%s</upnp:class>`+
			`<res protocolInfo="%s">%s</res>`+
			`</item></DIDL-Lite>`,
		xmlEscape(title), upnpClass, protocolInfo, xmlEscape(mediaURL),
	)
}

// xmlEscape replaces XML special characters to prevent malformed metadata.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
