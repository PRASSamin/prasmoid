#!/bin/bash

# update sha256 sums automatically
updpkgsums

# check PKGBUILD
namcap PKGBUILD

# build in chroot
pkgctl build

# check package
namcap *.pkg.tar.zst

# upload - to be run in the git tree for the arch package, not the project's git tree
#git add PKGBUILD .SRCINFO
#git commit -m "something useful"
#git push