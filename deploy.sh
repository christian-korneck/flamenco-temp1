#!/bin/bash

MY_DIR="$(dirname "$(readlink -e "$0")")"
ADDON_ZIP="$MY_DIR/web/static/flamenco3-addon.zip"

set -e

function prompt() {
  echo
  echo -------------------
  printf " \033[38;5;214m$@\033[0m\n"
  echo -------------------
  echo
}

prompt "Building Flamenco"
make

prompt "Deploying Manager"
ssh -o ClearAllForwardings=yes flamenco.farm.blender -t sudo systemctl stop flamenco3-manager
scp flamenco-manager flamenco.farm.blender:/home/flamenco3/
ssh -o ClearAllForwardings=yes flamenco.farm.blender -t sudo systemctl start flamenco3-manager

prompt "Deploying Worker"
cp -f flamenco-worker /shared/software/flamenco3-worker

prompt "Deploying Blender Add-on"
rm -rf /shared/software/addons/flamenco
pushd /shared/software/addons
unzip -q "$ADDON_ZIP"
popd

prompt "Done!"
echo "Deployment done, be sure to restart all the Workers and poke Artists to reload their Blender add-on."
echo
