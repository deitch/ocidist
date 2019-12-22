package cmd

import (
	"fmt"
	"log"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List tags for a repository",
	Long:  `List all of the tags for a given repository in a given registry`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			tags []string
			err  error
		)

		repo, err := name.NewRepository(image)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", image, err)
		}

		simple, msg, options := apiOptions()

		log.Println(msg)
		if simple {
			tags, err = crane.ListTags(image)
		} else {
			tags, err = remote.List(repo, options...)
		}
		if err != nil {
			log.Fatalf("error listing tags: %v", err)
		}
		fmt.Println(tags)
	},
}

func tagsInit() {
}
