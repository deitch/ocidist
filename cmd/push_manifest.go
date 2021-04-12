package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var pushManifestCmd = &cobra.Command{
	Use:   "manifest <image>",
	Short: "Push a manifest",
	Long:  `Create a new manifest, marking it with its hash, and saving the to the image repository provided.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			b   []byte
			err error
		)

		image := args[0]
		ref, err := name.ParseReference(image)
		if err != nil {
			log.Fatalf("error parsing name '%s': %v", image, err)
		}

		_, _, options := apiOptions()

		switch {
		case manifestSavePath == "-":
			b, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("could not read from stdin for reading to %s: %v", image, err)
			}
		case manifestSavePath == "":
			log.Fatalf("must provide source for manifest via --path")
		default:
			b, err = ioutil.ReadFile(manifestSavePath)
			if err != nil {
				log.Fatalf("could not open local file %s for reading to %s: %v", manifestSavePath, image, err)
			}
		}
		hash, _, err := v1.SHA256(bytes.NewReader(b))
		if err != nil {
			log.Fatalf("error calculating hash of manifest: %v", err)
		}

		manifest := taggableBytes{b}

		// this is cheating, since go-containerregistry doesn't support actually writing directly, but the API does,
		// see https://docs.docker.com/registry/spec/api/#manifest
		dig, err := name.NewDigest(fmt.Sprintf("%s@%s", ref.Context().String(), hash.String()))
		if err != nil {
			log.Fatalf("error creating manifest for digest %s: %v", dig, err)
		}

		if err := remote.Put(dig, manifest, options...); err != nil {
			log.Fatalf("error writing manifest for digest %s: %v", dig, err)
		}
		log.Printf("successfully wrote reference: %s", dig)
	},
}

func pushManifestInit() {
	pushManifestCmd.Flags().StringVar(&manifestSavePath, "path", "", "path where to retrieve the manifest, blank defaults to stdin")
}
