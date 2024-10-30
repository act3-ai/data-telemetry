FROM docker.io/library/golang:1.23.2-bookworm as deps

# Install debian packages
RUN apt-get update -q && apt-get install -yq --no-install-recommends \
unzip \
&& rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY Makefile Makefile
RUN make deps && rm -rf internal/webapp/assets/.downloads
RUN pwd

# hadolint ignore=DL3006
FROM cgr.dev/chainguard/static as dev

ENV GOTRACEBACK=all

COPY ./ci-dist/telemetry/linux/amd64/bin/telemetry /usr/local/bin/telemetry
ENTRYPOINT ["telemetry"]
EXPOSE 8100

WORKDIR /

# We use the same UID/GID as we use in the helm chart to support skaffold's sync capabilty
# COPY --chown does work with kaniko see https://github.com/GoogleContainerTools/kaniko/issues/1456 and https://github.com/GoogleContainerTools/kaniko/issues/1603
# COPY --chown=12345:12345 assets /assets

LABEL maintainers="Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>"
LABEL description="ACE Data Telemetry -- bottle and experiment tracking (dev image)"

FROM deps as ci

# hadolint ignore=DL3006
FROM cgr.dev/chainguard/static as prod

COPY ./ci-dist/telemetry/linux/amd64/bin/telemetry /usr/local/bin/telemetry
ENTRYPOINT ["telemetry"]
EXPOSE 8100

WORKDIR /

LABEL maintainers="Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>"
LABEL description="ACE Data Telemetry -- bottle and experiment tracking"

# hadolint ignore=DL3006
FROM cgr.dev/chainguard/python:latest-dev as builder-ipynb

ENV LANG=C.UTF-8
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PATH="/opt/venv/bin:$PATH"
ENV ACE_TELEMETRY_JUPYTER /opt/venv/bin/jupyter

WORKDIR /opt/venv
RUN python -m venv /opt/venv

COPY requirements.txt /opt/requirements.txt
RUN pip install --disable-pip-version-check --no-cache-dir --only-binary=:all: -r /opt/requirements.txt

# hadolint ignore=DL3006
FROM cgr.dev/chainguard/python as prod-ipynb

ENV PYTHONUNBUFFERED=1
ENV PATH="/opt/venv/bin:$PATH"
ENV ACE_TELEMETRY_JUPYTER /opt/venv/bin/jupyter

COPY ./ci-dist/telemetry/linux/amd64/bin/telemetry /usr/local/bin/telemetry
COPY --from=builder-ipynb /opt/venv /opt/venv
ENTRYPOINT ["telemetry"]
EXPOSE 8100
WORKDIR /home/nonroot

LABEL maintainers="Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>"
LABEL description="ACE Data Telemetry -- bottle and experiment tracking (includes ipynb support)"
