Concepts
########

Asynchronous I/O
================
While I have an entire page dedicated to my notes on asynchrony_ the
concept of Asynchronous I/O can be explained shortly.

Computers often have to wait for things and while waiting they could be doing
other things. Asynchronous I/O lets a program ask the computer to let it know
when some I/O is done. This means the computer can continue with more tasks it
has to do and has grown into an entire idiom often combined with terms such as
an event loop making it complex to understand.

The gist of it is that while your computer is waiting for some I/O another
thing can start its work. When that task in turn has to wait for some I/O it
will then let yet another task do its work.

When the I/O is done a task can be resumed again at the point where it started
waiting for I/O as if nothing ever happened. There is still one task executing
at a time but many tasks can be waiting for I/O at the same time.

This is neat because a lot of things on computers are related to I/O.

Encodings
=========
Computers deal in numbers not in letters or pictures. Humans do use letters and
other ways of expressing ourselves so there's a slight mismatch between the two.

When computers represent things that are not numbers they usually represent them
as numbers and then map these numbers on certain letters or colors. This is
called an encoding, converting between certain numbers and certain values.

A common encoding is called ASCII_. ASCII is an encoding for letters, each
letter gets a number assigned to it. This way the computer knows that when it
sees the number ``69`` it should really show the letter ``E``.

ASCII is not the only encoding. In other encodings that same number might mean
something else entirely. A problem arises when you have some numbers but you
don't know what these numbers might refer to. Perhaps an encoding BSCII exists
where the number ``69`` is actually the letter ``F``.

You should never guess at encodings, things can go horribly wrong. If you don't
know what some numbers represent then the only way to guess at it is by using
statistical analysis (the letter ``e`` is the most common letter in the English
alphabet but it's not in Chinese or Russian, so you can see where this can go
wrong easily).

A common encoding for characters and letters right now is UTF-8_ and it's the
only one you should use. In UTF-8 not all characters map onto a single number.
Sometimes up to four numbers need to be used to represent a character.

This allows it to represent all of Unicode_ which is a collection of many
characters, letters, and icons from around the world. Unicode is confusing
because it calls itself an encoding because it maps numbers (often called
code points) onto characters and icons. Computers however rarely deal with the
actual Unicode numbers instead using an encoding such as UTF-8 to represent
the Unicode numbers (yes, encoding an encoding).

Memoization
===========
The word memoization means to remember things. It's often a fast way to speed
up certain things to do on computers by remembering the previous result of an
action.

Take a pure function `f(x)`. Pure means that the function only depends on its
inputs and that it could be replaced by whatever it returns and your program
would not change.

Functions like those are prime targets for memoization. I can keep a map of all
previous calls and their results and when I go to call the function again I
first see if I've already called the function with my argument. If so I can
immediately get the result from the map instead of computing the value again.

This is a example of a time/space tradeoff where I trade some space in the form
of memory for some time in speed.

Memoization is a form of caching.

Side effects
============
When you have functions or parts of your program that modifies some things
outside of what you pass to it this is called a side-effect.

This makes it very hard to predict what a certain function or piece of code
does.

Trade-offs
==========

.. _ASCII: https://en.wikipedia.org/wiki/ASCII
.. _Unicode: https://home.unicode.org/basic-info/faq/
.. _UTF-8: https://en.wikipedia.org/wiki/UTF-8
