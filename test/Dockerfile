# MIT License
#
# (C) Copyright [2025] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

FROM docker.io/library/alpine:3.15

STOPSIGNAL SIGTERM

# Install the necessary packages
RUN set -ex \
    && apk -U upgrade \
    && apk add --no-cache \
        python3 \
        python3-dev \
        py3-pip \
        bash \
        curl \
        tar \
        gcc \
        musl-dev \
    && pip3 install --upgrade \
        pip \
        pytest==7.1.2 \
        tavern==1.23.1 \
        allure-pytest==2.12.0 \
    && apk del \
        python3-dev \
        tar \
        gcc \
        musl-dev

RUN mkdir -p /opt/ochami-pytest && \
    mkdir -p /var/logs/ochami-pytest && \
    mkdir -p /tests && \
    mkdir -p /.pytest_cache && \
    chown -R 65534:65534 /.pytest_cache && \
    chmod 755 /.pytest_cache && \
    chown -R 65534:65534 /opt/ochami-pytest && \
    chown -R 65534:65534 /tests && \
    chown -R 65534:65534 /var/logs/ochami-pytest && \
    chmod 755 /var/logs/ochami-pytest
COPY bin  /opt/ochami-pytest/bin
COPY libs /opt/ochami-pytest/libs
COPY ct /tests/ct
COPY smoke_pytest /opt/ochami-pytest/smoke_pytest



# Run as nobody
USER 65534:65534
ENV PATH="$PATH:/opt/ochami-pytest/bin"

WORKDIR /
CMD [ "/opt/ochami-pytest/bin/smd-test", "help" ]
