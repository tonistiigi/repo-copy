# syntax = docker/dockerfile-upstream:0.9.0-experimental

# The `buildkitd` stage and the `buildctl` stage are placed here
# so that they can be built quickly with legacy DAG-unaware `docker build --target=...`

FROM golang:1.11-alpine AS gobuild-base
RUN apk add --no-cache git make g++ libseccomp-dev


FROM gobuild-base AS repo-copy
WORKDIR /go/src/github.com/tonistiigi/repo-copy
RUN --mount=target=. go build -o /out/repo-copy ./

FROM gobuild-base AS containerd
RUN apk add --no-cache btrfs-progs-dev
ARG CONTAINERD_VERSION=55420c95
RUN git clone https://github.com/tonistiigi/containerd.git /go/src/github.com/containerd/containerd
WORKDIR /go/src/github.com/containerd/containerd
RUN git checkout -q "$CONTAINERD_VERSION" \
  && make bin/containerd \
  && make bin/ctr

FROM alpine
RUN apk add --no-cache libseccomp
COPY --from=linuxkit/ca-certificates:v0.6 / /
VOLUME /var/lib/containerd
RUN mkdir -p /root/.docker && echo "{}" > /root/.docker/config.json
COPY --from=containerd /go/src/github.com/containerd/containerd/bin/ /bin/
COPY --from=repo-copy /out/repo-copy /bin/
ENTRYPOINT ["/bin/repo-copy"]