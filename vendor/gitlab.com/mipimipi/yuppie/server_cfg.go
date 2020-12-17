package yuppie

// Config represents the configuration of the UPnP server
type Config struct {
	// Interfaces contain the names of the network interfaces to be used. If
	// Interfaces is empty, all available interfaces will be used
	Interfaces []string
	// Port is the port where the server listens
	Port int
	// MaxAge is the validity time period of the SSDP advertisement in seconds
	MaxAge int
	// ProductName is the product name used for the server string
	ProductName string
	// ProductVersion is the product version for the server string
	ProductVersion string
	// StatusFile is the path to the JSON file that persists data such as
	// state variables
	StatusFile string
}

// defaultCfg is the default configuration which is used if the server is created
// with an empty configuration
var defaultCfg = Config{
	Interfaces:     nil,
	Port:           8008,
	MaxAge:         86400,
	ProductName:    "yuppie server",
	ProductVersion: "1.0",
	StatusFile:     "./status.json",
}

// equal returns true if two config structures are equal, otherwise false is returned
func (a Config) equal(b Config) bool {
	if len(a.Interfaces) != len(b.Interfaces) {
		return false
	}
	for i := 0; i < len(a.Interfaces); i++ {
		if a.Interfaces[i] != b.Interfaces[i] {
			return false
		}
	}

	return (a.Port == b.Port && a.MaxAge == b.MaxAge && a.ProductName == b.ProductName && a.ProductVersion == b.ProductVersion && a.StatusFile == b.StatusFile)
}
