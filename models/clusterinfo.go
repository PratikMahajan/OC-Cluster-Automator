package models

type Cluster struct {
	Name     string `json:"name" validate:"required"`
	Dir      string `json:"dir" validate:"required"`
	Platform string `json:"platform" validate:"required"`
}

type ClusterStore struct {
	Clusters map[string][]Cluster `json:"clusters" validate:"required"`
}
