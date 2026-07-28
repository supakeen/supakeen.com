---
draft: false
date: 2026-07-24T06:00:00+00:00
title: The Fedora 45 Sausage Factory
tags: fedora, image-builder, kiwi, pungi, koji, bodhi, openqa
---

# The Fedora 45 Sausage Factory

This is a walkthrough of how Fedora turns source code and packages into the
artifacts you download and install. It follows the a package from a packager's
git push to a composed release: ISOs, cloud images, container images, and OSTree
deployments.

The walkthrough describes how the Fedora 'sausage' is created as of Fedora 45,
things change all the time; I hope to have time to update this document every
cycle or every few cycles of Fedora releases so there's both history and people
can find up to date information.

In fact this document is currently still living as Fedora 45 has not been released
yet so this post likely will get updated until the release. There's a few change
proposals in flight that change some of the bits and bobs here. Specifically how
the `boot.iso` is produced which is a decently large part of this post.

Thus now is also the right time to let me know if I got anything *wrong*, please
e-mail me at `cmdr@supakeen.com` if you find any errors.

---

## Starting Line: dist-git

Things starts with a packager pushing a commit to a package.

Fedora stores the source definitions of every package in individual Git
repositories at [src.fedoraproject.org](https://src.fedoraproject.org). Each
repo contains an RPM spec file, any downstream patches, and a `sources` file
that points to upstream tarballs stored in a separate lookaside cache. The
large binary files stay out of Git; the other files exist inside of Git and
thus have full version control.

Packagers usually interact with these repos through `fedpkg`, a CLI that wraps
the common operations: cloning repos, uploading source tarballs, submitting builds,
and creating updates. The important thing about `fedpkg build` is what it
actually does: it constructs a URL pointing at a specific commit in the Git
repo and hands that to Koji, the build system. The build is fully reproducible
from that commit hash.

Branches map to releases: `rawhide` for development, `f44` for Fedora 44.
The hosting is split between Pagure (src.fedoraproject.org) and Forgejo
([forge.fedoraproject.org](https://forge.fedoraproject.org)), with an ongoing
migration from the former to the latter.

Packagers can do work without using `fedpkg`, but for the purposes of this
post it's easier to assume `fedpkg` is in use since it abstract away some of
the bits that are irrelevant to the reader.

## Building Packages: Koji

When `fedpkg build` submits that Git URL, Koji takes over. Koji is Fedora's
build system. It has been around since Fedora 7, and it builds essentially
everything.

Koji follows a hub-and-spoke architecture. The hub is a passive XML-RPC server
that sits in front of a PostgreSQL database. Builder daemons poll the hub for
work, create a fresh [Mock](https://github.com/rpm-software-management/mock)
chroot environment for each build, run the build, and upload the results. Every
build starts from a clean room. You can never get a different result because
someone installed something on the builder last week.

The organizational model is built around **tags**. A tag is a named collection
of builds. Build targets map an incoming build request to two tags: a build tag
(which defines the buildroot, the packages available during the build) and a
destination tag (where the finished build lands).

Koji doesn't just build RPMs. Through its plugin system and content generators,
it also orchestrates image builds: Kiwi images via the `kiwiBuild` task type,
Image Builder artifacts via `imageBuilderBuild`, and OSTree composes via
runroot tasks. We'll get to those.

## Gating Updates: Bodhi

A fresh RPM build sitting in Koji doesn't automatically reach users. It goes
through [Bodhi](https://bodhi.fedoraproject.org/), Fedora's update management
system.

Bodhi gates the release of updates through a feedback and testing cycle. A
packager submits an update containing one or more builds. That update goes
through a sequence of states: pending, testing, stable. Users and automated
tests provide karma (+1 / -1). When an update hits +3 karma or spends enough
days in testing, it will be automatically pushed to stable. If an update hits
+1 karma the maintainer can manually push to stable (+2 for critical path).

Maintainers can configure (part of) the settings for automatic pushing of updates
and karma limits.

Behind the scenes, Bodhi manages all of this through Koji tags. When an update
moves from testing to stable, Bodhi moves the build from the
`f44-updates-testing` tag to the `f44-updates` tag. It then invokes Pungi to
compose the update repository, the actual yum/dnf repo that users pull from
when they run `dnf upgrade`.

Critical path packages, the ones your system needs to boot and function, get
stricter requirements: 14 days in testing instead of 7, and more karma needed.
Bodhi also integrates with Greenwave and ResultsDB for CI test gating, so
automated test failures can block an update from reaching stable.

For Rawhide, these processes are configured differently. Non-critical path
packages go stable immediately unless there's a gating policy (either because
the package is in the critical path or because the package itself has one) it
still applies and an update can be held back until things pass.

## Composing a Release: Pungi

Individual RPMs, even with Bodhi's gating, are just packages. Turning them into
something you can download and install, an ISO, a cloud image, a repository,
is the job of [Pungi](https://forge.fedoraproject.org/pungi/pungi).

Pungi is the compose orchestrator. It doesn't do much of the heavy lifting
itself. Instead, it coordinates the tools that do, ensuring everything is built
from the same consistent set of packages.

The name is a reference to the [pungi](https://en.wikipedia.org/wiki/Pungi),
a snake-charming instrument. It "charms" [Anaconda](https://fedoraproject.org/wiki/Anaconda),
Fedora's installer. The pun has survived since 2006.

A compose starts when `pungi-koji` runs, either triggered by cron for nightly
Rawhide composes, or manually for milestone releases. It loads a configuration
file (for Fedora, that's `fedora.conf` in the
[pungi-fedora](https://forge.fedoraproject.org/releng/pungi-fedora) repository)
and runs through a sequence of phases.

### Freezing the Package Set

The first real work Pungi does is snapshot the set of packages from a Koji tag.
This is the **Pkgset** phase, and it's critical: every subsequent phase works
from this frozen set. If someone submits a new build to Koji while the compose
is running, it won't sneak into the compose. The tag-based snapshot makes the
entire compose auditable. You can always determine exactly which versions of
which packages ended up in any given compose.

### Deciding What Goes Where: comps and variants

With the packages frozen, Pungi needs to figure out which packages belong in
which product. This is where two XML inputs come in.

**comps** (`fedora-comps`) defines package groups, collections of related
packages like `gnome-desktop`, `server-product`, or `core`. Each package in a
group has a type: mandatory, default, optional, or conditional. This is the
same grouping system that powers `dnf group install` and Anaconda's Software
Selection screen.

**variants XML** (`variants-fedora.xml` in pungi-fedora) defines the products
in a compose: Everything, Server, Workstation, KDE, Silverblue, and so on.
Each variant lists which comps groups it includes and which architectures it
supports. Most variants in modern Fedora are `is_empty="true"`, meaning they
produce only images (via Kiwi, Image Builder, or OSTree) using packages from
the Everything variant's repository. Only Everything and Server actually run
the package gathering pipeline.

### Building Boot Images

The **Buildinstall** phase runs [lorax](https://weldr.io/lorax/) to create
`boot.iso`, the image that boots the Anaconda installer. Lorax installs a
minimal package set into a temporary root, configures the Anaconda environment,
compresses it into `install.img`, and wraps it with a bootloader into an ISO.
One lorax task runs per variant-architecture combination, typically in a Koji
runroot environment for isolation.

The resulting boot.iso is the foundation for installer ISOs. It doesn't contain
the packages themselves. Those get layered on top in the next step to produce
a `dvd.iso`.

### Producing ISOs

The **Createiso** phase takes the boot.iso from Buildinstall, overlays packages
and repository metadata on top of it using xorriso, and produces the final
installer ISOs. For Fedora's production compose, this means primarily the
Server dvd.iso. Most other variants get their media from Kiwi or Image Builder
instead.

### Building Images with Kiwi

[Kiwi](https://osinside.github.io/kiwi/) is the primary image-building tool
in Fedora's compose. It builds cloud images (AWS, Azure, GCP, generic), Vagrant
boxes, container base images, WSL images, and live ISOs for most desktop spins.

Kiwi was originally created at SUSE around 2005 and is a two-step builder:
prepare a root filesystem by installing packages, then package it into the
target format. Image definitions are XML files validated against a RELAX NG
schema. Fedora's definitions live in the
[fedora-kiwi-descriptions](https://forge.fedoraproject.org/releng/fedora-kiwi-descriptions)
repository, organized under teams: `teams/cloud/`, `teams/kde/`,
`teams/spins/`, `teams/wsl/`.

Kiwi integrates with Koji through a built-in plugin providing the `kiwiBuild`
task type. The parent task checks out the image description from SCM, parses
the XML, and spawns one subtask per architecture. Each subtask creates a Mock
buildroot, replaces all repository definitions in the XML with Koji-controlled
repos (so the build uses the compose's frozen package set), runs
`kiwi-ng system build`, and uploads the results.

### Building Images with Image Builder

[Image Builder](https://github.com/osbuild/image-builder) handles the
ostree-based and bootc-based artifacts: Atomic Desktop ISOs, Fedora IoT,
Fedora Minimal, and boot.iso.

Under the hood, Image Builder is built on top of
[osbuild](https://github.com/osbuild/osbuild), a pipeline-based build engine
that processes JSON manifests. Each manifest describes a directed acyclic graph
of pipelines, and each pipeline runs sequential stages, small executables that
do one thing (install packages, configure a bootloader, create a partition
table, embed an OSTree commit). There are currently 176 stages. Stages run
in bubblewrap-sandboxed environments with filesystem and network isolation.

The `image-builder` CLI is stateless: it reads distribution definitions, resolves
package dependencies, generates an osbuild manifest, and runs osbuild directly.

Image Builder integrates with Koji through
[koji-image-builder](https://github.com/osbuild/koji-image-builder), a
separate Koji plugin providing the `imageBuilderBuild` task type. Each build
runs `image-builder` inside a Mock buildroot on a Koji worker.

### Building Atomic Desktops: rpm-ostree

Fedora Silverblue, Kinoite, Sway Atomic, Budgie Atomic, and COSMIC Atomic are
all built differently from traditional package-based variants. Instead of
shipping a pile of RPMs for the user to install, these variants ship a
versioned, checksummed filesystem tree, an OSTree commit, that the client
replicates atomically.

[rpm-ostree](https://github.com/coreos/rpm-ostree) is the tool that makes this
work. It's a hybrid image/package system, combining
[libostree](https://ostreedev.github.io/ostree/) (atomic OS tree deployment)
with RPM (package management via libdnf).

On the compose side, `rpm-ostree compose tree` takes a treefile (a YAML
specification listing packages, configuration, and post-processing scripts) and
a set of RPM repositories, and produces an OSTree commit. Pungi's OSTree phase
invokes this command in a Koji runroot environment for each Atomic Desktop
variant.

The treefiles live in the
[workstation-ostree-config](https://forge.fedoraproject.org/atomic-desktops/config)
repository. They use a layered include system:
`silverblue-ostree.yaml` includes `silverblue-common.yaml`, which includes
`common.yaml`, which conditionally includes `fedora.yaml` and architecture-
specific files. The `comps-sync.py` script bridges the gap between comps
groups and treefile package lists. It parses the comps XML, extracts the
default and mandatory packages from groups like `gnome-desktop` and
`workstation-product`, and writes them into generated YAML files. This way,
the desktop environment packages stay in sync with what comps defines, while
the treefile adds ostree-specific packages (bootupd, composefs) and applies
exclusions for packages that are Flatpak'd or incompatible with ostree.

### Compose Metadata: productmd

Every compose produces a set of metadata files that describe what's in it.
These are defined by [productmd](https://github.com/release-engineering/productmd),
a Python library that standardizes the format.

The key files are:
- `composeinfo.json`, release identity, variant definitions, filesystem paths
- `images.json`, every image with its type, format, checksums, and size
- `rpms.json`, every RPM organized by variant and architecture
- `.treeinfo`, per-variant metadata consumed by Anaconda to find installation
  trees

Pungi creates these as it goes. At the start, it generates a compose ID (e.g.,
`Fedora-Rawhide-20260723.n.0`). Each image-producing phase adds Image objects.
The Gather phase populates the RPMs manifest. At finalization, everything gets
written out.

Downstream consumers rely on these files heavily. Anaconda reads `.treeinfo`
to discover where to find packages and boot images during installation. Bodhi
inspects compose metadata for update management. openQA reads `images.json`
to locate images for testing. Mirror management tools query the metadata to
decide what to sync.

### Testing: openQA

Once a compose finishes, [openQA](https://openqa.fedoraproject.org/) picks it
up for automated testing. openQA is an automated test system for operating
systems, originally developed by SUSE and adapted for Fedora by [Adam
Williamson](https://www.happyassassin.net/) and the Fedora QA team.

openQA boots the compose's images in virtual machines and runs through test
scenarios: installation workflows (interactive, kickstart, various disk
layouts), desktop functionality, upgrade paths, and basic system validation.
Tests are written as a combination of test scripts ("needles" for visual
matching and Perl test modules for interaction) and are maintained in the
[os-autoinst-distri-fedora](https://forge.fedoraproject.org/quality/os-autoinst-distri-fedora)
repository.

Every nightly Rawhide compose and every milestone compose (Beta, Final)
triggers an openQA run. Results feed into the release validation process,
and blocking bugs found by openQA can hold a release. The system also tests
update composes from Bodhi, catching regressions before they reach users.

## The Sausage

Here's how the phases flow in a typical Fedora compose:

```
  pungi-koji
    │
    ├── Step 1: Init
    │     Parse comps + variants XML
    │
    ├── Step 2: Pkgset
    │     Snapshot packages from Koji tag
    │
    ├── Step 3: Essentials (parallel)
    │     ├── Buildinstall (lorax → boot.iso)
    │     ├── Gather + Createrepo (package repos)
    │     └── OSTree (rpm-ostree compose tree)
    │
    ├── Step 4: Images (parallel)
    │     ├── Createiso (dvd.iso from boot.iso + packages)
    │     ├── KiwiBuild (cloud, live, containers, WSL)
    │     └── ImageBuilder (Atomic Desktop ISOs, IoT, Minimal)
    │
    │   (OstreeContainer runs in background through steps 3-4)
    │
    ├── Step 5: Checksums
    │     Write productmd metadata, SHA256 checksums
    │
    └── Step 6: Test
          Run repoclosure, verify compose integrity
```

The barriers between steps matter. Step 3 must complete before step 4 starts,
because Createiso needs boot.iso, the image phases need package repos, and
ImageBuilder needs OSTree commits.

## Governance: Changes Process

Major modifications to Fedora, whether new tools, default changes, or mass
rebuilds, go through the [Changes process](https://docs.fedoraproject.org/en-US/program_management/changes_policy/).
This is the governance mechanism that tracks proposals like "use Kiwi for cloud
images" or "switch boot.iso to Image Builder."

Changes come in two flavors. **System-Wide** changes affect critical path
packages or system defaults and need detailed proposals, contingency plans,
test plans, and FESCo approval. **Self-Contained** changes are scoped to
packages the proposer owns and get lighter review but still need FESco
approval.

The lifecycle is: draft a wiki page, submit it, get reviewed by the Change
Wrangler, get voted on by FESCo (the 9-member elected engineering steering
committee), implement in Rawhide, and meet completion checkpoints tied to the
release schedule. If a change doesn't meet completion checkpoints it usually
gets deferred to the next Fedora release.
