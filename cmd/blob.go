package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var (
	blobSavePath string
	isManifest   bool
)

var pullBlobCmd = &cobra.Command{
	Use:   "blob <ref>",
	Short: "Pull a specific layer blob for a given repository and save it locally",
	Long: `For a given complete image URL, pull one blob and save it locally in the target format. To get a specific blob,
provide the <ref> with a hash, e.g. docker.io/library/alpine@abcdef5566. To get the manifest referenced by a tag, provide the <ref>
in the usual format, e.g. docker.io/library/alpine:3.11`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			manifest []byte
			desc     *remote.Descriptor
			//sum      [sha256.Size]byte
			err error
			ref name.Reference
			w   io.Writer
			r   io.Reader
		)
		image := args[0]
		ref, err = name.ParseReference(image)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", image, err)
		}

		_, msg, options := apiOptions()

		log.Println(msg)

		if _, ok := ref.(name.Tag); ok || isManifest {
			// we had a tag, so just get the root manifest/index
			log.Printf("requested manifest or had tag without hash, so just pulling root for %s", image)
			desc, err = remote.Get(ref, options...)
			if err != nil {
				log.Fatalf("error getting manifest: %v", err)
			}
			manifest = desc.Manifest
			var out bytes.Buffer
			if err = json.Indent(&out, manifest, "", "\t"); err != nil {
				log.Fatalf("unable to indent json: %v", err)
			}
			r = strings.NewReader(out.String())
		} else {
			// we had a hash, so get the actual layer
			d, ok := ref.(name.Digest)
			if !ok {
				log.Fatalf("ref wasn't a tag or digest")
			}
			log.Printf("had hash, so pulling blob for %s", image)
			layer, err := remote.Layer(d, options...)
			if err != nil {
				log.Fatalf("could not pull layer %s: %v", ref.String(), err)
			}
			// write the layer out to the file
			lr, err := layer.Compressed()
			if err != nil {
				log.Fatalf("could not get layer reader %s: %v", ref.String(), err)
			}
			defer lr.Close()
			r = lr
		}

		if blobSavePath != "" {
			f, err := os.Create(blobSavePath)
			if err != nil {
				log.Fatalf("could not open local file %s for writing from %s: %v", blobSavePath, ref.String(), err)
			}
			defer f.Close()
			w = f
		} else {
			w = os.Stdout
		}
		_, err = io.Copy(w, r)
		if err != nil {
			log.Fatalf("could not write to local file %s from %s: %v", blobSavePath, ref.String(), err)
		}

		if w != os.Stdout {
			log.Printf("saved to %s", blobSavePath)
		}

		/*



		 */
	},
}

func pullBlobInit() {
	pullBlobCmd.Flags().StringVar(&blobSavePath, "path", "", "path to save the blob, blank defaults to stdout")
	pullBlobCmd.Flags().BoolVar(&isManifest, "manifest", false, "whether the requested item is a manifest/index or not; defaults to false if a hash is provided, true otherwise")
}
