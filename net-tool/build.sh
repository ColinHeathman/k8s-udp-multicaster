#!/bin/sh
BASEDIR=$(dirname "$0")
docker build "${BASEDIR}" --tag net-tool:latest