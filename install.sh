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
dir=$(mktemp -d)

cd ${dir}

curl -sL https://github.com/${repo}/releases/download/v${version}/${file} --output ${file}
tar xzf ${file}
ls -la ${name}/

mkdir -p /usr/share/mdproxy4cs/

cp ${name}/mdproxy4cs         /usr/bin/
cp ${name}/pre-start.sh       /usr/share/mdproxy4cs/pre-start.sh
cp ${name}/mdproxy4cs.service /etc/systemd/system/
cp ${name}/default            /etc/default/mdproxy4cs

systemctl enable /etc/systemd/system/mdproxy4cs.service

rm -rf ${dir}
