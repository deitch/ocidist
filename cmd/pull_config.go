package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/match"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var (
	formatConfig bool
	platform     string
)
var pullConfigCmd = &cobra.Command{
	Use:   "config <image>",
	Short: "Get the config for a specific tag",
	Long: `Given a complete URL to an image, get the config for it. If the reference is an index, rather than a single manifest,
	it will resolve to whatever platform you provide, defaulting to your current arch, and 'linux' if your platform is not supported.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// this is the manifest referenced by the image. If it is an index, it returns the index.
		var (
			config []byte
			err    error
			desc   *remote.Descriptor
			ref    name.Reference
		)

		image := args[0]
		ref, err = name.ParseReference(image)
		if err != nil {
			log.Fatalf("parsing reference %q: %v", image, err)
		}
		log.Printf("ref %#v\n", ref)

		_, msg, options := apiOptions()

		if platform != "" {
			parts := strings.SplitN(platform, "/", 2)
			os, arch := parts[0], parts[1]
			options = append(options, remote.WithPlatform(v1.Platform{Architecture: arch, OS: os}))
		}

		log.Println(msg)
		desc, err = remote.Get(ref, options...)
		if err != nil {
			log.Fatalf("error getting manifest: %v", err)
		}

		// did we have an index or a manifest?
		img, err := desc.Image()
		if err != nil {
			ii, err := desc.ImageIndex()
			if err == nil {
				log.Fatalf("root was neither image nor index")
			}
			if platform == "" {
				log.Fatalf("referenced index, but platform not provided")
			}
			imgs, err := partial.FindImages(ii, match.Platforms())
			if err != nil {
				log.Fatalf("error finding image for platform %s: %v", platform, err)
			}
			if len(imgs) < 1 {
				log.Fatalf("no images found for platform %s", platform)
			}
			img = imgs[0]
		}

		config, err = img.RawConfigFile()
		if err != nil {
			log.Fatalf("error getting config file: %v", err)
		}

		if showInfo || verbose {
			log.Printf("referenced config hash sha256:%x size %d\n", sha256.Sum256(config), len(config))
		}
		var out bytes.Buffer
		if formatConfig {
			if err = json.Indent(&out, config, "", "\t"); err != nil {
				log.Fatalf("unable to indent json: %v", err)
			}
		} else {
			out = *bytes.NewBuffer(config)
		}
		fmt.Printf("%s", out.String())
	},
}

func pullConfigInit() {
	pullConfigCmd.Flags().BoolVar(&showInfo, "detail", false, "show additional detail for config, such as hash and size")
	pullConfigCmd.Flags().BoolVar(&formatConfig, "format", false, "format config for readability")
	pullConfigCmd.Flags().StringVar(&platform, "platform", "", "which platform to show, in case of a referenced index, in format 'os/arch', e.g. 'linux/amd64'")
}
