package types

const VERSION = "0.1.0"

// VPS represents a VPS instance.
type VPS struct {
	ID     int
	Name   string
	Region string
	Size   string
	Ipv4   string
}
