package services

// CastService: core dispatch — device-agnostic URL building + protocol routing.
// Device discovery is in cast_renderer.go.
// AirPlay implementation is in cast_airplay.go.
// DLNA implementation is in cast_dlna.go.

import (
"fmt"
"net/url"
"path"
"strings"
)

func mustParseURL(raw string) *url.URL {
u, err := url.Parse(raw)
if err != nil {
panic(fmt.Sprintf("mustParseURL(%q): %v", raw, err))
}
return u
}

// CastService handles device discovery and media casting (DLNA + AirPlay).
type CastService struct{}

func NewCastService() *CastService {
return &CastService{}
}

// BuildMediaURL constructs the HTTP URL that a renderer will use to fetch the
// media file. It picks the correct local IP (via OS routing table) so the URL
// is reachable from the renderer''s network segment, and encodes the path so
// that characters like ''&'', ''+'', ''='', '';'' don''t confuse device URL parsers.
func (s *CastService) BuildMediaURL(location, filePath string, port int) string {
localIP := GetBestLocalIP(location)

cleanPath := strings.TrimPrefix(filePath, "/")
segments := strings.Split(cleanPath, "/")
for i, seg := range segments {
encoded := url.PathEscape(seg)
encoded = strings.ReplaceAll(encoded, "&", "%26")
encoded = strings.ReplaceAll(encoded, "+", "%2B")
encoded = strings.ReplaceAll(encoded, "=", "%3D")
encoded = strings.ReplaceAll(encoded, ";", "%3B")
segments[i] = encoded
}
return fmt.Sprintf("http://%s:%d/raw/%s", localIP, port, strings.Join(segments, "/"))
}

// TitleFromPath extracts a human-readable title from a file path (last segment).
func TitleFromPath(filePath string) string {
return path.Base(filePath)
}

// CastTo sends a media URL to the device at location using the appropriate protocol.
func (s *CastService) CastTo(location, mediaURL, title string) error {
if strings.HasPrefix(location, "airplay://") {
return castAirPlay(location, mediaURL)
}
return castDLNA(location, mediaURL, title)
}
