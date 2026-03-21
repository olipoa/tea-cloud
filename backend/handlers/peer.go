package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"tea-cloud/backend/services"
)

type PeerHandler struct {
	discovery *services.DiscoveryService
}

func NewPeerHandler(discovery *services.DiscoveryService) *PeerHandler {
	return &PeerHandler{discovery: discovery}
}

// ListPeers godoc
// GET /api/peers
// Scans the LAN for other tea-cloud nodes and returns them.
func (h *PeerHandler) ListPeers(c *gin.Context) {
	peers, err := h.discovery.Discover()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if peers == nil {
		peers = []services.PeerInfo{}
	}
	c.JSON(http.StatusOK, peers)
}

// SelfInfo godoc
// GET /api/self
// Returns information about this node.
func SelfInfo(nodeName string, port int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ips := services.GetLocalIPs()
		c.JSON(http.StatusOK, gin.H{
			"name": nodeName,
			"port": port,
			"ips":  ips,
		})
	}
}
