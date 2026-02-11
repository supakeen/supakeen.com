---
draft: false
date: 2026-02-10T06:00:00+00:00
title: Bootable containers on the Raspberry Pi
tags: bootc, image-builder, raspberry-pi
aliases:
  - bootc-on-the-raspberry-pi.html
---

The question on how to use [bootable containers]() on the Raspberry Pi (or other single board computers that have 'interesting' boot setups) comes up quite regularly. For a while now it has been possible in [image-builder](https://github.com/osbuild/image-builder-cli) to create a Raspberry Pi compatible partition table. I don't think this has been written down in long-form anywhere except answers to GitHub issues and Matrix chats. Let's write it down in a longer format and go over the specifics and what will make it easier in the future.

When talking about the Raspberry Pi in this case I specifically mean the Raspberry Pi 3 (`rpi3`), and the Raspberry Pi Zero 2 W (`rpi02w`). These both run the same firmware which is stored in ROM and cannot be updated. The Raspberry Pi 4 (`rpi4`) and the Raspberry Pi 5 (`rpi5`) (more on that later) are easier to handle as they have updateable EEPROM firmware and later versions of this firmware allow more variety in the partition tables and disk images they can boot from.

---

This post describes how things work in general for the Raspberry Pi's but many things are applicable to other single board computers. I also assume that you want a Fedora-alike solution. Other distributions might do things slightly differently.

Fedora images for `aarch64` are made to boot on a variety of devices; they are not device specific. Fedora makes its images [SystemReady](https://developer.arm.com/documentation/107981/latest/) compliant and boots using EFI. Where there is no EFI environment Fedora uses [u-boot](https://u-boot.org/) to provide an EFI environment for the rest of the boot.

---

For the `rpi3` and `rpi02w` we must use a disk image that is partitioned using the Master Boot Record (MBR) format (also often referred to as DOS-partitioning). This is because the on-ROM firmware for these devices only understands this format and not the [GUID Partition Table](https://en.wikipedia.org/wiki/GUID_Partition_Table) (GPT) format.

We will also need to put some specific files into the partition that the device will boot from. Let's start with a `Containerfile` that creates a container with the necessary packages and workarounds. 

These workarounds to move files around might become unnecessary, or change, after `bootupd` gains the ability to do this for us and the packages providing these files are updated. You can [follow this upstream ticket](https://github.com/coreos/bootupd/issues/766). This workaround was created by [Ondrej Budai](https://budai.cz/) and can be found [on GitHub](https://github.com/ondrejbudai/fedora-bootc-raspi/) as well. The relevant packages install directly into `/boot` hence we install the packages and then copy the relevant contents into a new directory in `/usr/lib`. Afterwards we clean up the packages and files from `/boot`.

```Dockerfile
FROM quay.io/fedora/fedora-bootc:44

RUN dnf install -qy \
      bcm283x-firmware \
      uboot-images-armv8

RUN mkdir -p /usr/lib/bootc-rpi-firmware

# `bcm283x-firmware` installs into `/boot/efi` directly, let's move
# things into `/usr` so `bootupd` can later be adjusted to copy them
# for us.
RUN cp -rvP /boot/efi/*.dtb \
            /boot/efi/*.bin \
            /boot/efi/*.txt \
            /boot/efi/*.dat \
            /boot/efi/*.elf \
            /boot/efi/overlays \
            /usr/lib/bootc-rpi-firmware/

# u-boot needs a different name
RUN cp -vP /usr/share/uboot/rpi_arm64/u-boot.bin \
           /usr/lib/bootc-rpi-firmware/rpi-u-boot.bin
           
# clean up
RUN rm -rv /boot/efi && \
    dnf remove -qy \
      bcm283x-firmware \
      uboot-images-armv8

RUN dnf clean all && \
    rm -rv /var/lib/dnf && \
    rm -rv /var/cache/libdnf5 && \
    rm -v /var/log/dnf5.log

# make room for a wrapper around bootupd. this wrapper makes
# makes sure we copy our additional firmware and config files.
# this part might become obsolete (or replaced with a config
# file) after: https://github.com/coreos/bootupd/issues/766
# is addressed.
RUN mkdir -p /usr/bin/bootupctl-original && \
    mv /usr/bin/bootupctl /usr/bin/bootupctl-original

RUN cat <<'EOF' > /usr/bin/bootupctl
#!/bin/bash
if [[ $# -ge 2 ]]; then
    if [[ "$1" == "backend" && "$2" == "install" ]]; then
        # BASH_ARGV[0] is the last argument. In the case of bootupctl backend install, it's the path to the target
        # root directory.
        echo "Copying Raspberry Pi firmware files to ${BASH_ARGV[0]}/boot/efi/" >&2
        cp -av /usr/lib/bootc-rpi-firmware/. "${BASH_ARGV[0]}"/boot/efi/
        echo "Copying Raspberry Pi firmware files finished" >&2
    fi
fi
exec /usr/bin/bootupctl-original/bootupctl "$@"
EOF

RUN chmod +x /usr/bin/bootupctl

RUN bootc container lint
```

We can then build this container. We build the container as root so that the container ends up in the container storage for root. This is important since we need to run `image-builder` as root and it's going to look inside the container storage of the user it runs as.

At this point it's probably good to think about how we want to log in to the machine. We can add a user in the container (but that would mean it exists in all installations of this container). We can include one of the provisioning tools (`initial-setup`, for example) to set up a user on first boot; or you could add [a user to the blueprint](https://osbuild.org/docs/user-guide/blueprint-reference/#additional-users). The latter options are per-boot, and per disk image respectively.

Note that this Containerfile has to be built on `aarch64` as the required packages are only available on `aarch64`. You can attempt a cross-arch build if you have no `aarch64` hardware available by using `podman build --platform linux/arm64` which will be quite slow, but better than nothing.

```console
$ sudo podman build \
    -t localhost/fedora-for-rpi3 \
    -f Containerfile
```

Then turn it into a disk image using `image-builder`. We will use the following blueprint to set up the partition tables in a way that the `rpi3` and `rpi02w` understand. You can experiment with the sizes of the partitions but these are the 'default' Fedora sizes for the mountpoints.

```toml
[customizations.disk]
type = "dos"

[[customizations.disk.partitions]]
mountpoint = "/boot/efi"
minsize = "500 MiB"
fs_type = "vfat"
part_type = "06"

[[customizations.disk.partitions]]
mountpoint = "/boot"
minsize = "2000 MiB"
fs_type = "ext4"

[[customizations.disk.partitions]]
mountpoint = "/"
minsize = "3500 MiB"
fs_type = "ext4"

# root password if you want it
# [[customizations.user]]
# name = "root"
# password = "root"  # don't use this ;)
```

With that in place we can start the build:

```console
$ sudo image-builder build \
    --bootc-ref localhost/fedora-for-rpi3 \
    --bootc-default-fs ext4 \
    --blueprint blueprint.toml \
    raw
```

We'll end up with a ready made raw image that can be `dd`'ed straight onto an SD card. Alternatively you can use `arm-image-installer` to also resize the filesystem to your micro SD card's size. Note that for `arm-image-installer` you'll need to compress your image with `xz`.

```console
$ sudo dd \
    if=bootc-fedora-44-raw-aarch64/bootc-fedora-44-raw-aarch64.raw
    of=/path/to/your/sd/card
```

or:

```console
$ xz bootc-fedora-44-raw-aarch64.raw
$ sudo arm-image-installer \
    --image bootc-fedora-44-raw-aarch64.raw.xz \
    --media /path/to/your/sd/card \
    --resizefs
```

After which we can boot the system and end up with:

```
[root@fedora ~]# hostnamectl
  Transient hostname: fedora
     Static hostname: (unset)                                
           Icon name: computer
          Machine ID: 42ede2152cc2439599472df6a033ddc6
             Boot ID: eb6ac23f6b424134ad0ae8a1c0618187
        Product UUID: 30303030-3030-3030-3735-306363613200
    Operating System: Fedora Linux 44 (Forty Four Prerelease)
         CPE OS Name: cpe:/o:fedoraproject:fedora:44
      OS Support End: Wed 2027-05-19
OS Support Remaining: 1y 3month 1w 2d                        
              Kernel: Linux 6.19.0-59.fc44.aarch64
        Architecture: arm64
     Hardware Vendor: raspberrypi
      Hardware Model: Raspberry Pi 3 Model B Plus Rev 1.4
     Hardware Serial: 00000000750cca2c
    Firmware Version: 2026.04-rc1
       Firmware Date: Wed 2026-04-01
[root@fedora ~]# bootc status
[   53.499890] SELinux:  Context unconfined_u:object_r:invalid_bootcinstall_testlabel_t:s0 is not valid (left unmapped).
● Booted image: localhost/fedora-for-rpi3
        Digest: sha256:f4ae4184f879820ef9b539074a243dde0325508c0e4ce7c5c1ef1c073480fed6 (arm64)
       Version: 44.20260210.0 (2026-02-11T10:48:10Z)
```

If you're having issues with your system coming up; locking up during boot try editing your kernel cmdline in grub to include `modprobe.blacklist=vc4`. If that solves the issue then [add a kernel argument](https://bootc-dev.github.io/bootc//building/kernel-arguments.html#usrlibbootckargsd) to your `Containerfile` and rebuild the container and disk image.

I've noticed that *some* of the Pi 3's and Pi Zero 2 W's I've tested this on had this issue and some did not.

---

For the `rpi4` the firmware that ships from the factory can only read the MBR-partitioned disks. If the firmware has been updated then it can also read GPT-partitioned disks. For the former case you can follow the same blueprint we used for the `rpi3` and `rpi02w`. If you want a GPT partitioned image you can build without a blueprint and you will get a partition table that works.

Here's the output from a direct boot of the image produced in the previous step.

```
[root@fedora ~]# hostnamectl
  Transient hostname: fedora
     Static hostname: (unset)                                
           Icon name: computer
          Machine ID: 42ede2152cc2439599472df6a033ddc6
             Boot ID: 9ac710b5f36d4f9691a0eeec5c634baf
        Product UUID: 30303031-3030-3030-6663-333437313000
    Operating System: Fedora Linux 44 (Forty Four Prerelease)
         CPE OS Name: cpe:/o:fedoraproject:fedora:44
      OS Support End: Wed 2027-05-19
OS Support Remaining: 1y 3month 1w 2d                        
              Kernel: Linux 6.19.0-59.fc44.aarch64
        Architecture: arm64
     Hardware Vendor: raspberrypi
      Hardware Model: Raspberry Pi 4 Model B Rev 1.2
     Hardware Serial: 10000000fc347103
    Firmware Version: 2026.04-rc1
       Firmware Date: Wed 2026-04-01
[root@fedora ~]# bootc status
[   55.455803] SELinux:  Context unconfined_u:object_r:invalid_bootcinstall_testlabel_t:s0 is not valid (left unmapped).
● Booted image: localhost/fedora-for-rpi3
        Digest: sha256:f4ae4184f879820ef9b539074a243dde0325508c0e4ce7c5c1ef1c073480fed6 (arm64)
       Version: 44.20260210.0 (2026-02-11T10:48:10Z)
```

For the `rpi5` not all support is in the upstream kernel yet and thus not yet available in Fedora. The above images should boot; but it's likely that various peripherals don't work yet. For example, directly booting from NVMe, HDMI, that sort of stuff. It is possible to build with one of the custom `rpi5` kernels such as the one from [Peter Robinsons COPR](https://copr.fedorainfracloud.org/coprs/pbrobinson/a64-kernel/) or one from [rpmfusion](https://rpmfusion.org/Howto/RaspberryPi).

*Note that in this post I use `image-builder` for all examples. Generally you can swap out the examples for `bootc-image-builder`. I [wrote about](https://supakeen.com/weblog/merging-bootc-image-builder-and-image-builder/) what the differences between the two are and what their future is. So read that for the reasoning behind my preference.*

---

> If `image-builder` isn't shipped on your system; or you prefer to use containers you can use a container instead.
> ```
> sudo podman run --pull=newer --privileged -it \
>   -v ./blueprint.toml:/blueprint.toml \
>   -v /var/lib/containers/storage:/var/lib/containers/storage \
>   -v ./output:/output ghcr.io/osbuild/image-builder-cli:latest \
>     build \
>     --bootc-ref localhost/fedora-for-rpi3 \
>     --bootc-default-fs ext4 \
>     --blueprint /blueprint.toml \
>       raw
> ```
