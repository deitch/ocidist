package layoututil

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/match"
	"github.com/google/go-containerregistry/pkg/v1/partial"
)

// GetCache get or initialize the cache
func GetCache(cache string) (layout.Path, error) {
	// initialize the cache path if needed
	p, err := layout.FromPath(cache)
	if err != nil {
		p, err = layout.Write(cache, empty.Index)
		if err != nil {
			return p, fmt.Errorf("could not initialize cache at path %s: %v", cache, err)
		}
	}
	return p, nil
}

func FindImageFromRoot(p layout.Path, imageName, architecture string) (v1.Image, error) {
	rootIndex, err := p.ImageIndex()
	// of there is no root index, we are broken
	if err != nil {
		return nil, err
	}
	// need to get the Image; if it is an Index, then resolve to our architecture
	var image v1.Image
	// first try the root tag as an image itself
	images, err := partial.FindImages(rootIndex, match.Name(imageName))
	if err == nil && len(images) > 0 {
		// if we found the root tag as an image, just use it
		image = images[0]
	} else {
		// we did not find the root tag as an image, it is an index, get the index
		indexes, err := partial.FindIndexes(rootIndex, match.Name(imageName))
		if err != nil || len(indexes) < 1 {
			return nil, fmt.Errorf("no image found in cache for %s", imageName)
		}
		ii := indexes[0]
		// we have the index, get the manifest that represents the manifest for the desired architecture
		platform := v1.Platform{OS: "linux", Architecture: architecture}
		images, err := partial.FindImages(ii, match.Platforms(platform))
		if err != nil || len(images) < 1 {
			return nil, fmt.Errorf("error retrieving image %s for platform %v from cache: %v", imageName, platform, err)
		}
		image = images[0]
	}
	return image, nil
}
