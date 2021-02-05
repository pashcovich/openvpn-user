#!/usr/bin/env sh

set -e

openvpn-user auth --user $(head -1 $1) --password $(tail -1 $1)
