package imageutil

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/match"
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
	// need to get the Image; if it is an Index, then resolve to our architecture
	var image v1.Image
	// first try the root tag as an image itself
	im, err := rootImage(p, match.Name(imageName))
	// if we found the root tag as an image, just use it
	if err == nil && im != nil {
		image = im
	} else {
		// we did not find the root tag as an image, it is an index, get the index
		ii, err := rootIndex(p, match.Name(imageName))
		if err != nil {
			return nil, fmt.Errorf("error retrieving image or index %s from cache: %v", imageName, err)
		}
		if ii == nil {
			return nil, fmt.Errorf("no image found in cache for %s", imageName)
		}
		// we have the index, get the manifest that represents the manifest for the desired architecture
		platform := v1.Platform{OS: "linux", Architecture: architecture}
		manifests, err := manifestsFromIndex(ii, match.Platform(platform))
		if err != nil || len(manifests) < 1 {
			return nil, fmt.Errorf("error retrieving image %s for platform %v from cache: %v", imageName, platform, err)
		}
		// we have the manifest, get the image
		image, err = ii.Image(manifests[0].Digest)
		if err != nil {
			return nil, fmt.Errorf("error making image from manifest for %s: %v", manifests[0].Digest, err)
		}
		/*
			blobReader, err := p.Blob(manifests[0].Digest)
			if err != nil {
				return nil, fmt.Errorf("error retrieving manifest %s for image %s for platform %v from cache: %v", manifests[0].Digest, imageName, platform, err)
			}
			defer blobReader.Close()
			manifest, err := v1.ParseManifest(blobReader)
			if err != nil {
				return nil, fmt.Errorf("error reading manifest %s for image %s for platform %v from cache: %v", manifests[0].Digest, imageName, platform, err)
			}
			// we have the blob, now create an Image from it
			image, err = p.Image(manifest.Digest)
			if err != nil {
				return nil, fmt.Errorf("unable to get image for digest %v from cache", manifest.Digest)
			}
		*/
	}
	return image, nil
}

func rootImage(p layout.Path, matcher match.Matcher) (v1.Image, error) {
	manifests, err := manifestsFromLayout(p, matcher)
	if err != nil || len(manifests) < 1 {
		return nil, err
	}
	rootIndex, err := p.ImageIndex()
	// of there is no root index, we are broken
	if err != nil {
		return nil, err
	}
	return rootIndex.Image(manifests[0].Digest)
}
func rootIndex(p layout.Path, matcher match.Matcher) (v1.ImageIndex, error) {
	manifests, err := manifestsFromLayout(p, matcher)
	if err != nil || len(manifests) < 1 {
		return nil, err
	}
	rootIndex, err := p.ImageIndex()
	// of there is no root index, we are broken
	if err != nil {
		return nil, err
	}
	return rootIndex.ImageIndex(manifests[0].Digest)
}

// manifestsFromLayout given a layout and a matcher function, return all
// of the manifests that match
func manifestsFromLayout(p layout.Path, matcher match.Matcher) ([]v1.Descriptor, error) {
	// get the root index from index.json
	rootIndex, err := p.ImageIndex()
	// of there is no root index, we are broken
	if err != nil {
		return nil, err
	}
	return manifestsFromIndex(rootIndex, matcher)
}

func manifestsFromIndex(index v1.ImageIndex, matcher match.Matcher) ([]v1.Descriptor, error) {
	// get the actual manifest list
	indexManifest, err := index.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("unable to get raw index: %v", err)
	}
	manifests := []v1.Descriptor{}
	// try to get the root of our image
	for _, manifest := range indexManifest.Manifests {
		if matcher(manifest) {
			manifests = append(manifests, manifest)
		}
	}
	return manifests, nil
}
