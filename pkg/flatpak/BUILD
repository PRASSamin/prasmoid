#!/bin/bash
set -e

# build
flatpak-builder --force-clean --user --install-deps-from=flathub --repo=repo --install builddir io.github.prassamin.prasmoid.yml

# check
flatpak run --command=flatpak-builder-lint org.flatpak.Builder manifest io.github.prassamin.prasmoid.yml

# test
flatpak run io.github.prassamin.prasmoid --version

# bundle
flatpak build-bundle repo prasmoid-0.0.4.flatpak io.github.prassamin.prasmoid --runtime-repo=https://flathub.org/repo/flathub.flatpakrepo

# upload - to be run in the git tree for the flat package, not the project's git tree
#git add flathub.json io.github.prassamin.prasmoid.metainfo.xml io.github.prassamin.prasmoid.yml
#git commit -m "something useful"
#git push
# then open a pull request - https://docs.flathub.org/docs/for-app-authors/submission