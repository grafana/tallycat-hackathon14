package schema

import (
	"fmt"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
)

type Producer struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace,omitempty"`
	Version    string    `json:"version,omitempty"`
	InstanceID string    `json:"instanceId,omitempty"`
	FirstSeen  time.Time `json:"firstSeen"`
	LastSeen   time.Time `json:"lastSeen"`
}

// ProducerID generates a unique ID for a producer using xxhash
func (p *Producer) ProducerID() string {
	parts := []string{
		p.Name,
		p.Namespace,
		p.Version,
		p.InstanceID,
	}

	// Filter out empty parts
	var nonEmptyParts []string
	for _, part := range parts {
		if part != "" {
			nonEmptyParts = append(nonEmptyParts, part)
		}
	}

	h := xxhash.New()
	h.Write([]byte(strings.Join(nonEmptyParts, "|")))
	return fmt.Sprintf("%x", h.Sum64())
}
