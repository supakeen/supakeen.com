---
draft: false
date: 2020-04-25T12:00:00+00:00
title: A varnishncsa logfile per host
tags: varnish, configuration
aliases:
  - varnishncsa-per-host.html
---

Varnish_ is a caching proxy server in quite widespread use. Its common role is
to set in between your load balancer and web servers to lessen the requests
ending up at the latter.

A often recurring situation is that your Varnish instance serves multiple
different hostnames. If you want to use some of the older tools such as
Webalizer_ or awstats_ to process your logs you probably need a log file per
hostname.

For this we can use `varnishncsa` a utility that comes with Varnish to
retrieve logfiles. This post was written with Ubuntu 18.04 as a base system
but should work on your system even if you might have to change some paths.

Normally a `varnishncsa.service` file is created for `systemd` to run a
default instance of it, these logs end up in `/var/log/varnish`. We're going
to add our own parametrized service files that filter out specific hostnames.

Note that you will end up running multiple instances of `varnishncsa`, this
hasn't been a problem for me as they are quite lightweight.

Let's get started.

## Setup

Start by creating a new service file in `/lib/systemd/system` under the name
`varnishncsa-per-host@.service`.

```
# cat /lib/systemd/system/varnishncsa-per-host\@.service
[Unit]
Description=Varnish HTTP accelerator log daemon for Host %I
Documentation=https://supakeen.com/weblog/varnishncsa-per-host.html
After=varnish.service

[Service]
Type=forking
PIDFile=/run/varnishncsa/varnishncsa-%I.pid
RuntimeDirectory=varnishncsa
User=varnishlog
Group=varnish
ExecStart=/usr/bin/varnishncsa -a -q "ReqHeader ~ '^Host: %I'" -w /var/log/varnish/varnishncsa-%I.log -D -P /run/varnishncsa/varnishncsa-%I.pid
ExecReload=/bin/kill -HUP $MAINPID
PrivateDevices=true
PrivateNetwork=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
```

In this file you see multiple `%I` format specifiers. These will be replaced
by `systemd`. We are using `varnishncsa`'s filtering on the `Host` header
to find only the host we're interested in.

After we've created this file we can now create symlinks to enable this service
for a hostname. Repeat this step per Host header you want to filter out.

```
# ln -s /lib/systemd/system/varnishncsa-per-host\@.service ./varnishncsa-per-host@supakeen.com.service
# systemctl daemon-reload
# systemctl enable varnishncsa-per-host@supakeen.com.service
# systemctl start varnishncsa-per-host@supakeen.com.service
# systemctl status varnishncsa-per-host@supakeen.com.service
● varnishncsa-per-host@supakeen.com.service - Varnish HTTP accelerator log daemon for Host supakeen.com
   Loaded: loaded (/lib/systemd/system/varnishncsa-per-host@.service; indirect; vendor preset: enabled)
   Active: active (running) since Sat 2020-04-25 09:26:34 UTC; 18min ago
     Docs: https://supakeen.com/weblog/varnishncsa-per-host.html
 Main PID: 20822 (varnishncsa)
    Tasks: 1 (limit: 1151)
   CGroup: /system.slice/system-varnishncsa\x2dper\x2dhost.slice/varnishncsa-per-host@supakeen.com.service
           └─20822 /usr/bin/varnishncsa -a -q ReqHeader ~ '^Host: supakeen.com' -w /var/log/varnish/varnishncsa-supakeen.com.log -D -P /run/varnishncsa/varnishncsa-supakeen.com.pid

Apr 25 09:26:34 var.tty.cat systemd[1]: Starting Varnish HTTP accelerator log daemon for Host supakeen.com...
Apr 25 09:26:34 var.tty.cat systemd[1]: varnishncsa-per-host@supakeen.com.service: Failed to parse PID from file /run/varnishncsa/varnishncsa-supakeen.com.pid: Invalid argument
Apr 25 09:26:34 var.tty.cat systemd[1]: Started Varnish HTTP accelerator log daemon for Host supakeen.com.
root@var:~#
```

And a new file now shows up in `/var/log/varnish`

```
# stat /var/log/varnish/varnishncsa-supakeen.com.log
  File: /var/log/varnish/varnishncsa-supakeen.com.log
  Size: 1946            Blocks: 8          IO Block: 4096   regular file
Device: fc01h/64513d    Inode: 260830      Links: 1
Access: (0644/-rw-r--r--)  Uid: (  113/varnishlog)   Gid: (  115/ varnish)
Access: 2020-04-25 09:26:34.904152413 +0000
Modify: 2020-04-25 09:34:43.891850255 +0000
Change: 2020-04-25 09:34:43.891850255 +0000
 Birth: -
```

That's it! You can now use this file any way you want. Good luck.

.. _Varnish: https://varnish-cache.org/
.. _webalizer: http://www.webalizer.org/
.. _awstats: https://awstats.sourceforge.io/
