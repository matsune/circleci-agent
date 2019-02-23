package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not passed project name")
		os.Exit(1)
	}

	token := os.Getenv("CI_TOKEN")
	if len(token) == 0 {
		panic("$CI_TOKEN is not set")
	}

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	ymlPath := filepath.Join(usr.HomeDir, ".ci-arts.yml")
	_, err = os.Stat(ymlPath)
	if os.IsNotExist(err) {
		panic("~/.ci-arts.yml not found")
	}

	bytes, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		panic(err)
	}
	var v struct {
		Projects []Project
	}
	err = yaml.Unmarshal(bytes, &v)
	if err != nil {
		panic(err)
	}

	var p *Project
	for _, project := range v.Projects {
		if os.Args[1] == project.Name {
			p = &project
		}
	}
	if p == nil {
		panic("could not find project")
	}

	u, err := p.buildURL(token)
	if err != nil {
		panic(err)
	}

	arts, err := getArtifacts(u)
	if err != nil {
		panic(err)
	}

	for _, art := range arts {
		var dst string
		for _, f := range p.Files {
			if art.Path == f.Src {
				dst = f.Dst
				break
			}
		}
		if len(dst) == 0 {
			fmt.Printf("%s setting could not find\n", art.Path)
		} else {
			fmt.Println("cp", art.Path, dst)
		}
	}
}
