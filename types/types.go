package types

const VERSION = "0.3.0"

// A flask is an environment that can run in LXC or in the cloud
type Flask struct {
	Name string
	Ipv4 string
	LXD  FlaskLXDProperties
	DO   FlaskDOProperties
}

// FlaskConfig holds the configuration for creating a flask
type FlaskConfig struct {
	// Region is the region where the flask will be created
	// Only valid for the DigitalOcean provider
	Region string
	// Size is the size of the flask
	// Only valid for the DigitalOcean provider
	Size string
	// Profile is the LXD profile to use for the flask
	// Only valid for the LXD provider
	Profile string
	// Image is the image to use for the flask
	// Only valid for the LXD provider
	Image string
}

// FlaskDOProperties holds the DigitalOcean specific properties for a flask
type FlaskDOProperties struct {
	ID     int
	Region string
	Size   string
}

// FlaskLXDProperties holds the LXD specific properties for a flask
type FlaskLXDProperties struct {
	ID       string
	Status   string
	Profiles []string
}

// FlaskManager is the interface that wraps the basic flask management methods.
type FlaskManager interface {
	CreateFlask(name string, cfg FlaskConfig) (Flask, error)
	GetFlask(id string) (Flask, error)
	ListFlasks() ([]Flask, error)
	RemoveFlask(flask Flask) error
	WaitUntilFlaskIsReady(flask *Flask) error
}
