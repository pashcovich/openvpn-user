#!/usr/bin/env bash

apt-get update
apt-get install -y gcc-multilib libc6 libc6-dev libc6-dev-i386 sqlite3 libc6-dev-arm64-cross gcc-arm-linux-gnueabi gcc-aarch64-linux-gnu
