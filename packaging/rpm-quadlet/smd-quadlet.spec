# SPDX-FileCopyrightText: 2026 OpenCHAMI Contributors
# SPDX-License-Identifier: MIT
#
# See `make rpm-build` and docs/RPM_PACKAGING.md for the tag-to-version
# mapping and how the packaged quadlet's image tag is pinned to it.

Name:           smd-quadlet
Version:        %{version}
Release:        %{rel}%{?dist}
Summary:        OpenCHAMI smd Quadlet units

License:        MIT
URL:            https://github.com/OpenCHAMI/smd
Source0:        %{name}-%{version}.tar.gz

BuildArch:      noarch

Requires(post,preun,postun):  systemd
Requires:                     podman >= 4.4.0
Requires:                     openchami-postgres-quadlet >= 0.0.1 tokensmith-quadlet >= 0.4.0

%description
Podman Quadlet unit files (container + volume) for running smd
as part of an OpenCHAMI deployment.

%prep
%setup -q

%install

install -d %{buildroot}/usr/share/containers/systemd
for f in smd smd-init; do
    grep -q '@IMAGE_TAG@' $f.container
    sed "s|@IMAGE_TAG@|v%{version}|" $f.container \
        | install -m 644 /dev/stdin %{buildroot}/usr/share/containers/systemd/$f.container
done

%files
%license LICENSES/MIT.txt
/usr/share/containers/systemd/smd.container
/usr/share/containers/systemd/smd-init.container

%post
# reload systemd so the new Quadlet-generated unit is seen
systemctl daemon-reload || :
if [ $1 -ge 2 ]; then
    systemctl try-restart smd.service || :
fi

%preun
if [ $1 -eq 0 ]; then
    systemctl stop smd.service >/dev/null 2>&1 || :
fi

%postun
# reload systemd so the removed unit is dropped
systemctl daemon-reload || :
