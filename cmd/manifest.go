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

var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Get the manifest for a specific tag",
	Long:  `Given a complete URL to an image, get the manifest and its sha256 hash`,
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			manifest []byte
			err      error
			desc     *remote.Descriptor
			ref      name.Reference
		)

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

		if showHash || verbose {
			log.Printf("referenced manifest hash sha256:%x\n", sha256.Sum256(manifest))
		}
		var out bytes.Buffer
		if err = json.Indent(&out, manifest, "", "\t"); err != nil {
			log.Fatalf("unable to indent json: %v", err)
		}
		fmt.Printf("%s\n\n", out.String())
	},
}

func manifestInit() {
	manifestCmd.Flags().BoolVar(&showHash, "hash", false, "show hashes for manifests and indexes")
}
