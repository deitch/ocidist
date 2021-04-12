package cmd

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/spf13/cobra"
)

var (
	blobLoadPath string
)

var pushBlobCmd = &cobra.Command{
	Use:   "blob <image>",
	Short: "Push a specific layer blob to a given repository from a local location",
	Long:  `For a given image URL, push one blob from a local file. Will return the hash. In the <image>, hash or tag is ignored.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			err   error
			ref   name.Reference
			layer v1.Layer
		)
		image := args[0]
		ref, err = name.ParseReference(image)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", image, err)
		}

		_, msg, options := apiOptions()

		log.Println(msg)

		// we will need to see if the provided path is actually a registry reference
		layerRef, layerErr := name.NewDigest(blobSavePath)

		switch {
		case blobSavePath == "-":
			layer = stream.NewLayer(ioutil.NopCloser(os.Stdin))
		case blobSavePath == "":
			log.Fatalf("must provide source to blob via --path")
		case layerErr == nil:
			layer, err = remote.Layer(layerRef, options...)
			if err != nil {
				log.Fatalf("recognized remote layer '%s' but had an error connecting to it: %v", layerRef, err)
			}
		default:
			f, err := os.Open(blobSavePath)
			if err != nil {
				log.Fatalf("could not open local file %s for reading to %s: %v", blobSavePath, ref.String(), err)
			}
			defer f.Close()
			layer = stream.NewLayer(f)
		}

		if err := remote.WriteLayer(ref.Context(), layer, options...); err != nil {
			log.Fatalf("error writing blob: %v", err)
		}
		digest, _ := layer.Digest()
		log.Printf("write blob to %s@%s", ref.String(), digest)
	},
}

func pushBlobInit() {
	pushBlobCmd.Flags().StringVar(&blobSavePath, "path", "", "path where to load the blob, use '-' for stdin, or a full reference for cross-mounting")
}
