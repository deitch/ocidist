package cmd

import (
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull parts of an image from a registry, saving to stadout or local files",
	Long:  `Given a complete URL to an image, get various parts for it, including manifest, config, tags, etc.`,
}

func pullInit() {
	pullCmd.AddCommand(pullImageCmd)
	pullImageInit()
	pullCmd.AddCommand(pullBlobCmd)
	pullBlobInit()
	pullCmd.AddCommand(pullManifestCmd)
	pullManifestInit()
	pullCmd.AddCommand(pullConfigCmd)
	pullConfigInit()
	pullCmd.AddCommand(pullTagsCmd)
	pullTagsInit()
}
