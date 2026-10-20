---
draft: false
date: 2026-10-19T09:00:00+00:00
title: Introducing TeamSBC Linux
tags: os, image-based, linux, systemd, teamsbc
---

# Introducing TeamSBC Linux

For the past few years I've been working on [bootc](https://bootc.dev/) based operating systems through my work on [image-builder](https://github.com/osbuild/image-builder). Before that I worked on [ostree](https://ostreedev.github.io/ostree/introduction/) based operating systems as well. All of this mostly in the [Fedora Linux](https://fedoraproject.org) corner of distributions.

As you might've gathered from my previous blog post on [Cohesive Linux](../cohesive-linux) I am also quite fond of the [systemd](https://systemd.io) approach to image-based Linux. This approach isn't very common amongst Fedora derivatives (remixes) and thus issues with integrating that approach don't get found as easily.

To this end I decided to combine both my interests in `systemd`, my interests in Fedora, and my interests in affordable Single Board Computers (not that affordable anymore...) into a new and very experimental Fedora Remix called [TeamSBC Linux](https://teamsbc.org).

TeamsBC Linux offers two variants with the primary one being called "Makalu", which is a Cohesive Linux. It uses `systemd-sysupdate`, A/B partitions, system extensions, UKIs, basically everything I've spoken about previously. It uses Fedora packages (plus some of its own where I want things to be slightly different) and packs them up slightly differently.

Over the past year or so I've been building support into `image-builder` for these bits and bobs, done a bunch of upstream and midstream work to get things to be *easier*, and set up a whole lot of YAML to get it all to build.

---

> Here be dragons. This remix is very experimental so don't run it on anything you really care about having uptime. I've ran it for a week or so now on a variety of devices and it has been "OK" but your mileage may vary.

Together this now creates a new Fedora Remix that we can consider highly experimental, but far enough along that I think it's worth a peak and it can take some criticism and/or questions. The goal is to see how far we can integrate the cohesive linux approach into RPM packaging and its assumptions.

If you want to give it a try you can visit the [downloads](https://teamsbc.org/downloads) page. It's probably easiest to get one of the virtual machine images, or you can grab a board-specific image which are currently *only* supplied for the Raspberry Pi 4 and 5 series; though I will likely grow more.

You can chat with me about this on Matrix, I have a room that has only me in it on the [Fedora homeserver](https://matrix.to/#/#teamsbc:fedora.im).

There's still many issues to address, take a look at the [distribution bug tracker](https://github.com/teamsbc/distribution/issues) but perhaps I can get it to a state where it's generally usable *and* offers real benefits around Fedora 46's release.

---

I'll leave this post with some things I've been proud of to get done upstream, these things were *driven* by wanting to make TeamSBC Linux work but hopefully benefit many more, this is an incomplete list:

- [systemd: SELinux alternate root support](https://github.com/systemd/systemd/pull/42768) made it possible to use `systemd-firstboot --image` (or `--root`) on SELinux enabled images without relabeling at boot.
- [systemd: kernel-install: configurable entry name format](https://github.com/systemd/systemd/pull/43290) made it possible to configure the format string used for UKIs (or BLS entries).
- [systemd: dissect: relate verity partitions](https://github.com/systemd/systemd/pull/43896) made it easier to activate a given `/usr` and automatically find its corresponding `dm-verity` partitions.
- [image-builder: user-selectable multiple exports](https://github.com/osbuild/image-builder/pull/2700) made it possible to build sysexts together with an image-build (and export specific partitions).
- [systemd: sysext: IMAGE_ID, IMAGE_VERSION in extension-release](https://github.com/systemd/systemd/pull/43339) made it possible to be much more specific with system extensions.

Much thanks go out to the maintainers of these projects and their packagers (and others I've touched) for dealing with my shenanigans and endless questions.
