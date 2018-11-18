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

    if "github" in targets:
        # Do a pretty straightforward git push first
        subprocess.check_output(["git", "add", path])
        subprocess.check_output(["git", "commit", "-m", "generated new html"])
        subprocess.check_output(["git", "push"])

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
