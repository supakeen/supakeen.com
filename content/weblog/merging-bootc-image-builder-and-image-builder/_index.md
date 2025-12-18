---
draft: false
date: 2025-12-17T12:00:00+00:00
title: Merging bootc-image-builder and image-builder
tags: bootc, image-builder, lore
aliases:
  - merging-bootc-image-builder-and-image-builder.html
---

When [we](https://github.com/osbuild) started working on building images from [`bootc`-containers](https://github.com/bootc-dev/bootc) we created a tool called [bootc-image-builder](https://github.com/osbuild/bootc-image-builder). This is a tool that is shipped as a container to users that can turn `bootc`-containers into disk images or installers for deployment to virtual machines or bare metal.

Having this in a separate executable allowed us to iterate faster without having to integrate (as much) into the libraries we use in our other tooling such as `image-builder`, and `osbuild-composer`. It also allowed us to test having an executable that did not rely on daemons and sockets like we did with `osbuild-composer`.

As time went on we wanted to apply the no-daemons approach to our other executables as well. Thus [image-builder](https://github.com/osbuild/image-builder-cli) was born and is geared up to eventually replace usage of `osbuild-composer`. `image-builder`, like `bootc-image-builder` is available as a container or can be installed through normal package management on Fedora, CentOS, and RHEL (starting in 9.7 and 10.1). `bootc` support in `image-builder` is currently only available on Fedora and CentOS but will be available in RHEL 9.8 and 10.2.

Over the past month or two we've largely merged the two different executables into `image-builder`. This means that `image-builder` itself now has all the functionality that `bootc-image-builder` also has. We've also made it so that `bootc-image-builder` can be [multi-call binary](https://github.com/osbuild/image-builder-cli/pull/374) to `image-builder` itself.

We still have a bunch of work to do before this merge is complete and you can keep using `bootc-image-builder` as your entrypoint to building `bootc`-containers into artifacts. Eventually we'll switch over to using the multi-call binary by default for `bootc-image-builder` and then at some point in the further future it is likely that we will only offer `image-builder`.

This is a good time to take a look at using `image-builder` directly, we have some [documentation](https://osbuild.org/docs/developer-guide/projects/image-builder/usage/#bootc) describing its command line and functionality. While the arguments are different the functionality stays the same.

For a quick reference let's go over some common `bootc-image-builder` commands mapped onto `image-builder`.

---

Building a `bootc`-container into a qcow2.

```console
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ mkdir output
$ sudo podman run \
    --rm \
    -it \
    --privileged \
    --pull=newer \
    --security-opt label=type:unconfined_t \
    -v ./config.toml:/config.toml:ro \
    -v ./output:/output \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    quay.io/centos-bootc/bootc-image-builder:latest \
    --type qcow2 \
    --use-librepo=True \
    quay.io/centos-bootc/centos-bootc:stream10
```

Would become:

```shell
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ sudo image-builder build --bootc-ref quay.io/centos-bootc/centos-bootc:stream10 qcow2
```

Building a `bootc`-container into an Anaconda installer.

```
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ mkdir output
$ sudo podman run \
    --rm \
    -it \
    --privileged \
    --pull=newer \
    --security-opt label=type:unconfined_t \
    -v ./config.toml:/config.toml:ro \
    -v ./output:/output \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    quay.io/centos-bootc/bootc-image-builder:latest \
    --type anaconda-iso \
    --use-librepo=True \
    quay.io/centos-bootc/centos-bootc:stream10
```

Would become:

```shell
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ sudo image-builder build --bootc-ref quay.io/centos-bootc/centos-bootc:stream10 anaconda-iso
```

The above examples assume that you've done a `dnf install image-builder`. You can also use the same container-based workflow with `image-builder`. Here's an example of using the `image-builder` upstream container to build a `bootc`-container into a qcow2 disk image:

```shell
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ sudo podman run \
    --rm \
    -it \
    --privileged \
    --pull=newer \
    --security-opt label=type:unconfined_t \
    -v ./output:/output \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    ghcr.io/osbuild/image-builder-cli:latest \
    build \
    --output-dir /output \
    --bootc-ref quay.io/centos-bootc/centos-bootc:stream10 \
    qcow2
```
