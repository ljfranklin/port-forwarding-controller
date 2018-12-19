#!/bin/bash

set -eu

project_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd )"

pushd "${project_dir}" > /dev/null
  if [ "$(uname -s)" == "Darwin" ]; then
    platform="darwin"
  else
    platform="linux"
  fi

  wget -O /tmp/kubebuilder.tgz "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v1.0.6/kubebuilder_1.0.6_${platform}_amd64.tar.gz"
  tar xvf /tmp/kubebuilder.tgz --strip-components=1
  rm /tmp/kubebuilder.tgz

  wget -O ./bin/kustomize "https://github.com/kubernetes-sigs/kustomize/releases/download/v1.0.11/kustomize_1.0.11_${platform}_amd64"
  chmod +x ./bin/kustomize

  # required to build cross-platform images
  # skipping this step results in 'standard_init_linux.go:190: exec user process caused "exec format error"'
  docker run --rm --privileged multiarch/qemu-user-static:register --reset

  if ! grep -q experimental "$HOME/.docker/config.json"; then
    echo 'Manual Step: add `"experimental": "enabled"` to `$HOME/.docker/config.json` to build cross-platform images'
  fi

  echo ""
  echo "Done!"
popd > /dev/null
