%bcond check 1

%global goipath         github.com/PRASSamin/prasmoid
Version:                0.0.4

%gometa -L -f

Name:           prasmoid
Release:        1%{?dist}
Summary:        The All in One Development Toolkit for KDE Plasmoids

License:        Apache-2.0 AND BSD-2-Clause AND BSD-3-Clause AND ISC AND MIT
URL:            %{gourl}
Source0:        %{gosource}
Source1:        %{archivename}-vendor.tar.bz2
Source2:        go-vendor-tools.toml

BuildRequires:  go-vendor-tools
BuildRequires:  gettext
BuildRequires:  dnf
BuildRequires:  sudo
BuildRequires:  qt5-qtdeclarative-devel
BuildRequires:  plasma-sdk

Requires:       git
Requires:       plasma-sdk
Requires:       qt5-qtdeclarative-devel

%description
The All in One Development Toolkit for KDE Plasmoids.

%prep
%goprep -A
%setup -q -T -D -a1 %{forgesetupargs}
# %autopatch -p1

%generate_buildrequires
%go_vendor_license_buildrequires -c %{S:2}

%build
%global gomodulesmode GO111MODULE=on
export PATH=$PATH:
%gobuild -o %{gobuilddir}/bin/prasmoid %{goipath}
# for cmd in prasmoid/*; do
#   %gobuild -o %{gobuilddir}/bin/$(basename $cmd) %{goipath}/$cmd
# done
# rm %{gobuilddir}/bin/prasmoid
# mv %{gobuilddir}/bin/src %{gobuilddir}/bin/prasmoid

%install
%go_vendor_license_install -c %{S:2}
install -m 0755 -vd %{buildroot}%{_bindir}
install -m 0755 -vp %{gobuilddir}/bin/prasmoid %{buildroot}%{_bindir}/
# rm %{buildroot}/lib/debug/usr/bin/*.debug

%check
%go_vendor_license_check -c %{S:2}
%if %{with check}
%gotest ./...
%endif

%files -f %{go_vendor_license_filelist}
%license vendor/modules.txt
%doc CHANGELOG.md README.md
/usr/bin/prasmoid

%changelog
%autochangelog
