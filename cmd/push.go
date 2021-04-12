package cmd

import (
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push parts of an image to a registry, from stdin or local files",
	Long: `Send various parts of an image to a registry, given a complete URL to an image, including manifest, config, tags, for an entire image.
	It is important to recall that registries garbage-collect orphaned things, so pushing out just a blob or just a manifest means it is likely to
	disappear within a few hours or days, unless they are referenced by a tag.`,
}

func pushInit() {
	pushCmd.AddCommand(pushBlobCmd)
	pushBlobInit()
	pushCmd.AddCommand(pushManifestCmd)
	pushManifestInit()
	pushCmd.AddCommand(pushTagCmd)
	pushTagInit()
}
