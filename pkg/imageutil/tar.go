package imageutil

import (
	"archive/tar"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/deitch/ocidist/pkg/util"
)

const (
	whiteoutPrefix    = ".wh."
	whiteoutOpaqueDir = whiteoutPrefix + whiteoutPrefix + ".opq"
)

func ApplyLayers(w io.Writer, layers []util.GetReadCloser) error {

	// we cannot simply expand it to a directory, since we cannot necessarily do the right permissions
	// or ownership. We want to preserve those in a single tar file.
	// Instead, we do a two-pass. In the first pass, we track every entry to build up what file should exist,
	// adding when we see one, deleting when we see a whiteout. In the second, we actually write the files,
	// based on the index created in the first pass.

	fileIndex := map[string]bool{}

	// first pass, find out each file
	for i, layer := range layers {
		rc, err := layer()
		if err != nil {
			return fmt.Errorf("could not get ReadCloser for layer %d: %v", i, err)
		}
		tr := tar.NewReader(rc)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				// clear the error, since we do not pass an io.EOF
				err = nil
				break // End of archive
			}
			if err != nil {
				// pass the error on
				return fmt.Errorf("tar file header read error: %v", err)
			}
			// is it a whiteout file?
			base := path.Base(hdr.Name)
			dir := path.Dir(hdr.Name)
			switch {
			case base == whiteoutOpaqueDir:
				// opaque dir means to delete all of the children but keep the item itself
				removedFull := path.Clean(dir)
				for k := range fileIndex {
					if k != removedFull && k != removedFull+"/" && strings.HasPrefix(k, removedFull+"/") {
						delete(fileIndex, k)
					}
				}
			case strings.HasPrefix(base, whiteoutPrefix):
				// whiteout means to delete this file/dir and all children
				removedBase := base[len(whiteoutPrefix):]
				removedFull := path.Clean(path.Join(dir, removedBase))
				// need to delete this and every child, in case it is a dir
				delete(fileIndex, removedFull)
				// this is inefficient, looping through every time we find a dir; a tree might be more efficient, for later
				for k := range fileIndex {
					if strings.HasPrefix(k, removedFull+"/") {
						delete(fileIndex, k)
					}
				}
			default:
				fileIndex[hdr.Name] = true
			}
		}
		if err := rc.Close(); err != nil {
			return fmt.Errorf("error closing layer %d: %v", i, err)
		}
	}
	// now we need a writer
	tw := tar.NewWriter(w)
	defer tw.Close()
	// second pass, write it all out
	for i, layer := range layers {
		rc, err := layer()
		if err != nil {
			return fmt.Errorf("could not get ReadCloser for layer %d: %v", i, err)
		}
		tr := tar.NewReader(rc)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				// clear the error, since we do not pass an io.EOF
				err = nil
				break // End of archive
			}
			if err != nil {
				// pass the error on
				return fmt.Errorf("tar file header read error: %v", err)
			}
			// does this file exist?
			if _, ok := fileIndex[hdr.Name]; !ok {
				continue
			}
			// if we made it this far, this file was not whited out, just stream the header and content
			if err := tw.WriteHeader(hdr); err != nil {
				return fmt.Errorf("error writing header for layer %d file %s: %v", i, hdr.Name, err)
			}
			if _, err := io.Copy(tw, tr); err != nil {
				return fmt.Errorf("error writing file for layer %d file %s: %v", i, hdr.Name, err)
			}
		}
		if err := rc.Close(); err != nil {
			return fmt.Errorf("error closing layer %d: %v", i, err)
		}
	}

	return nil
}
