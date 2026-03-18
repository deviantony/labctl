package types

const VERSION = "0.8.0"

// Droplet represents a DigitalOcean droplet managed by labctl.
type Droplet struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IPv4   string `json:"ipv4"`
	Region string `json:"region"`
	Size   string `json:"size"`
}
