# OCI Distribution Utility

This is a simple OCI distribution utility, based on [https://github.com/google/go-containerregistry/](https://github.com/google/go-containerregistry/), which allows you to get the tags for an image, get the manifest for a reference, and pull the actual referenced image to a local tar file.

It can exercise both the simple [go-containerregistry](https://github.com/google/go-containerregistry/) API and the advanced one, depending on the options you choose.

It supports the following commands:

* `tags` - list the tags for an image, e.g. `ocidist tags docker.io/library/alpine`
* `manifest` - get the manifest for an image reference, e.g. `ocidist manifest docker.io/library/alpine:3.10`
* `pull` - pull an image based on its reference, e.g. `ocidist pull docker.io/library/alpine:3.10 --path /tmp/foo.tar `

A pulled image will be saved in the standard tar file format used for `docker save` and `docker load`, as well as `docker2aci` for `rkt`.

## Manifests

When using the `manifest` command, you will get the referenced manifests. When using the pull command, you also can get the manifest, as well as the resolved manifest for an image index. You also can get optional hashes for both.

### manifest command

When using the `manifest` command, `ocidist` will provide you with the output manifest in a proper json-formatted string, along with the hash of the manifest. This is the manifest to which you referred directly (technically, an OCI descriptor). This, if you provided a reference to an actual image, e.g. `docker.io/library/alpine@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a`, you will get the manifest for that image:

```sh
$ ./dist/ocidist manifest docker.io/library/alpine@sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a
simple API
referenced manifest e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a
{
	"schemaVersion": 2,
	"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
	"config": {
		"mediaType": "application/vnd.docker.container.image.v1+json",
		"size": 1512,
		"digest": "sha256:965ea09ff2ebd2b9eeec88cd822ce156f6674c7e99be082c7efac3c62f3ff652"
	},
	"layers": [
		{
			"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
			"size": 2787134,
			"digest": "sha256:89d9c30c1d48bac627e5c6cb0d1ed1eec28e7dbdfbcc04712e4c79c0f83faf17"
		}
	]
}
```

On the other hand, if you provide a reference to an index, e.g. `docker.io/library/alpine:3.10`, you will get the manifest for that index:

```sh
$ ./dist/ocidist manifest docker.io/library/alpine:3.10
simple API
referenced manifest c19173c5ada610a5989151111163d28a67368362762534d8a8121ce95cf2bd5a
{
	"manifests": [
		{
			"digest": "sha256:e4355b66995c96b4b468159fc5c7e3540fcef961189ca13fee877798649f531a",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "amd64",
				"os": "linux"
			},
			"size": 528
		},
		{
			"digest": "sha256:29a82d50bdb8dd7814009852c1773fb9bb300d2f655bd1cd9e764e7bb1412be3",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "arm",
				"os": "linux",
				"variant": "v6"
			},
			"size": 528
		},
		{
			"digest": "sha256:915a0447d045e3b55f84e8456de861571200ee39f38a0ce70a45f91c29491a21",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "arm",
				"os": "linux",
				"variant": "v7"
			},
			"size": 528
		},
		{
			"digest": "sha256:1827be57ca85c28287d18349bbfdb3870419692656cb67c4cd0f5042f0f63aec",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "arm64",
				"os": "linux",
				"variant": "v8"
			},
			"size": 528
		},
		{
			"digest": "sha256:77cbe97593c890eb1c4cadcbca37809ebff2b5f46a036666866c99f08a708967",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "386",
				"os": "linux"
			},
			"size": 528
		},
		{
			"digest": "sha256:6dff84dbd39db7cb0fc928291e220b3cff846e59334fd66f27ace0bcfd471b75",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "ppc64le",
				"os": "linux"
			},
			"size": 528
		},
		{
			"digest": "sha256:d8d321ec5eec88dee69aec467a51fe764daebdb92ecff0d1debd09840cbd86c6",
			"mediaType": "application\/vnd.docker.distribution.manifest.v2+json",
			"platform": {
				"architecture": "s390x",
				"os": "linux"
			},
			"size": 528
		}
	],
	"mediaType": "application\/vnd.docker.distribution.manifest.list.v2+json",
	"schemaVersion": 2
}
```

### pull command

When using the `pull` command, if the reference is to an image manifest, then it will give you the manifest for that image. On the other hand, if the reference is to an image _index_, it will do the following:

1. Give you the manifest (and hash) for the index
1. Resolve the image index to a specific image manifest for your platform
1. Give you the manifest (and hash) for the platform-specific image manifest
1. Pull the image for your platform

## Options

### image

You **must** set the name of the image as the first argument. This must be in the format:

```
<host>[:<port>]/path[:tag][@sha256:hash]
```

Where:

* `host` is the host, e.g. `docker.io` ; _REQUIRED_
* `port` is the port, e.g. `:443` ; _OPTIONAL_ unless non-default
* `path` is the path, e.g. `/library/alpine` ; _REQUIRED_
* `tag` is the tag, e.g. `3.10` or `latest` ; _REQUIRED_ for commands `manifest` and `pull`, unless `hash` is given, _FORBIDDEN_ for command `tags`
* `hash` is the sha256 hash of the image manifest or index, e.g. `@sha256:c19173c5ada610a5989151111163d28a67368362762534d8a8121ce95cf2bd5a` ; _REQUIRED_ for commands `manifest` and `pull`, unless `hash` is given, _FORBIDDEN_ for command `tags`

#### Examples

For commands `manifest` and `pull`:

* `docker.io/library/alpine:3.10@sha256:c19173c5ada610a5989151111163d28a67368362762534d8a8121ce95cf2bd5a`
* `docker.io/library/alpine:3.10`
* `docker.io/library/alpine@sha256:c19173c5ada610a5989151111163d28a67368362762534d8a8121ce95cf2bd5a`

Following the OCI distribution convention, you can supply the tag, the hash, or both, but must provide at least one.

For command `tags`:

* `docker.io/library/alpine`

You must provide _neither_ the tag, nor the hash.

### Authentication

By default, `ocidist` uses whatever is configured on your system, normally the `~/.docker/config.json`. You can override it with the following options:

* `--anonymous` - use anonymous authentication
* `--username user --password pass` - use the provided credentials

### http client

By default, `ocidist` uses the default http client that [go-containerregistry](https://github.com/google/go-containerregistry/) provides from Golang's package [net/http](https://golang.org/pkg/net/http/). You can choose to provide a custom http client by setting the option `--http`. This doesn't change much, but does exercise providing an override.

### http proxy

You can set the http proxy URL, which also overrides the http client, by setting the option `--proxy url`.

## Releases

We have not cut any releases, so you still need to build it on your own with `make build`. We would be happy to consider it.
