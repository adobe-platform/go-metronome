package metronome

// A Container defines a chronos container
type Container struct {
	Type    string              `json:"type,omitempty"`
	Image   string              `json:"image,omitempty"`
	Network string              `json:"network,omitempty"`
	Volumes []map[string]string `json:"volumes,omitempty"`
}
