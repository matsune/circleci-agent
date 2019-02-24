package agent

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Artifact struct {
	Path       string `json:"path"`
	PrettyPath string `json:"pretty_path"`
	NodeIndex  int    `json:"node_index"`
	URL        string `json:"url"`
}

func getArtifacts(url string) ([]Artifact, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return nil, err
	}
	c := new(http.Client)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var v []Artifact
	if err = json.Unmarshal(bytes, &v); err != nil {
		return nil, err
	}
	return v, nil
}
