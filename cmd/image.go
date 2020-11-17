package cmd

import (
	"io"
	"log"
	"os"
	"runtime"

	"github.com/deitch/ocidist/pkg/imageutil"
	"github.com/deitch/ocidist/pkg/layoututil"
	"github.com/deitch/ocidist/pkg/util"

	"github.com/spf13/cobra"
)

var (
	layoutPath   string
	rootDir      string
	targetPath   string
	architecture string
)

var mergeImageCmd = &cobra.Command{
	Use:   "merge <ref>",
	Short: "merge the layers of an image in a local layout into a single tar file, applying all layers",
	Long: `For an image located locally in a v1/layout, merge all of the layers of the the image to get a single tar file representing the image filesystem
If the provided image is an index, will use the provided architecture, defaulting to the local machine architecture.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		imageName := args[0]

		// get the cache
		p, err := layoututil.GetCache(layoutPath)
		if err != nil {
			log.Fatalf("unable to get v1 layout at %s: %v", layoutPath, err)
		}

		// get a reference to the image
		image, err := layoututil.FindImageFromRoot(p, imageName, architecture)
		if err != nil {
			log.Fatalf("unable to get root image for %s at %s: %v", imageName, layoutPath, err)
		}

		// get the layers
		layers, err := image.Layers()
		if err != nil {
			log.Fatalf("unable to get layers for %s at %s: %v", imageName, layoutPath, err)
		}

		layerReaders := []util.GetReadCloser{}
		for _, layer := range layers {
			layerReaders = append(layerReaders, func() (io.ReadCloser, error) {
				return layer.Uncompressed()
			})
		}
		outfile, err := os.Create(targetPath)
		if err != nil {
			log.Fatalf("unable to open target file %s: %v", targetPath, err)
		}
		defer outfile.Close()
		if err := imageutil.ApplyLayers(outfile, layerReaders); err != nil {
			log.Fatalf("could not merge layers: %v", err)
		}
		log.Printf("Done! Image expanded at %s", targetPath)
	},
}

func mergeImageInit() {
	mergeImageCmd.Flags().StringVar(&layoutPath, "path", "", "path to the local v1 layout")
	mergeImageCmd.Flags().StringVar(&targetPath, "target", "", "where to write the output tar file")
	mergeImageCmd.Flags().StringVar(&architecture, "arch", runtime.GOARCH, "architecture for which to build an image")
}
