#!/usr/bin/python
"""
Script to build and cross-compile clerb.

Usage:
    ./build.py [os (linux, windows, darwin)] [arch]
"""

import os, sys

SUPPORTED_OS = ["linux", "windows", "darwin"]
NODE_WEBKIT_VERSION = "0.9.2"

def build_for(plat, arch):
    print "Building Clerb for os %s, arch %s" % (plat, arch)

    if not os.path.exists("build"):
        os.mkdir("build")

    bin_name = "build/bin"
    if plat == "windows":
        bin_name = "build/bin.exe"

    # First build the go binary
    cmd = "GOOS=%s GOARCH=%s CGO_ENABLED=0 go build -o %s ../run_clerb.go" % (
        plat, arch, bin_name)
    os.popen(cmd)

    # TODO: node-webkit part

if __name__ == "__main__":
    arch = sys.argv[2] if len(sys.argv) > 2 else "amd64"

    osb = sys.argv[1] if len(sys.argv) > 1 else os.uname()[0].lower()
    if osb not in SUPPORTED_OS:
        sys.exit("OS must be one of %s, is %s" % (
            ', '.join(SUPPORTED_OS), osb))
    build_for(osb, arch)
