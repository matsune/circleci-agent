package agent

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type SpecFile struct {
	Projects []Project
}

// read ~/cispec.yml
func readSpecFile() (*SpecFile, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	ymlPath := filepath.Join(usr.HomeDir, ".cispec.yml")
	_, err = os.Stat(ymlPath)
	if os.IsNotExist(err) {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(ymlPath)
	if err != nil {
		return nil, err
	}
	var v SpecFile
	err = yaml.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
