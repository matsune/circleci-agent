package agent

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

func Run(target string) error {
	s, err := readSpecFile()
	if err != nil {
		return err
	}

	var p *Project
	for _, project := range s.Projects {
		if target == project.Name {
			p = &project
		}
	}
	if p == nil {
		return fmt.Errorf("could not find target %q in spec file.", target)
	}

	u, err := p.buildURL()
	if err != nil {
		return err
	}

	arts, err := getArtifacts(u)
	if err != nil {
		return err
	}

	d := []dest{}
	for _, art := range arts {
		var dst string
		for _, f := range p.Files {
			if art.Path == f.Src {
				dst = f.Dst
				break
			}
		}
		if len(dst) == 0 {
			log.Printf("could not find destination for %s\n", art.Path)
		} else {
			d = append(d, dest{
				Artifact: art,
				Dst:      dst,
			})
		}
	}

	if err = resolves(d); err != nil {
		return err
	}

	var out string
	log.Println("Exec stop...")
	if out, err = execute(p.Stop); err != nil {
		return err
	}
	log.Print(out)

	log.Println("Exec start...")
	if out, err = execute(p.Start); err != nil {
		return err
	}
	log.Print(out)

	return nil
}

func execute(str string) (string, error) {
	args := strings.Split(str, " ")
	cmd := exec.Command(args[0], args[1:]...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return stdout.String(), nil
}

type dest struct {
	Artifact
	Dst string
}

func resolves(dests []dest) error {
	eg := errgroup.Group{}
	for _, d := range dests {
		d := d
		eg.Go(func() error {
			return resolve(d)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func resolve(d dest) error {
	return download(d.Artifact.URL, d.Dst)
}

func download(url, path string) error {
	log.Printf("download %s to %s", url, path)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
