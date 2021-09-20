#!/bin/sh -e

if [ -n "${HELM_LINTER_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

version="$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)"
echo "Downloading and installing helm3-monitor v${version} ..."

url="https://github.com/yq314/helm3-monitor/releases/download/v${version}/helm3-monitor_${version}"
if [ "$(uname)" = "Darwin" ]; then
    if [ "$(uname -m)" = "arm64" ]; then
        url="${url}_darwin_arm64.tar.gz"
    else
        url="${url}_darwin_amd64.tar.gz"
    fi
elif [ "$(uname)" = "Linux" ] ; then
    if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
        url="${url}_linux_arm64.tar.gz"
    else
        url="${url}_linux_amd64.tar.gz"
    fi
else
    url="${url}_windows_amd64.tar.gz"
fi

echo "$url"

#mkdir -p "bin"
#mkdir -p "releases/v${version}"
#
#if [ -x "$(which curl 2>/dev/null)" ]; then
#    curl -sSL "${url}" -o "releases/v${version}.tar.gz"
#else
#    wget -q "${url}" -O "releases/v${version}.tar.gz"
#fi
#tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
mv "releases/v${version}/helm3-monitor" "bin/helm3-monitor" || \
    mv "releases/v${version}/helm3-monitor.exe" "bin/helm3-monitor"
mv "releases/v${version}/completion.yaml" .
mv "releases/v${version}/plugin.yaml" .
mv "releases/v${version}/README.md" .
mv "releases/v${version}/LICENSE" .
