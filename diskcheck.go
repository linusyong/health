package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers/disk"
	"github.com/InVisionApp/go-health/handlers"

	"gopkg.in/yaml.v2"
)

func main() {
	// Read config.yaml
	type Conf struct {
		Disk []struct {
			Path    string  `yaml:"path"`
			Warning float64 `yaml:"warning"`
			Critial float64 `yaml:"critical"`
		}
	}

	confYaml, _ := ioutil.ReadFile("config.yaml")

	var conf Conf
	yaml.Unmarshal(confYaml, &conf)

	// Create a new health instance
	h := health.New()

	diskCheck := make([]*diskchk.DiskUsage, len(conf.Disk))

	for i, d := range conf.Disk {
		// Create a couple of checks
		diskCheck[i], _ = diskchk.NewDiskUsage(&diskchk.DiskUsageConfig{
			Path:              d.Path,
			WarningThreshold:  d.Warning,
			CriticalThreshold: d.Critial,
		})
		// Add the checks to the health instance
		h.AddChecks([]*health.Config{
			{
				Name:     "disk-check-" + d.Path,
				Checker:  diskCheck[i],
				Interval: time.Duration(10) * time.Second,
				Fatal:    true,
			},
		})
	}

	//  Start the healthcheck process
	if err := h.Start(); err != nil {
		log.Fatalf("Unable to start healthcheck: %v", err)
	}

	log.Println("Server listening on :8080")

	// Define a healthcheck endpoint and use the built-in JSON handler
	http.HandleFunc("/healthcheck", handlers.NewBasicHandlerFunc(h))
	http.HandleFunc("/healthreport", handlers.NewJSONHandlerFunc(h, nil))
	http.ListenAndServe(":8080", nil)
}
