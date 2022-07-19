package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/katallaxie/run/pkg/spec"

	"gopkg.in/yaml.v3"
)

type archiveProvider struct {
	opts *ProviderOpts
}

// NewArchive ...
func NewArchive(opts ...ProviderOpt) Provider {
	options := new(ProviderOpts)
	options.Configure(opts...)

	p := new(archiveProvider)
	p.opts = options

	return p
}

// CloneWithContext ...
func (a *archiveProvider) CloneWithContext(ctx context.Context, url, folder string) error {
	path, err := filepath.Abs(folder)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// reading the zip from the body
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	// this gives the files in the zip in a slice
	var assets []*zip.File
	var y *zip.File

	// stream all the files ... maybe copy to the filesystem
	for _, f := range zipReader.File {
		if !strings.Contains(f.Name, ".sc.yaml") {
			assets = append(assets, f)

			continue
		}

		y = f
	}

	if y == nil {
		return errors.New(`no sc spec found`)
	}

	for _, f := range assets {
		parts := strings.Split(f.Name, string(os.PathSeparator))

		// Store filename/path for returning and using later on
		fpath := filepath.Join(path, filepath.Join(parts[1:]...))

		if f.FileInfo().IsDir() {
			// Make Folder
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	// this should be later filtered to be the root of the files
	// base := path.Base(y.Name)

	var s *spec.Spec
	raw, err := readZipFile(y)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(raw, &s); err != nil {
		return err
	}

	return nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}
