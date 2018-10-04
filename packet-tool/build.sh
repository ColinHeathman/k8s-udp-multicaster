#!/bin/sh
BASEDIR=$(dirname "$0")
docker build "${BASEDIR}" --tag packet-tool:latest