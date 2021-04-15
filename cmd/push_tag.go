package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var (
	manifestSavePath string
	manifestSaveHash string
)

type taggableBytes struct {
	b []byte
}

func (t taggableBytes) RawManifest() ([]byte, error) {
	return t.b, nil
}

var pushTagCmd = &cobra.Command{
	Use:   "tag <image:tag>",
	Short: "Push a tag pointing to a hash",
	Long: `Provide an image with a tag pointing to a manifest hash. Can either provide hash of existing manifest in the registry,
	or one from stdin or a file. Must provide exactly one of --path or --hash`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var (
			manifest remote.Taggable
		)

		image := args[0]
		tag, err := name.NewTag(image)
		if err != nil {
			log.Fatalf("error creating manifest: %v", err)
		}

		_, _, options := apiOptions()

		switch {
		case manifestSaveHash != "" && manifestSavePath != "":
			log.Fatalf("must provide exactly one of '--path' or '--hash'")
		case manifestSaveHash != "":
			ref, err := name.NewDigest(fmt.Sprintf("%s@%s", tag.Context().String(), manifestSaveHash))
			desc, err := remote.Get(ref, options...)
			if err != nil {
				log.Fatalf("error getting manifest: %v", err)
			}
			manifest = desc
		case manifestSavePath == "-":
			b, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("could not read from stdin for reading to %s: %v", image, err)
			}
			manifest = taggableBytes{b}
		case manifestSavePath != "":
			b, err := ioutil.ReadFile(manifestSavePath)
			if err != nil {
				log.Fatalf("could not open local file %s for reading to %s: %v", manifestSavePath, image, err)
			}
			manifest = taggableBytes{b}
		}

		if err := remote.Tag(tag, manifest, options...); err != nil {
			log.Fatalf("error writing tag: %v", err)
		}
		log.Printf("successfully wrote tag: %s", image)
	},
}

func pushTagInit() {
	pushTagCmd.Flags().StringVar(&manifestSavePath, "path", "", "path where to retrieve the manifest, blank defaults to stdin")
	pushTagCmd.Flags().StringVar(&manifestSaveHash, "hash", "", "hash of existing manifest to use")
}
