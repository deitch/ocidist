package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	legacytarball "github.com/google/go-containerregistry/pkg/legacy/tarball"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	v1tarball "github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/spf13/cobra"
)

const (
	digestTag    = "digest-without-tag"
	FORMATV1     = "v1"
	FORMATLEGACY = "legacy"
	FORMATLAYOUT = "layout"
)

var (
	savePath, writeFormat string
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the image for a given repository and save it as a tar file",
	Long:  `For a given complete image URL, pull it and save it to a local tar file`,
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			manifest []byte
			img      v1.Image
			desc     *remote.Descriptor
			//sum      [sha256.Size]byte
			err error
			ref name.Reference
		)

		ref, err = name.ParseReference(image)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", image, err)
		}

		simple, msg, options := apiOptions()

		log.Println(msg)

		// first get the root manifest. This might be an index or a manifest
		if simple {
			manifest, err = crane.Manifest(image)
			if err != nil {
				log.Fatalf("error getting manifest: %v", err)
			}
		} else {
			desc, err = remote.Get(ref, options...)
			if err != nil {
				log.Fatalf("error getting manifest: %v", err)
			}
			manifest = desc.Manifest
		}
		if showHash || verbose {
			log.Printf("referenced manifest %x\n", sha256.Sum256(manifest))
		}
		var out bytes.Buffer
		if err = json.Indent(&out, manifest, "", "\t"); err != nil {
			log.Fatalf("unable to indent json: %v", err)
		}
		fmt.Printf("%s\n\n", out.String())

		// This is where it gets the image manifest, but does not actually save anything
		// It is the manifest of the image itself, not of the index (if it is
		// an index), so it actually does resolve platform-specific
		start := time.Now()
		if simple {
			img, err = crane.Pull(image)
		} else {
			img, err = desc.Image()
			//img, err = remote.Image(ref, options...)
		}
		if err != nil {
			log.Fatalf("error pulling image ref: %v", err)
		}
		log.Printf("ended pull, duration %d milliseconds", time.Since(start).Milliseconds())

		// check out the manifest and hash
		manifest, err = img.RawManifest()
		if err != nil {
			log.Fatalf("error getting manifest: %v", err)
		}
		digest, err := img.Digest()
		if err != nil {
			log.Fatalf("error getting digest: %v", err)
		}
		if showHash || verbose {
			log.Printf("image manifest %s\n", digest.Hex)
		}
		fmt.Println(string(manifest))

		// This is where it uses the manifest to save the layers
		start = time.Now()
		if simple {
			err = crane.Save(img, image, savePath)
		} else {
			// taken straight from pkg/crane.Save, but they don't have the options there
			tag, ok := ref.(name.Tag)
			if !ok {
				d, ok := ref.(name.Digest)
				if !ok {
					log.Fatalf("ref wasn't a tag or digest")
				}
				tag = d.Repository.Tag(digestTag)
			}

			switch writeFormat {
			case FORMATV1:
				err = v1tarball.WriteToFile(savePath, tag, img)
			case FORMATLEGACY:
				w, err := os.Create(savePath)
				if err != nil {
					log.Fatalf("unable to open %s to write legacy tar file: %v", savePath, err)
				}
				defer w.Close()
				err = legacytarball.Write(tag, img, w)
			case FORMATLAYOUT:
				ii, err := desc.ImageIndex()
				if err != nil {
					log.Fatalf("provided image is not an index: %s", image)
				}
				_, err = layout.Write(savePath, ii)
			default:
				err = fmt.Errorf("unknown format: %s", writeFormat)
			}
		}
		if err != nil {
			log.Fatalf("error saving: %v", err)
		}
		log.Printf("ended save, duration %d milliseconds", time.Since(start).Milliseconds())
		log.Printf("saved in tar format to %s", savePath)

	},
}

func pullInit() {
	pullCmd.Flags().StringVar(&savePath, "path", "", "path to save the image as a tar file, or directory for layout")
	pullCmd.MarkFlagRequired("path")
	pullCmd.Flags().BoolVar(&showHash, "hash", false, "show hashes for manifests and indexes")
	pullCmd.Flags().StringVar(&writeFormat, "format", "v1", "format to save the image, can be one of 'v1', 'layout', 'legacy'")
}
