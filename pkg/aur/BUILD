#!/bin/bash

# check PKGBUILD
namcap PKGBUILD

# build in chroot
pkgctl build

# check package
namcap *.pkg.tar.zst

# upload
#git add PKGBUILD .SRCINFO
#git commit -m "something useful"
#git push