# Multi-architecture building with Docker

## Setup

```shell
brew install docker-buildx
sudo apt install qemu
docker run --privileged --rm tonistiigi/binfmt --install all
docker buildx create --name mybuilder --driver docker-container --bootstrap
docker buildx use mybuilder
docker buildx ls
```

[Tutorial for more information](https://docs.docker.com/build/building/multi-platform/#build-multi-arch-images-with-buildx)

## Building

Build the Intel and ARM multi-platform image with QEMU on Intel.

```shell
cp ../requirements.txt .
docker buildx build --platform linux/amd64,linux/arm64 -t reg.git.act3-ace.com/ktarplee/test/telemetry/base:latest --push .
```

This might be possible in podman as well...

```shell
podman build --platform linux/amd64,linux/arm64 --manifest reg.git.act3-ace.com/ktarplee/test/telemetry/base:latest .
podman manifest push reg.git.act3-ace.com/ktarplee/test/telemetry/base:latest reg.git.act3-ace.com/ktarplee/test/telemetry/base:latest
```

You can also have KO build images from this as the base image by changing `fromImage` in `skaffold.yaml` or `defaultBaseImage` in `.ko.yaml`.
