#!/usr/bin/env sh

PATH=$PATH:/usr/local/bin
set -e
auth_usr=$(head -1 $1)
auth_passwd=$(tail -1 $1)

if [ $common_name = ${auth_usr} ]; then
  openvpn-user auth --user ${auth_usr} --password ${auth_passwd}
else
  echo "Authorization failed"
  exit 1
fi