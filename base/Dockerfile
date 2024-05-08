# syntax=docker/dockerfile:1
FROM docker.io/library/python:3.12-slim as prod-ipynb-deps

ENV VIRTUAL_ENV=/opt/venv
RUN python -m venv $VIRTUAL_ENV
ENV PATH="$VIRTUAL_ENV/bin:$PATH"

COPY requirements.txt /requirements.txt
RUN pip install --disable-pip-version-check --no-cache-dir --only-binary=:all: -r /requirements.txt

ENV ACE_TELEMETRY_JUPYTER $VIRTUAL_ENV/bin/jupyter

# For testing
RUN uname -a
