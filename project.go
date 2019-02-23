package main

import (
	"errors"
	"fmt"
)

type File struct {
	Src string
	Dst string
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

func (p *Project) buildURL(token string) (string, error) {
	if len(p.User) == 0 {
		return "", errors.New("user is empty")
	}
	if len(p.Name) == 0 {
		return "", errors.New("name is empty")
	}
	if len(token) == 0 {
		return "", errors.New("Token is empty")
	}
	u := fmt.Sprintf("https://circleci.com/api/v1.1/project/%s/%s/%s/latest/artifacts?circle-token=%s", p.VCS, p.User, p.Name, token)
	if len(p.Branch) > 0 {
		u += fmt.Sprintf("&branch=%s", p.Branch)
	}
	if len(p.Filter) > 0 {
		u += fmt.Sprintf("&filter=%s", p.Filter)
	}
	return u, nil
}
