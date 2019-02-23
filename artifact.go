package main

type Artifact struct {
	Path       string `json:"path"`
	PrettyPath string `json:"pretty_path"`
	NodeIndex  int    `json:"node_index"`
	URL        string `json:"url"`
}
