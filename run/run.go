package run

import (
	"fmt"
	"github.com/Portshift/klar/clair"
	"github.com/Portshift/klar/config"
	"github.com/Portshift/klar/docker"
	log "github.com/sirupsen/logrus"
)

type ScanResults struct {
	Vulnerabilities    []*clair.Vulnerability
	FsLayerCommands    []*docker.FsLayerCommand
	ImageHash          string
	ImageSchemaVersion int
}

func ExecuteScan(conf *config.Config) (*ScanResults, error) {
	image, err := docker.NewImage(&conf.DockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse name: %v", err)
	}

	err = image.Pull()
	if err != nil {
		return nil, fmt.Errorf("failed to pull image: %w", err)
	}

	if err := image.FetchFsCommands(&conf.DockerConfig); err != nil {
		return nil, fmt.Errorf("failed to fetch layer commands: %v", err)
	}

	if len(image.FsLayers) == 0 {
		return nil, fmt.Errorf("failed to pull pull fsLayers")
	}

	commands := image.GetFsCommands()

	log.Infof("Analysing %d layers", len(image.FsLayers))

	var vulnerabilities []*clair.Vulnerability

	c := clair.NewClair(conf.ClairAddr, conf.ClairTimeout)
	vulnerabilities, err = c.Analyse(image)
	if err != nil {
		log.Errorf("Failed to analyze using API: %s", err)
	} else {
		if !conf.JSONOutput {
			log.Infof("Got results from Clair API")
		}
	}

	return &ScanResults{
		Vulnerabilities:    vulnerabilities,
		FsLayerCommands:    commands,
		ImageHash:          docker.TrimDigest(image.Digest),
		ImageSchemaVersion: image.SchemaVersion,
	}, err
}
