---
draft: false
date: 2026-09-25T09:00:00+00:00
title: Cohesive Linux
tags: os, image-based, linux, systemd
---

# Cohesive Linux

Over the past few years I've been in contact with and working with ["image-based" Linux systems](https://supakeen.com/weblog/image-based-linux) and grown fond of the ideas and possibilities they offer while being wary of the things they can take away from tinkerers and hackers (in the [correct sense of the word](https://en.wikipedia.org/wiki/Hacker#Programmer_and_free-software_hackers)). My work has mostly been on the Fedora side of things and the implementations of "image-based" Linux there through [ostree](https://ostreedev.github.io/ostree/introduction/) and nowadays [bootc](https://bootc.dev/). Lately I've also been interested in the [systemd](https://github.com/systemd/systemd) approach described by Lennart Poettering in his [Fitting Everything Together](https://0pointer.net/blog/fitting-everything-together.html) blog.

All of these systems have been described as "immutable", "atomic", "image-based", and those terms have become very overloaded, wide, and often don't even fit. They're applied to a huge set of implementations ([Fedora Atomic]() and its derivatives, [NixOS](), ChromeOS, [openSUSE's MicroOS](), [ParticleOS](), and many others) all of which use very different ways of achieving "image-based" Linux. The words are useful but non-specific and prescribe behavior that many of these systems do not hold, or hold partially.

Lately I've slowly started using different terminology to describe the `systemd` approach to these forms of systems: **cohesive**. A cohesive system is often "immutable", "atomic", and is definitely "image-based" but these properties don't describe the technical implementation behind it. Which is a bunch of parts that work together and are kept together by shared specifications and discoverability instead of a top-down description (e.g. an `ostree` commit, a `bootc` reference, a `nix` deriviation graph, `transactional-update` snapshot). The word has started growing on me and maybe it fits your brain as well (or it doesn't).

---

The building blocks for a cohesive linux are the [Discoverable Partitions Specification](https://uapi-group.org/specifications/specs/discoverable_partitions_specification/), [Discoverable Disk Images](https://uapi-group.org/specifications/specs/discoverable_disk_image/), [Unified Kernel Images](https://uapi-group.org/specifications/specs/unified_kernel_image/), `dm-verity`, and an update mechanism able to fetch these parts (`systemd-sysupdate`). These are specifications that are currently most completely implemented by `systemd` and its bazillion parts but are by no means exclusive (`bli` in `grub2`, etc).

In practice that looks (somewhat) like this (starting from the kernel) for the boot of a system. The `UKI` here defines what it wants and the parts are found through automatic discovery:

* A bootloader (or firmware directly) finds and loads a trusted `UKI`.
* The `initrd` contained in the `UKI` uses `udev`'s built-in `dissect-image` to find the appropriate partitions that have what the `UKI` requires. These are then mounted appropriately by `systemd-gpt-auto-generator`. [GPT](https://en.wikipedia.org/wiki/GUID_Partition_Table) partitions are identified by their type UUIDs so we know where to find `/usr`, `/`, or others for a given architecture. Their related verity hash and verity signature partitions can be found in the same way and verified.

  Identifying the partitions that a UKI wants is done through kernel command-line arguments embedded in it; you'd either use `usrhash` for a cryptographic coupling based on `dm-verity` or use `systemd.image_filter` to match on (for example) a label.
* Potentially necessary partitions are created, removed, or emptied as need be by `systemd-repart`. For example in case of a factory reset request or because the initial image shipped without a `/` partition and only a `/usr`.

Rollbacks of the system work because the cohesion is per-`UKI`. Booting an older `UKI` finds the `/usr` partition (or other bits) that conforms to what it wants.

With everything mounted we can transition into the root for the actual system. The parts required by the `UKI` to boot the system are found based on their shared implementation of the standards not because the `UKI` encodes them directly.

On the booted system we can go further with things with systemd extensions. These are overlays that can be activated on systems that match their `extension-release.d` specifiers making them only activate on systems that share the same specifiers.

We can use `systemd-sysupdate` to fetch all of these components based on their shared specifiers and place them into the right places where things can look for them.

> Note I've been writing a lot of UKI here, but this also works with [Bootloader Specification](https://uapi-group.org/specifications/specs/boot_loader_specification/) snippets (type 1) with slightly different guarantees.

---

With the above base explanation written down, how does that compare to say `bootc`, `ostree`, `transactional-update`, or other various tools that implement some form of "atomic", or "image-based" linux?

The reality is that the end *result* doesn't differ much, they achieve the same style of managing systems in different ways. However, cohesive linux and the systemd approach allow for easier swapping of parts which is for me hugely beneficial while testing (I can rebuild a single part and put it into my image to test while still having all the guarantees). There isn't value judgement in this post, and in fact things here can also be used to enhance the other approaches.

I will leave with what I think is the current strongest suit of cohesive linux: it's *easy* to build your own things around it because its specification based. The main implementer is `systemd`, but `image-builder` (which I tend to work on) has support for many of these things to be able to build cohesive linux systems with more coming down the line.

The `systemd` tooling of course works together with all of this by allowing me to update a downloaded image offline based on its `sysupdate` snippets, recreate or create its partitions with `repart`, provision it with `firstboot`, mount and inspect with `dissect`, and so much more (food for another post at some time). All these subtools understand the specifications and thus can all work with and together to develop, maintain, and adjust images.

---

P.S. I've been very slowly working on an experimental Fedora remix that fits all of this together and specifically targets Single Board Computers (because I really like those). This has caused a bunch of my work to shift to better support cohesive linux in Fedora packaging and related build tooling. I'll write about that, what all that entails (lots of SELinux...), and what's missing hopefully soon as well.
