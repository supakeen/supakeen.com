---
draft: false
date: 2020-04-06T12:00:00+00:00
title: What is Image Based Linux?
tags: os, image-based
aliases:
  - what-is-image-based-linux.html
---

In operating system land things are always changing. You might have heard about
image based distributions such as [Endless OS](https://www.endlessos.org), [Fedora IoT](https://fedoraproject.org/iot/),
[RHEL for Edge](https://www.redhat.com/en/technologies/linux-platforms/enterprise-linux/edge-computing), [Ubuntu Core](https://ubuntu.com/core) and others. Perhaps you've heard other terms
such as [Unified Kernel Images](https://uapi-group.org/specifications/specs/unified_kernel_image/) (UKI) or [Discoverable Disk Images](https://uapi-group.org/specifications/specs/discoverable_disk_image/) (DDI) and
wondered what that's about. Let's get into what image based Linux is and what its
possible benefits are to you.

This post is the introduction part to a series I'm writing about image based Linux
systems, their advantages and disadvantages and how to build and test them.

Before we get started on the difference let me describe a more traditional
package based Linux distribution. You install it from an ISO and update packages
with that distributions package manager. If you have a lot of machines it might
be difficult to track which state they are in. They all have different packages
installed and are perhaps in different states regarding the updates they've
installed and when they've last rebooted to use a new kernel.

With an image based distribution the system as a whole is built and updated as
a single thing. A system being on a certain version means you know the state of
everything on the system.

Image based distributions are useful in situations where you run many machines
in many locations, often called the 'device edge'. Think of a factory full of
single board computers that collect sensor data. Point-of-sale systems.
Information screens. Cars. Places where it isn't immediately feasible to send
an engineer if something breaks. Image based distributions are also finding
their place in the mobile market.

Running a normal distribution on these sorts of systems makes it hard to test
the system before deploying updates. Image based distributions often support a
form of A/B booting and monitored boot. This means that they can roll back from
a broken update to a previously known good image.

Image based distributions often have a focus on providing easy provisioning of
devices through things such as [FDO](https://fidoalliance.org/specs/FDO/FIDO-Device-Onboard-RD-v1.0-20201202.html) or [Ignition](https://coreos.github.io/ignition/). Try to provide a secure
physical environment through dm/fsverity, secure boot, TPMs, and UKI. And a
solid update cycle that can always restore to a previous state. Don't worry if
you're not yet familiar with this forest of abbreviations and concepts, these will
be the subjects of followup posts.

In the next post I'll talk about the ways disk contents are managed in image based
distributions. This is one of the important considerations that sets apart an image
based distribution from a traditional one.
