#!/usr/bin/env python3
import subprocess

path = "."

subprocess.check_output(["git", "add", path])
subprocess.check_output(["git", "commit", "-m", "generated new html"])
