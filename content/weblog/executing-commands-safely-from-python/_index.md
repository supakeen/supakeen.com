---
draft: false
date: 2018-05-25T12:00:00+00:00
title: Executing commands safely from Python
tags: python, subprocess
category: python
aliases:
  - executing-commands-safely-from-python.html
---

Python provides multiple ways to execute commands on the system it is running
on. Some of them inherently unsafe, some of them safe in nature but easy to
use in an unsafe way.

Here I will set out to document the current ways to execute commands with
modules included in Python 3's standard library. Their pros, and their cons.
This article assumes that you are familiar with shells, you don't need to know
everything about them but you do need to know about their basic syntax_. I also
assume you are using Python 3 and are on Linux while concepts will carry over
to all languages and operating systems.

## Command Injection

To start out you need to understand why executing commands from Python can be
dangerous. This principle applies to all languages and is called
**Command Injection**, there are  some examples on the OWASP_ pages and the CWE-77_
page. I will provide my own here.

Here is some code that will restart a service on your system by the name of
the argument it receives. I name this program `service.py` and its goal is
to restart services. To do that it uses a function to execute commands called 
os.system_.

```python
import sys
import os

os.system("systemctl restart {}".format(sys.argv[1]))
```

If we call our program with `python service.py nginx` the string that gets put
into our `os.system`-call will be the string `systemctl restart nginx` and all
is good in the world. However, if someone calls our program as 
`python service.py 'nginx;cat /etc/passwd'` our executed command will become:

```bash
systemctl restart nginx;
cat /etc/passwd
```

Where I have added the newline myself for clarity. Our program was not intended to
be reading the `/etc/passwd` file at all! This is a command injection and it comes
in many shapes and forms and is something you want to prevent.

Any place where input is passed into a command to be executed one needs to be
especially careful. This can be in scripts such as the example above or websites,
network protocols, and others. Sometimes input can be things you wouldn't
expect to be input and is a reason why I won't call it *user input* in this
article. It can be, for example, an HTTP request made by your application that is
changed by a man in the middle attack on an unsafe network, which can put the client
at risk.

## How does a command get executed?

Before I can talk about how to prevent these types of attacks it is important
to dive a tiny bit deeper. How does a command get executed by your operating
system?

In general your operating system's library will use a set of functions called
`exec*` functions where the `*` can be filled with a variety of letters. They
are documented in the man-pages_.

These seem a bit daunting but in general all these functions follow the same
pattern. They all take a `path` or `file` to execute, if the function takes a
`file` the path to the name of that file will be looked up by parsing the
`PATH` environment variable.

Some of these functions also allow one to pass the environment to be set for
the executable that will be executed. However they all share a common idiom
which is **executable** followed by a varying number of arguments.

This means that whenever we execute a string in the form of
`systemctl restart nginx` something needs to parse that string into the parts
`systemctl`, `restart`, and `nginx` and give it to one of the functions in the
`exec*` family. This tends to be done by your shell.

If we jump back to our previous `os.system` program it will call the system_
function in your standard C library which will in turn execute the command
`sh -c 'systemctl restart nginx'` to allow the `sh` executable, which is a
shell, to parse the command into the parts necessary for the `exec*` function
used.

## Shells

As soon as a shell gets involved in parsing your command we are entering a very
dangerous state regarding the characters that are in our command to be executed.
Shells allow executing multiple commands at once, they have built-ins that allow
you to do things without calling commands. Someone can chain everything they want
in there by gaining control of a parameter that gets fed to a shell and shells get
involved in places where you sometimes don't know they will be.

Can we make arguments passed to shells safe? No, not really. You want to
use a function which does not use a shell at all to prevent shell-based
exploits.

## Ways to execute commands in Python

Python 3 offers a variety of ways for executing commands but there is one which
springs out and that is the subprocess_-module.

The subprocess_-module allows us to execute commands without opening a shell to
parse our string into the appropriate parts. This puts us at minimal risk for
being exploited.

**Note: Of course the program you are executing through subprocess can still have
its own flaws that allow it to be subverted to do things you don't want.**

Let's make a version of our previous program using subprocess. Subprocess offers
many functions but they all follow the same rules for their arguments:

```python
import sys
import subprocess

subprocess.run(["systemctl", "restart", sys.argv[1]])
```

Subprocess's methods take either a list of arguments or a single string. Remember
the previous explanation about the `exec*` family of functions.

When you pass a list to subprocess as I've done above then your list will be split,
the first item will be the first argument to the `exec*` function and the rest of
the arguments will each be passed as a separate argument.

This means arguments are not interpreted by a shell first and this makes it impossible
for someone to execute other commands through the shell.

If you pass a single string to subprocess such as:

```python
import subprocess

subprocess.run("systemctl restart nginx")
```

Then that string will be the first argument to the `exec*` without any splitting,
the arguments will be left empty. If you execute the command above then the `exec*`
function will look for an executable called `systemctl restart nginx` on your `PATH`
which will likely not exist.

This is a safe way to execute commands in Python even when input is passed as
arguments to your executable.

### shell=True

Subprocess's methods take an additional keyword argument called `shell` which
can be set to `True`. If you do so then you can only pass a string which will
be passed the same way, as `sh -c 'command'`, if you do pass a list then it will
be passed as:

```
execve("/bin/sh", ["/bin/sh", "-c", "systemctl", "restart", "nginx"], ...
```

## What if I need a shell?

Executing commands in the safe way as described above means that you can't use
those handy shell features you are used to such as `|`, `<`, `>` and their friends.

Most of these functions can be implemented separately in Python. If you need
a `|` it is often better to execute the first command, store its output and then
execute the second command giving the output to the new process.

File redirection (`>`, and others) can be done in the same way by storing the
output and then writing it to a file in Python.

For most command line utilities you would normally use with these operators you
can either trivially implement them in Python. You can also try to find a library
on PyPI_ to give you the output directly instead of trying to parse `ip`, `ifconfig`,
or others in a shell.

## What if I really really need a shell?

You could use Python's shlex_-module which tries to implement the proper escaping
rules for shells. Specifically you could try to use `shlex.quote` for each argument
you fill in. Reasoning about what is 'safe' or 'unsafe' becomes very difficult in
this context.

.. _syntax: https://www.w3resource.com/linux-system-administration/control-operators.php
.. _OWASP: https://www.owasp.org/index.php/Command_Injection
.. _CWE-77: https://cwe.mitre.org/data/definitions/77.html
.. _os.system: https://docs.python.org/3/library/os.html#os.system
.. _man-pages: https://linux.die.net/man/3/exec
.. _system: https://linux.die.net/man/3/system
.. _subprocess: https://docs.python.org/3/library/
.. _PyPI: https://pypi.python.org
.. _shlex: https://docs.python.org/3/library/shlex.html
