FROM golang:1.22.5 AS build

RUN git clone https://github.com/OpenCHAMI/smd.git /smd
WORKDIR /smd
RUN make binaries

FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache tini

# Include curl in the final image.
RUN set -ex \
    && apk -U upgrade \
    && apk add --no-cache curl

COPY --from=build /smd/smd /
COPY --from=build /smd/smd-loader /
COPY --from=build /smd/smd-init /
RUN mkdir /persistent_migrations
COPY migrations/* /persistent_migrations/

EXPOSE 27779

# nobody 65534:65534
USER 65534:65534

CMD [ "/smd" ]

ENTRYPOINT [ "/sbin/tini", "--" ]
