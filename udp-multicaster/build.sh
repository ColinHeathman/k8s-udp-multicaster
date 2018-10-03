#!/bin/sh
BASEDIR=$(dirname "$0")
docker build "${BASEDIR}" --tag udp-multicaster:latest