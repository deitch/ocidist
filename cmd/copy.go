package cmd

import (
	"crypto/sha256"
	"log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy <from:tag> <to-tag>",
	Short: "copy a tag on a registry from one to another, creating the new one",
	Long: `Given an image tag that already exists on a registry, create a new one pointing to the same root manifest. The <to-tag> should
just be a tag, not a full name. For example:

copy docker.io/foo/bar:sometag othertag
`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			manifest []byte
			err      error
			desc     *remote.Descriptor
			ref      name.Reference
		)

		image, to := args[0], args[1]
		ref, err = name.ParseReference(image)
		if err != nil {
			log.Fatalf("parsing from reference %q: %v", image, err)
		}
		log.Printf("ref %#v\n", ref)

		_, msg, options := apiOptions()

		log.Println(msg)
		desc, err = remote.Get(ref, options...)
		if err != nil {
			log.Fatalf("error getting manifest: %v", err)
		}
		manifest = desc.Manifest

		if showInfo || verbose {
			log.Printf("referenced manifest hash sha256:%x size %d\n", sha256.Sum256(manifest), desc.Size)
		}

		totag := ref.Context().Tag(to)

		log.Printf("totag: %#v", totag)

		if err := remote.Tag(totag, desc, options...); err != nil {
			log.Fatalf("error pushing up new tag %s: %v", to, err)
		}
		log.Printf("done, copied %s to %s", image, to)
	},
}

func copyInit() {
}
