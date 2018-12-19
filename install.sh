#!/bin/bash

get_latest_release() {
  curl -sL "https://api.github.com/repos/$1/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/' |
    sed -E 's/v?(.*)/\1/g'
}

repo="orange-cloudfoundry/mdproxy4cs"
version=$(get_latest_release "${repo}")
name="mdproxy4cs-${version}.linux-amd64"
file="${name}.tar.gz"

echo curl -sL https://github.com/${repo}/releases/download/v${version}/${file} --output /tmp/${file}

echo tar -C /tmp xvzf /tmp/${file}
echo cp /tmp/${name}/mdproxy4cs /usr/bin/

echo mkdir -p /usr/share/mdproxy4cs/
echo cp /tmp/${name}/pre-start.sh /usr/share/mdproxy4cs/pre-start.sh

echo cp /tmp/${name}/mdproxy4cs.service /etc/systemd/system/
echo systemctl enable /etc/systemd/system/mdproxy4cs.service
