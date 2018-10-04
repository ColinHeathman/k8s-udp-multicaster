#!/bin/sh
BASEDIR=$(dirname "$0")
docker build "${BASEDIR}" --tag k8s-udp-multicaster:latest