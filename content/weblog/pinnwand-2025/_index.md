---
draft: false
date: 2026-01-01T06:00:00+00:00
title: pinnwand and bpa.st in 2025
tags: pinnwand
aliases:
  - pinnwand-in-2025.html
---

[pinnwand](https://github.com/supakeen/pinnwand) has existed for about a decade. It started as a separate pastebin to be used by [bpython](https://bpython-interpreter.org) on the `bpaste.net` domain. Eventually I rewrote it into its current shape which is a [Python](https://python.org) package based on the [Tornado Web Server](https://www.tornadoweb.org). It is currently used by a bunch of people (the [Python Discord](https://https://paste.pythondiscord.com/), [Rocky Linux](https://rpa.st), are the first that spring to mind; if you host a public instance I'd love to know about it. Perhaps we can keep a list on GitHub. Send me [an email](mailto:cmdr@supakeen.com)).

I've been steadily hosting the pinnwand instance that powers `bpaste.net`, and it's short and preferred domain [bpa.st](https://bpa.st) for the past 15 years or so and will continue doing so (not that type of weblog post). However, over the past year there have been some trends that are harmful to hosting your own pastebin that I sadly want to shine some lights on and am interested to know if other pastebin operators deal with the same trends.

Everything here is based on my experience with hosting `bpa.st`.

---

As usual every year and thus also in 2025 a pastebin receives spam. These are usually bots that either dump arrays of links and/or fake comments assuming they will be publically listed. The URLs for these pastes are rarely shared further as the bots don't really know what to do with them. Since I don't offer permanent pastes on `bpa.st` this is a minor problem; the links themselves aren't harmful and they don't garner much attention.

In the same rhythm there are users that accidentally post identifying, or things they shouldn't have shared. Usually users are able to [remove these themselves](https://bpa.st/removal) as they have a cookie identifying them as the owner. Sometimes external tools (such as [steck](https://github.com/supakeen/steck)) are used that might not forward the removal code to the user. A few users have contacted me over 2025 for paste removal and I'm happy to say that I've been able to respond to all of them quickly.

---

So far these have been operations as-per-usual. Sadly a new trend has emerged in 2025. In the previous 15 years or so I've only had to deal with 3rd party removal requests less than 5 times. All of these involved the sharing of warez.

In 2025 I've had to respond to 3rd party removal requests no less than 12 times. All of them involved the sharing of links to Child Sexual Abuse Material (CSAM).

The operation is much the same as others: a list of links gets posted but instead of referring to scam websites, or warez downloads I can only presume these refer to CSAM. I can only assume as I, by principle, do not open links or content in pastes that are requested to be removed but trust the reporter to be in good faith.

Generally I receive these removal requests directly in my inbox from organizations who detect this kind of thing. I'm very thankful for all the work people put into this. Sometimes these URLs are directly forward to my hoster's ([Hetzner](https://hetzner.com)) abuse department. My hoster gives me a single hour to remove this information; here's an example email (redacted):

```
We have received information about Child Sexual Abuse Material (CSAM) on your server:

-----
https://bpa.st/XXXXX
-----

Please remove the content within the next 1 (one!) hour.

If you do not remove the content within the next hour, we will lock the IP.
```

My hoster gives me one hour to remove the offending URLs before blackholing my IP address. Luckily I caught this email on time but I worry about missing one of these e-mails. So much so that I've set up extra notifications for any email sent to me by `abuse@`.

---

I have a small call to action: I have no good idea how to deal with this. `pinnwand` contains a naive [spam filter](https://github.com/supakeen/pinnwand/blob/master/src/pinnwand/defensive.py#L63-L82) which triggers on posts where too large of a percentage of the content consists of URLs but most of the pastes in this category evade it.

My question is thus, can you help me improve the defenses that `pinnwand` has to deal with this sort of thing? Perhaps you have ideas, or you've previously dealt with this. I don't want `bpa.st` to become a victim of tragedy of the commons. At the current rate this is not a risk as I can deal with the reports.

A potential stopgap that I'll likely be implementing is to allow operators to configure a denylist of words; this can then be used to block entire domains from appearing in any paste.

If you have ideas on how to deal with this in a way that keeps the spirit of a *simple* (to use, and to operate) pastebin alive don't hesitate hop onto [this issue on GitHub](https://github.com/supakeen/pinnwand/issues/326) or if you don't have an account there to [contact me through email](mailto:cmdr@supakeen.com).

---

Now that the nasty bits are out of the way there's one other change that I'm trialling on `bpa.st`. Historically I've only done statistics through [webalizer](https://webalizer.net/) but this year I'm trying out a javascript-based tracker called [Simple Analytics](https://www.simpleanalytics.com/) this tracker should honor Do-Not-Track and privacy. I'll be evaluating this in about a month or two and write about my experience. However, if you have any concerns about the tracker I'm using; or alternatives please [let me know](mailto:cmdr@supakeen.com).

Note that this tracker only applies to the deployment on `bpa.st` and isn't shared across other `pinnwand` instances.
