package agent

import (
	"errors"
	"fmt"
)

type File struct {
	Src   string
	Dst   string
	Unzip bool
}

type Project struct {
	Name   string
	VCS    string
	User   string
	Token  string
	Branch string
	Filter string
	Files  []File
	Stop   string
	Start  string
}

func (p *Project) buildURL() (string, error) {
	if len(p.User) == 0 {
		return "", errors.New("user is empty")
	}
	if len(p.Name) == 0 {
		return "", errors.New("name is empty")
	}
	if len(p.Token) == 0 {
		return "", errors.New("token is empty")
	}
	u := fmt.Sprintf("https://circleci.com/api/v1.1/project/%s/%s/%s/latest/artifacts?circle-token=%s", p.VCS, p.User, p.Name, p.Token)
	if len(p.Branch) > 0 {
		u += fmt.Sprintf("&branch=%s", p.Branch)
	}
	if len(p.Filter) > 0 {
		u += fmt.Sprintf("&filter=%s", p.Filter)
	}
	return u, nil
}
