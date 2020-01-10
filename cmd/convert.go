package cmd

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	legacytarball "github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	v1tarball "github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/spf13/cobra"
)

var (
	convertFromPath, convertToPath, convertToFormat, convertFromHash, convertTag string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a downloaded image from one format locally to another locally",
	Long:  `Convert a downloaded image from one format locally to another locally`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			img v1.Image
		)
		inputFormat, err := guessFormat(convertFromPath)
		if err != nil {
			log.Fatalf("unable to determine format of input file: %v", err)
		}
		switch inputFormat {
		case FORMATLAYOUT:
			p, err := layout.FromPath(convertFromPath)
			if err != nil {
				log.Fatalf("unable to get image from OCI layout on disk input: %v", err)
			}
			hash, err := v1.NewHash(convertFromHash)
			if err != nil {
				log.Fatalf("invalid hash %s: %v", convertFromHash, err)
			}
			img, err = p.Image(hash)
			if err != nil {
				log.Fatalf("unable to get image with hash %s from path %s: %v", hash.String(), convertFromPath, err)
			}
			if convertTag == "" {
				log.Fatal("must provide a tag when converting from an OCI layout on disk")
			}
		case FORMATV1:
			img, err = v1tarball.ImageFromPath(convertFromPath, nil)
			if err != nil {
				log.Fatalf("unable to get image from tarball input: %v", err)
			}
			// get the tag
			if convertTag == "" {
				tags, err := getTagsFromV1Tar(convertFromPath)
				if err != nil {
					log.Fatalf("unable to read tags from v1 tar at %s: %v", convertFromPath, err)
				}
				if len(tags) < 1 {
					log.Fatalf("no tags in tar file at %s and none provided on command line", convertFromPath)
				}
				convertTag = tags[0]
			}
		}

		// taken straight from pkg/crane.Save, but they don't have the options there
		ref, err := name.ParseReference(convertTag)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", convertTag, err)
		}
		tag, ok := ref.(name.Tag)
		if !ok {
			d, ok := ref.(name.Digest)
			if !ok {
				log.Fatalf("ref wasn't a tag or digest")
			}
			tag = d.Repository.Tag(digestTag)
		}

		// now write it to the output
		switch convertToFormat {
		case FORMATV1:
			err = v1tarball.WriteToFile(convertToPath, tag, img)
		case FORMATLEGACY:
			var w *os.File
			w, err = os.Create(convertToPath)
			if err != nil {
				log.Fatalf("unable to open %s to write legacy tar file: %v", convertToPath, err)
			}
			defer w.Close()
			err = legacytarball.Write(tag, img, w)
		default:
			err = fmt.Errorf("unknown format: %s", convertToFormat)
		}
		if err != nil {
			log.Fatalf("failure to write to %s in format %s: %v", convertToPath, convertToFormat, err)
		}

		log.Printf("saved to %s as format %s", convertToPath, convertToFormat)

	},
}

func convertInit() {
	// convertFromPath, convertToPath, convertToFormat
	convertCmd.Flags().StringVar(&convertToPath, "to", "", "path to output save the converted image as a tar file")
	convertCmd.MarkFlagRequired("to")
	convertCmd.Flags().StringVar(&convertFromPath, "from", "", "path to input to convert, must be a tar file or layout directory")
	convertCmd.MarkFlagRequired("from")
	convertCmd.Flags().StringVar(&convertToFormat, "format", "v1", "format to save the image, can be one of 'v1' or 'legacy'")
	convertCmd.Flags().StringVar(&convertFromHash, "hash", "", "when reading from an on-disk OCI layout, the hash of the image to extract, in 'sha256:<hash>' format")
	convertCmd.Flags().StringVar(&convertTag, "tag", "", "when reading from an on-disk OCI layout, the tag of the image as to be saved")
}

func guessFormat(p string) (string, error) {
	var (
		fi  os.FileInfo
		err error
	)

	// check our input file or directory exists
	if fi, err = os.Stat(convertFromPath); os.IsNotExist(err) {
		log.Fatalf("input path %s does not exist", convertFromPath)
	}

	if fi.IsDir() {
		return FORMATLAYOUT, nil
	}

	return FORMATV1, nil
}

func getTagsFromV1Tar(tarfile string) ([]string, error) {
	// open the tar file for reading
	var (
		f     *os.File
		err   error
		repob []byte
	)
	type tags map[string]string
	type apps map[string]tags

	// open the existing file
	if f, err = os.Open(tarfile); err != nil {
		return nil, err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	// cycle through until we find the "repositories" file
tarloop:
	for {
		header, err := tr.Next()

		switch {
		// if no more files are found
		case err == io.EOF:
			break tarloop
		case err != nil:
			return nil, fmt.Errorf("error reading tar entry: %v", err)
		case header == nil:
			continue
		// we only care about a regular file named "repositories"
		case header.Typeflag == tar.TypeReg:
			clean := filepath.Clean(header.Name)
			// we only are looking at the repositories file
			if clean != "repositories" {
				continue
			}
			repob, err = ioutil.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("error reading repositories file: %v", err)
			}
			// we already saved the bytes, so break; we are done with the file
			break tarloop
		}
	}

	// did we load anything?
	if len(repob) == 0 {
		return nil, nil
	}
	// load the json content of the "repositories" file into an apps struct
	var repos apps
	if err := json.Unmarshal(repob, &repos); err != nil {
		return nil, fmt.Errorf("error unmarshaling repositories file")
	}

	tagList := make([]string, 0)
	for reponame, v := range repos {
		for repotag, _ := range v {
			tagList = append(tagList, fmt.Sprintf("%s:%s", reponame, repotag))
		}
	}
	return tagList, nil
}
