# Hashes

This doc explains how to understand the various hashes that float around when working with manifests and images.

There are several hashes to consider.

* list: if you pull a manifest list, then the hash reported is that for the manifest list, not the manifest of the referenced image for your platform. E.g. 50cf965a6e08ec5784009d0fccb380fc479826b6e0e65684d9879170a9df8566 in the above example is the hash of the index, not the referenced one for amd64, where I ran the examples.
* manifest: if you pull a specific image, then you will get the hash for the manifest itself. E.g. in the above example is the hash of the manifest for the instance on  
* inspect: when you inspect the image, you get the hash of the config, the specific sub-field of a manifest, and not of the complete manifest itself. E.g. 231d40e811cd970168fb0c4770f2161aa30b9ba6fe8e68527504df69643aa145 in the above is the hash of the config.
* layers: each layer, which is a single binary blob, is hashed and is content-addressable. The hash of each layer is given in the `Layers` field. In the example above for `docker pull`, there are three layers, one each for 000eee12ec04, eb22865337de, bee5d581ef8b. 

Here is where each one shows up, and what it represents.

When you do docker pull, it pulls the manifest down. If that manifest actually is an index, then it resolves it to the specific manifest for the given image and pulls that. Then it pulls the layers. The hashes reported are:

* docker pull
  * layers: hash of the layer itself, precisely as it appears in the manifest `layers:` field.
  * digest reported at end of pull: hash of the manifest you passed to `docker pull`. So if the URI you passed refers to a single image manifest, you get the hash of the single image manifest; if it refers to an index (manifest list), you get the hash of the index. It would be great if it returned both, but it doesn't.
* docker image ls: hash of the config in the manifest (not the manifest itself) of the specific image for your system.
* Manifest tool: all of the above. If you run manifest tool on a manifest, then you get the hash of the manifest at the beginning; if you run it on an index, you get the hash of the index. When it is an index, you then get the hashes for each image in the index. Notably, the hash for each item is the hash of the manifest for that image, and the hashes of the layers are precisely the hashes of those layers as referred to in the manifest.

Let's try to tie this all together with an example. We will take the index `library/nginx:latest`, which, as of this writing, has the tag `1.17.6`, and the following hashes:

* index hash: 50cf965a6e08ec5784009d0fccb380fc479826b6e0e65684d9879170a9df8566
* image hash for amd64: 189cce606b29fb2a33ebc2fcecfa8e33b0b99740da4737133cdbcee92f3aba0a
* config hash for amd64: 231d40e811cd970168fb0c4770f2161aa30b9ba6fe8e68527504df69643aa145
* layers for amd64:
  1. 000eee12ec04cc914bf96e8f5dee7767510c2aca3816af6078bd9fbe3150920c
  1. eb22865337de3edb54ec8b52f6c06de320f415e7ec43f01426fdafb8df6d6eb7
  1. 5c31a34dd429119ed5e032f77a32a0f209dd72036d904b91a406c1949ff71726

Here is my incomplete output from running the various commands on my `amd64` laptop:

```
$ docker pull nginx:latest
latest: Pulling from library/nginx
000eee12ec04: Pull complete
eb22865337de: Pull complete
bee5d581ef8b: Pull complete
Digest: sha256:50cf965a6e08ec5784009d0fccb380fc479826b6e0e65684d9879170a9df8566
Status: Downloaded newer image for nginx:latest
docker.io/library/nginx:latest
```

The above gives the final digest of `50cf`, which is the digest of the index, while each layer pulled, listed as `Pull complete`, is the hash of the specific layer in the registry. Notice that these align precisely with the hashes in my summary table.

```
$ docker image inspect library/nginx:latest | jq
[
  {
    "Id": "sha256:231d40e811cd970168fb0c4770f2161aa30b9ba6fe8e68527504df69643aa145",
    "RepoTags": [
      "nginx:latest"
    ],
    "RepoDigests": [
      "nginx@sha256:50cf965a6e08ec5784009d0fccb380fc479826b6e0e65684d9879170a9df8566"
    ],
    ...
    "RootFS": {
      "Type": "layers",
      "Layers": [
        "sha256:831c5620387fb9efec59fc82a42b948546c6be601e3ab34a87108ecf852aa15f",
        "sha256:5fb987d2e54d85820d95d6c31f3fe4cd95bf71fe6d9d9e4684082cb551b728b0",
        "sha256:4fc1aa8003a3d0d2481f10d17773869cbff12c1008df30e0bab8259086a0311c"
      ]
    },
```

The `RepoDigests` gives the tags used to pull this, in this case `nginx`, with the correct index hash from above of `50cf`. The `Id`, on the other hand, is the hash of the _config_, not the manifest itself that appears on the registry. This isn't all that useful, but it is good to know.

The `Layers`, on the other hand, are not at all clear what they represent.

```
$ docker image ls | grep nginx
nginx                                           latest                                                    231d40e811cd        3 weeks ago         126MB
```

Once again, this is the same `Id` as given in `docker image inspect`.

```
$ manifest-tool inspect nginx:latest
Name:   nginx:latest (Type: application/vnd.docker.distribution.manifest.list.v2+json)
Digest: sha256:50cf965a6e08ec5784009d0fccb380fc479826b6e0e65684d9879170a9df8566
 * Contains 6 manifest references:
1    Mfst Type: application/vnd.docker.distribution.manifest.v2+json
1       Digest: sha256:189cce606b29fb2a33ebc2fcecfa8e33b0b99740da4737133cdbcee92f3aba0a
1  Mfst Length: 948
1     Platform:
1           -      OS: linux
1           - OS Vers:
1           - OS Feat: []
1           -    Arch: amd64
1           - Variant:
1           - Feature:
1     # Layers: 3
         layer 1: digest = sha256:000eee12ec04cc914bf96e8f5dee7767510c2aca3816af6078bd9fbe3150920c
         layer 2: digest = sha256:eb22865337de3edb54ec8b52f6c06de320f415e7ec43f01426fdafb8df6d6eb7
         layer 3: digest = sha256:bee5d581ef8bfee2b5a54685813ba6ad9bbe922115d7aef84a21a9dbfcc2d979
```

The great `manifest-tool` gives the most accurate information of them all:

* `Digest` at the beginning is the digest of whatever you passed to it, whether an index or a manifest. In our example, the exact same `50cf` we have seen. 
* references are each of the referred images in an index, if it is an index, and begins with its own digest. In our example, where we included the `amd64` reference in the output (skipped the rest for brevity), that is `189c`, which is exactly what we expect from above.
* `Layers` are the hashes of the layers as referenced on the registry. In our example, these are the exact `000e`, then `eb22`, then `bee5` as expected.

