---
draft: false
date: 2025-12-25T06:00:00+00:00
title: Installer types for bootable containers
tags: bootc, image-builder
aliases:
  - installer-types-for-bootable-containers.html
---

With [`image-builder`](https://github.com/osbuild/image-builder-cli) (and [`bootc-image-builder`](https://github.com/osbuild/bootc-image-builder)) you can turn bootable containers into installers. We currently have two different types of installers though most users are likely only familiar with the first. Let's go over the two installer types, their differences, their future, and have some examples.

In a [previous post](/weblog/building-interactive-installer-bootc/) I went over how to build an interactive installer for a bootable container. This used the `anaconda-iso` image type. Let's go into how that is built. In this post I'll be using `image-builder` uncontainerized for the `bootc-installer` so my examples are a bit shorter and use `bootc-image-builder` for the `anaconda-iso`. Feel free to use `bootc-image-builder` for both instead. You can [read more](/weblog/merging-bootc-image-builder-and-image-builder) about their differences and future.

The reason for using `bootc-image-builder` for the `anaconda-iso` is that `image-builder` 44 will [remove the `anaconda-iso` type](https://github.com/osbuild/image-builder-cli/pull/401) as we consider the `anaconda-iso` type deprecated.

When `bootc-image-builder` is told to build an `anaconda-iso` image type:

```console
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ mkdir -p output
$ sudo podman run --rm -it --privileged --pull=newer \
    --security-opt label=type:unconfined_t \
    -v ./output:/output \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    quay.io/centos-bootc/bootc-image-builder:latest \
    --type anaconda-iso \
    --use-librepo=True \
    quay.io/centos-bootc/centos-bootc:stream10
```

You end up with an ISO with Anaconda on it that will do an unattended install of the container you passed in. But how do we get there?

For installers like these there are two moving parts. One is the installer environment (which contains [Anaconda](https://github.com/rhinstaller/anaconda) for this image type). The other is the payload to be installed, in this example that's the `centos-bootc:stream10` bootable container.

It was difficult for us to set up the Anaconda environment. Anaconda is not contained in the bootable container so we fell back to the same way we produce `ostree`-based ISOs and set up the Anaconda environment from RPM packages. We inspect the bootable container to find the repository definitions and then download and install the relevant package from there.

While this works it leads to some problems; for example we require [dnf](https://github.com/rpm-software-management/dnf) and its repository configuration to be present in the container which leads to not being able to support non-`dnf`-based distributions or bootable containers without a package manager. It also means we need to *know* the distribution (as found in `/etc/os-release`) so we know which packages to get for the installation environment.

Due to these problems we felt that it is time to iterate on how we do installers for bootable containers. We'd like to not have to care about the package manager being used, nor the distribution contained inside the bootable container.

---

Since our users are already familiar with bootable container concepts we thought: let's make the installer environment be a bootable container as well. This is how the `bootc-installer` image type works.

You can build one:

```
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ sudo image-builder build \
    --bootc-ref quay.io/centos-bootc/centos-bootc:stream10 \
    --bootc-installer-payload-ref quay.io/centos-bootc/centos-bootc:stream10 \
    bootc-installer
```

And you'll see it fail pretty quickly. `image-builder` still has *some* requirements on the container that is being used to build the installer environment. Much like there are *some* requirements on any bootable container to work with `bootc`.

Currently there are no "base installer" bootable containers so we'll have to build our own. Let's go through a `Containerfile` we can use to set up the installer environment:

```Dockerfile
FROM quay.io/centos-bootc/centos-bootc:stream10
RUN dnf install -y \
     anaconda \
     anaconda-install-env-deps \
     anaconda-dracut \
     dracut-config-generic \
     dracut-network \
     net-tools \
     squashfs-tools \
     grub2-efi-x64-cdboot \
     python3-mako \
     lorax-templates-* \
     biosdevname \
     prefixdevname \
     && dnf clean all

# shim-x64 is marked installed but the files are not in the expected
# place for https://github.com/osbuild/osbuild/blob/v160/stages/org.osbuild.grub2.iso#L91, see
# workaround via reinstall, we could add a config to the grub2.iso
# stage to allow a different prefix that then would be used by
# anaconda.
# as https://github.com/osbuild/osbuild/pull/2202 is merged we
# should update images/ to set the correct efi_src_dir and this can
# be removed
RUN dnf reinstall -y shim-x64

# lorax wants to create a symlink in /mnt which points to /var/mnt
# on bootc but /var/mnt does not exist on some images.
#
# If https://gitlab.com/fedora/bootc/base-images/-/merge_requests/294
# gets merged this will be no longer needed
RUN mkdir /var/mnt
```

As you'll see this is not *yet* the nicest thing in the world as we have to work around some idiosyncrasies. When these go away in the future we'll have a much cleaner installation environment `Containerfile`

Let's build that into a container. Remember we have to have the container in container storage for the user that you run `image-builder` as. Usually that would be `root`. Let's store the above `Containerfile` in `Containerfile.my-installer` and build it.

```console
$ sudo podman build -t my-installer -f Containerfile.my-installer
```

Now that we have a container with an installer environment in our container storage we can retry the above command again. Adjusting the reference to point at our newly created installer container. This also shows us that `--bootc-ref` refers to the installer environment and `--bootc-installer-payload-ref` refers to the payload to be installed onto the system by the installer.

```
$ sudo podman pull quay.io/centos-bootc/centos-bootc:stream10
$ sudo image-builder build \
    --bootc-ref localhost/my-installer:latest \
    --bootc-installer-payload-ref quay.io/centos-bootc/centos-bootc:stream10 \
    bootc-installer
```

This time the build completes and we have a container that is functionally the same as the `anaconda-iso` image type except it's built fully from containers. This gives you as the user more control over the installer environment. You could adjust Anaconda configuration, include additional packages, and of course manage your installer environment as a container with all the benefits that brings.

There are limitations to the approach that was taken here. `image-builder` still makes assumptions about the installer environment. Amongst them are that Anaconda is present, and that certain files are at certain places. However, we do feel this is already a nice upgrade to the installer building workflow.

---

We've already started work on making the installer environment and workflow even more agnostic. This is mostly driven by the needs of [Universal Blue](https://universal-blue.org/), [Bazzite](https://bazzite.gg/), and other bootable container based distributions.

We'd like to decouple the environment from requiring Anaconda, and make doing live environments easy.

My colleague [Ondřej Budai](https://budai.cz/) will be presenting our progress on more flexible ISO workflows at [FOSDEM 2026](https://fosdem.org/2026/) so if you're there make sure you come visit and chat with us. I'll likely be around as well, as well as other colleagues.

If you have any input, questions, or ideas, please drop in on our [issue tracker](https://github.com/osbuild/image-builder-cli) or our [Matrix channel](https://matrix.to/#/#image-builder:fedoraproject.org?web-instance%5Belement.io%5D=chat.fedoraproject.org).

---

For this post I'm mostly the messenger. Most of the work I've spoken about in this post was implemented by [Michael Vogt](https://github.com/mvo5).
