package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var pullManifestCmd = &cobra.Command{
	Use:   "manifest <image>",
	Short: "Get the manifest for a specific tag",
	Long:  `Given a complete URL to an image, get the manifest and its sha256 hash.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			manifest []byte
			err      error
			desc     *remote.Descriptor
			ref      name.Reference
		)

		image := args[0]
		ref, err = name.ParseReference(image)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", image, err)
		}
		log.Printf("ref %#v\n", ref)

		simple, msg, options := apiOptions()

		log.Println(msg)
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

		if showInfo || verbose {
			log.Printf("referenced manifest hash sha256:%x size %d\n", sha256.Sum256(manifest), desc.Size)
		}
		var out bytes.Buffer
		if formatManifest {
			if err = json.Indent(&out, manifest, "", "\t"); err != nil {
				log.Fatalf("unable to indent json: %v", err)
			}
		} else {
			out = *bytes.NewBuffer(manifest)
		}
		fmt.Printf("%s", out.String())
	},
}

func pullManifestInit() {
	pullManifestCmd.Flags().BoolVar(&showInfo, "detail", false, "show additional detail for manifests and indexes, such as hash and size")
	pullManifestCmd.Flags().BoolVar(&formatManifest, "format", false, "format manifest for readability")
}
