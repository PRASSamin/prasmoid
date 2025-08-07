# Maintainer: Clem Lorteau <spam at lorteau dot fr>
Name:           prasmoid
Version:        0.0.3
Release:        1%{?dist}
Summary:        The All in One Development Toolkit for KDE Plasmoids
License:        MIT
URL:            https://github.com/PRASSamin/prasmoid

# copy the following 3 files to the same folder as this file
# TODO: check sha256 of binary
Source0:         https://github.com/PRASSamin/prasmoid/releases/download/v0.0.3/prasmoid
Source1:         https://raw.githubusercontent.com/PRASSamin/prasmoid/refs/tags/v0.0.3/LICENSE.md
Source2:         https://raw.githubusercontent.com/PRASSamin/prasmoid/refs/tags/v0.0.3/README.md

Requires:       git
Requires:       plasma-sdk
Requires:       qt5-qtdeclarative-devel

%description
Build, test, and manage your plasmoids with unparalleled ease and efficiency.

%install
install -Dm755 %{SOURCE0} "$RPM_BUILD_ROOT/usr/bin/prasmoid"
install -Dm644 %{SOURCE1} "$RPM_BUILD_ROOT/usr/share/licenses/prasmoid/LICENSE.md"
install -Dm644 %{SOURCE2} "$RPM_BUILD_ROOT/usr/share/doc/prasmoid/README.md"

%files
%{_bindir}/prasmoid
%license %{_datadir}/licenses/prasmoid/LICENSE.md
%doc %{_docdir}/prasmoid/README.md

%changelog
%autochangelog
