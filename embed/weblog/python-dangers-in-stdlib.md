---
{
    "date": "",
    "tags": ["python", "stdlib"],
    "slug": "dangers-in-pythons-standard-library",
    "summary": "Things to be wary of when using Python's standard library."
}
---

# Dangers in Python's standard library

The Python programming language comes with "Batteries Included". A philosophy
to ship a comprehensive, immediately-useful standard library. However, since
the standard library comes with Python it is hard to refactor for older code
depends on it. Because of this the standard library can in many cases lag quite
far behind what is available in the ecosystem as a whole.

Some of the Python standard library is downright dangerous. By dangerous I mean
that special care has to be taken when using certain functions. This article
highlights some of the more well-known issues with the standard library, but is
by no means a comprehensive list of everything that can go wrong, nor a claim
that the Python standard library is bad.

This article is only about modules in the standard library; dangerous syntax,
types, and other language abilities are saved for another day. As always, pay
attention and try to think 'what does this actually do and can it be used in
a different way'.

Don't consider this to be a definitive list at any point in time. The initial
listed modules are popularly used.

## pickle

Pickle_ is a module in Python land to serialize 'arbitrary' objects. It is
often used when someone needs an easy way to send an objects' state elsewhere.

The Pickle module has many pitfalls. One of them is the fact that Pickled data
is meant to run on only the same Python version, and while it might sometimes
work on different versions (Pickle has a notion of a protocol version) its
interoperability leaves some things to be desired.

More damning is the fact that loading Pickled data allows for arbitrary code
execution. If you load Pickled data from sources you cannot trust -- something
much harder to guarantee than it might seem -- it is woeful. This combined
with the fact that a serialization format has interoperability issues should be
enough to steer well clear of it.

See the following example of unpickling some data, causing it to print
`hello`.

```python
import pickle

pickle.loads(b'\x80\x03cbuiltins\nprint\nq\x00X\x05\x00\x00\x00helloq\x01\x85q\x02Rq\x03.')

# hello
```

If you do ever need to exchange data, use a format that does not allow for any
'clever' things. A good option is to use JSON, which is available in the standard
library. You would need to write some code to explicitly convert your objects
to a format you are happy with, and some code to explicitly convert some
serialized data back to your objects.

There are libraries to help you with this such as marshmallow_.

## os.system, subprocess

`os.system` is inherently not safe with any user input and is likely not a
function you want to ever use due to its working. You would want to use the
`subprocess.*` functions and those come with a manual_ to use securely.

Seriously scrutinize any use of these classes of functions as they lead to
mistakes, most of which are easy to turn against you.

## shlex

shlex_ was a suggestion of mine in one of my previous articles. It is however
commonly used improperly. Consider if the following is the splitting you want:

```python
import shlex
import subprocess

user_input = shlex.split("foo;echo${IFS}hello")[0]
command = "echo {}".format(user_input)
subprocess.check_output(command, shell=True)

# b'foo\nhello\n'
```

The above shows a mismatch between what shlex thinks are separators for a shell
and what the actual outcome is. Someone less familiar with shlex might
assume that a shlex split always makes sure only a single argument is possible.

shlex can give a false sense of security if you are not absolutely certain you
know how shells work and what is in your input. Since you can never be certain
of the latter, you should prefer to work around having to use it.

## re

Regular expressions, you think you know them and now you have two problems.
While a powerful language that likely doesn't lead to direct exploitation one
does have to take care when writing these.

A possible attack is a denial of service by turning your own regular
expressions against you. Note the timings below.

```python
>>> timeit.timeit("import re;re.match('^(a+)+$', '{}!')".format("a" * 1))
1.4856852590019116
>>> timeit.timeit("import re;re.match('^(a+)+$', '{}!')".format("a" * 8))
40.852224354999635
```

## xml

XML, or eXtensible Markup Language is a format commonly (or less commonly
in current times) used to exchange data between different systems or for
general data serialization. XML is extremely flexible with a lot of knobs, this
has also led to a large amount of flaws possible in certain implementations.

This is well documented at the Python documentation website on xml_.

The excellent defusedxml_ package written by Christian Heimes has an amazing
README explaining all the issues, and has patches to make the standard Python
libraries and some other libraries less vulnerable. Read the description on
PyPI.

Any use of the built-in xml libraries should be scrutinized and where possible
be replaced with lxml_. lxml is a binding to libxml2 which comes with generally
secure defaults and a network sandbox.

## random

The default random_ module in Python will use a predictable random number
generator. If you use it for anything that is supposed to be secret please
use the secrets_ module or `random.SystemRandom()`.

It is a good idea to always use `secrets` or `SystemRandom` unless you are
certain you don't need it, instead of assuming the reverse.

.. _Pickle: https://docs.python.org/3/library/pickle.html#module-pickle
.. _marshmallow: https://marshmallow.readthedocs.io/
.. _manual: https://supakeen.com/weblog/0x01.html
.. _shlex: https://docs.python.org/3/library/shlex.html#module-shlex
.. _xml: https://docs.python.org/3/library/xml.html#module-xml
.. _defusedxml: https://pypi.org/project/defusedxml/
.. _lxml: https://lxml.de/
.. _random: https://docs.python.org/3/library/random.html
.. _secrets: https://docs.python.org/3/library/secrets.html
