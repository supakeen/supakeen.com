#!/usr/bin/env python3
import subprocess
import sys
import json
import re

path = "."

def main():
    targets = sys.argv[1:]

    if not len(targets):
        targets = ["github", "ipns"]

    if "ipns" in targets:
        print("Publishing to IPFS")
        site = subprocess.check_output(["ipfs", "add", "-r", path])
        site = site.splitlines()[-1].split()[1]
        print("Content hash", site.decode("ascii"))
        print("Updating ipns name")
        subprocess.check_output(["ipfs", "name", "publish", site])

        # TODO this is obviously not easy to solve and will always point to the
        # TODO old content hash for the site, however do people on IPFS and are
        # TODO already on the site really want to confirm??

        print("Rewriting index")
        with open("index.html", "r") as handle:
            index = handle.read()

        index = re.sub("IPFS <em>([a-zA-Z0-9]+)</em>", "IPFS <em>{}</em>".format(site.decode("ascii")), index)

        print(index)

        with open("index.html", "w") as handle:
            handle.write(index)

    if "github" in targets:
        # Do a pretty straightforward git push first
        subprocess.check_output(["git", "add", path])
        subprocess.check_output(["git", "commit", "-m", "generated new html"])
        subprocess.check_output(["git", "push"])

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
