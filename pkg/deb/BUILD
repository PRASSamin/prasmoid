#!/bin/bash
#
# Builds source and binary debian packages by compiling from source in a base chroot
# Assumes a fresh pbuilder up-to-date environment for the target distro with nothing in 'result'

set -e

DIST=plucky

# build in chroot
sudo cowbuilder build --logfile build.log --distribution $DIST --hookdir $(pwd)/prasmoid/debian/pbuilderhook.d prasmoid_0.1.0-1.dsc

# check
lintian /var/cache/pbuilder/result/*source.changes

# upload
#sudo sed -i '/\.buildinfo/d' /var/cache/pbuilder/result/*source.changes #don't upload binary
#debsign /var/cache/pbuilder/result/*source.changes
#dput ppa:northern-lights/prasmoid /var/cache/pbuilder/result/*source.changes
