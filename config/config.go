package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

// Config holds application configuration

type Config struct {

	// Prefix of cluster name
	ClusterNamePrefix string `json:"clusternameprefix"  required:"true" `

	// Directory where all cluster credentials and related files will be stored
	OCStorePath string `json:"ocstorepath" required:"true"`

	// Cluster pull secret to create an openshift cluster
	ClusterPullSecret string `json:"clusterpullsecret" required:"true"`

	// SSH key for the cluster
	SSHKey string `json:"sshkey" required:"true"`

	Platform string `json:"platform" required:"true"`
}

// NewConfig loads configuration values from environment variables
func NewConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("app", &config); err != nil {
		return nil, fmt.Errorf("error loading values from environment variables: %s",
			err.Error())
	}

	return &config, nil
}
