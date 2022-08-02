package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andersnormal/pkg/utils/files"
	gg "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

var (
	// ErrFolderNotEmpty signals that the destination folder not empty
	ErrFolderNotEmpty = errors.New("destination folder not empty")
)

type git struct {
	opts *ProviderOpts
}

// NewGit ...
func NewGit(opts ...ProviderOpt) Provider {
	options := new(ProviderOpts)

	g := new(git)
	g.opts = options
	g.opts.Configure(opts...)

	return g
}

// CloneWithContext ...
func (g *git) CloneWithContext(ctx context.Context, url string, folder string) error {
	path, err := filepath.Abs(folder)
	if err != nil {
		return err
	}

	empty, err := files.IsDirEmpty(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if !os.IsNotExist(err) && !empty {
		return ErrFolderNotEmpty
	}

	r, err := gg.CloneContext(ctx, memory.NewStorage(), nil, &gg.CloneOptions{
		URL:   url,
		Depth: 1,
	})
	if err != nil {
		return err
	}

	head, err := r.Head()
	if err != nil {
		return err
	}

	ref, err := r.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	ff, err := ref.Files()
	if err != nil {
		return err
	}

	if err := ff.ForEach(func(f *object.File) error {
		parts := strings.Split(f.Name, string(os.PathSeparator))
		fpath := filepath.Join(path, filepath.Join(parts...))

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		mode, err := f.Mode.ToOSFileMode()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		defer outFile.Close()

		r, err := f.Reader()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, r)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
