# SPDX-FileCopyrightText: Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

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

RUN mkdir -p /opt/smd-test && \
    mkdir -p /var/logs/smd-test && \
    mkdir -p /tests && \
    mkdir -p /.pytest_cache && \
    chown -R 65534:65534 /.pytest_cache && \
    chmod 755 /.pytest_cache && \
    chown -R 65534:65534 /opt/smd-test && \
    chown -R 65534:65534 /tests && \
    chown -R 65534:65534 /var/logs/smd-test && \
    chmod 755 /var/logs/smd-test
COPY bin  /opt/smd-test/bin
COPY libs /opt/smd-test/libs
COPY ct /tests/ct
COPY smoke_pytest /opt/smd-test/smoke_pytest



# Run as nobody
USER 65534:65534
ENV PATH="$PATH:/opt/smd-test/bin"

WORKDIR /
CMD [ "/opt/smd-test/bin/smd-test" ]
