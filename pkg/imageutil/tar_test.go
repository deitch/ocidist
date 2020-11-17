package imageutil_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"testing"

	"github.com/deitch/ocidist/pkg/imageutil"
	"github.com/deitch/ocidist/pkg/util"
)

func TestMergeLayers(t *testing.T) {
	tests := []struct {
		in  [][]string
		out []string
		err error
	}{
		// basic mix single files
		{[][]string{{"/etc/foo"}, {"/a/b"}}, []string{"/a/b", "/etc/foo"}, nil},
		// basic mix multiple files
		{[][]string{{"/etc/foo", "/a/c"}, {"/a/b"}}, []string{"/a/b", "/etc/foo", "/a/c"}, nil},
		// duplicates
		{[][]string{{"/etc/foo", "/a/c", "/a/b", "/a/d"}, {"/a/b", "/a/.wh.c"}}, []string{"/a/b", "/a/b", "/a/d", "/etc/foo"}, nil},
		// whiteout single file
		{[][]string{{"/etc/foo", "/a/c"}, {"/a/b", "/a/.wh.c"}}, []string{"/a/b", "/etc/foo"}, nil},
		// whiteout single file with peers and duplicates
		{[][]string{{"/etc/foo", "/a/c", "/a/b", "/a/d"}, {"/a/b", "/a/.wh.c"}}, []string{"/a/b", "/a/b", "/a/d", "/etc/foo"}, nil},
		// whiteout directory
		{[][]string{{"/etc/foo", "/a/c", "/a/b", "/a/d"}, {"/.wh.a"}}, []string{"/etc/foo"}, nil},
		// opaque directory
		{[][]string{{"/etc/foo", "/a/c", "/a/b", "/a/d"}, {"/a/.wh..wh..opq"}}, []string{"/etc/foo"}, nil},
	}

	for i, tt := range tests {
		buf := bytes.NewBuffer(nil)
		rcgs := []util.GetReadCloser{}
		for _, files := range tt.in {
			// we do this to ensure we capture the current state when the func is called, otherwise
			// we get the last state of `files`
			infiles := files
			rcgs = append(rcgs, func() (io.ReadCloser, error) {
				// just make a simple tar stream
				buf := bytes.NewBuffer(nil)
				tw := tar.NewWriter(buf)
				for _, f := range infiles {
					tw.WriteHeader(&tar.Header{
						Name: f,
					})
				}
				tw.Close()

				return ioutil.NopCloser(buf), nil
			})
		}
		err := imageutil.ApplyLayers(buf, rcgs)
		// mismatched error
		if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) {
			t.Errorf("%d: mismatched error, actual %v expected %v", i, err, tt.err)
			continue
		}
		// check the files exist
		files, err := tarFiles(buf)
		if err != nil {
			t.Errorf("%d: unable to read filenames from tar stream: %v", i, err)
		}
		if !stringSliceEqualIgnoreOrder(files, tt.out) {
			t.Errorf("%d: mismatched files, actual %v expected %v", i, files, tt.out)
		}
	}
}

// tarFiles get the list of files in this tar stream
func tarFiles(r io.Reader) ([]string, error) {
	files := []string{}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			// pass the error on
			return nil, fmt.Errorf("tar file header read error: %v", err)
		}
		files = append(files, hdr.Name)
	}
	return files, nil
}

// stringSliceEqual compares 2 string slices and returns if their contents are identical.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, elm := range a {
		if elm != b[i] {
			return false
		}
	}
	return true
}

// stringSliceEqualIgnoreOrder compares 2 string slices and returns if their contents are identical, ignoring order
func stringSliceEqualIgnoreOrder(a, b []string) bool {
	a1, b1 := a[:], b[:]
	if a1 != nil && b1 != nil {
		sort.Strings(a1)
		sort.Strings(b1)
	}
	return stringSliceEqual(a1, b1)
}
