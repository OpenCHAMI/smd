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

# podman 5.0.0 is the first release whose Quadlet generator enables
# systemd-style *.container.d drop-in directories
Requires:                     podman >= 5.0.0

# NOTE: tokensmith-quadlet and postgres-quadlet are Suggests rather than
# Requires. Neither is necessary to run the service: auth is optional and
# tokensmith is simply our recommended provider, while postgres may live
# elsewhere depending on the deployment strategy used by the site. Requiring
# postgres-quadlet would also force the inclusion of an internal network
# quadlet file, which conflicts with the standalone premise of these RPMs.
#
# The OpenCHAMI devs will provide separate meta-conf RPMs that glue the
# services together for specific deployment styles, tuning behavior via
# systemd drop-in overrides. The first will target single-node quadlet
# deployments.
Suggests:                     openchami-postgres-quadlet >= 0.0.1
Suggests:                     tokensmith-quadlet >= 0.4.0

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
    install -d %{buildroot}/usr/share/containers/systemd/$f.container.d
    install -m 644 $f.container.d/10-defaults.conf \
        %{buildroot}/usr/share/containers/systemd/$f.container.d/
done

%files
%license LICENSES/MIT.txt
/usr/share/containers/systemd/smd.container
/usr/share/containers/systemd/smd.container.d
/usr/share/containers/systemd/smd.container.d/10-defaults.conf
/usr/share/containers/systemd/smd-init.container
/usr/share/containers/systemd/smd-init.container.d
/usr/share/containers/systemd/smd-init.container.d/10-defaults.conf

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
