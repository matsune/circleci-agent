package agent

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
	log.Println(arts)

	d := []dest{}
	for _, art := range arts {
		var file *File
		for _, f := range p.Files {
			if art.Path == f.Src {
				file = &f
				break
			}
		}
		if file == nil {
			log.Printf("could not find destination for %s\n", art.Path)
		} else {
			d = append(d, dest{
				Artifact: art,
				File:     *file,
			})
		}
	}

	if err = resolves(d, p.Token); err != nil {
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
	File
}

func resolves(dests []dest, token string) error {
	eg := errgroup.Group{}
	for _, d := range dests {
		d := d
		eg.Go(func() error {
			return resolve(d, token)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func resolve(d dest, token string) error {
	url := d.URL + "?circle-token=" + token
	if d.Unzip {
		_, err := unzip(url, d.Dst)
		if err != nil {
			return err
		}
		return nil
	} else {
		log.Printf("download %s to %s", url, d.Dst)
		return download(url, d.Dst, d.Chmod)
	}
}

func download(url, path string, chmod os.FileMode) error {
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

	if chmod > 0 {
		if err = os.Chmod(path, chmod); err != nil {
			return err
		}
	}
	return err
}

func unzip(url string, dest string) ([]string, error) {
	var filenames []string

	r, err := readZipURL(url)
	if err != nil {
		return nil, err
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}
			_, err = io.Copy(outFile, rc)
			outFile.Close()
			if err != nil {
				return filenames, err
			}
		}
	}
	return filenames, nil
}

func readZip(res *http.Response) (*zip.Reader, error) {
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body with error %s", err)
	}
	r := bytes.NewReader(b)
	return zip.NewReader(r, int64(r.Len()))
}

func readZipURL(url string) (*zip.Reader, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf(
			"Fetching zip URL %s failed with error %s.", url, err)
	}
	defer res.Body.Close()
	return readZip(res)
}
