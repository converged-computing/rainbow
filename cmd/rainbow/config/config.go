package match

import (
	"log"
	"os"

	"github.com/converged-computing/rainbow/pkg/config"
	"gopkg.in/yaml.v3"
)

// Run will init a new config
func RunInit(path string) error {

	// Generate an empty config - providing an empty filename ensures we don't read an existing one
	// This defaults to an in-memory vanilla database
	cfg, err := config.NewRainbowClientConfig("", "rainbow-cluster", "chocolate-cookied", "")
	if err != nil {
		return err
	}

	// Write to filename
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	log.Printf("Writing rainbow config to %s\n", path)
	err = os.WriteFile(path, out, 0644)
	if err != nil {
		return err
	}
	return nil
}