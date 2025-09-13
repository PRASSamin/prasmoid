#!/bin/bash

set -e

# download specified version of sources
spectool -g prasmoid.spec

# build
fedpkg --release f42 mockbuild

# check
# rpm --query --package --list results_prasmoid/0.0.4/1.fc42/prasmoid-0.0.4-1.fc42.x86_64.rpm
# fedpkg --release f42 lint

# upload
#copr-cli build prasmoid prasmoid-0.0.4-1.fc42.src.rpm

