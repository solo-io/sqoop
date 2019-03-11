#!/bin/sh

set -eu


SQOOP_VERSIONS=$(curl -sH"Accept: application/vnd.github.v3+json" https://api.github.com/repos/solo-io/sqoop/releases | python -c "import sys; from json import loads as l; releases = l(sys.stdin.read()); print('\n'.join(release['tag_name'] for release in releases))")

if [ "$(uname -s)" = "Darwin" ]; then
  OS=darwin
else
  OS=linux
fi

for SQOOP_VERSION in $SQOOP_VERSIONS; do

tmp=$(mktemp -d /tmp/sqoop.XXXXXX)
filename="sqoopctl-${OS}-amd64"
url="https://github.com/solo-io/sqoop/releases/download/${SQOOP_VERSION}/${filename}"

if curl -f ${url} >/dev/null 2>&1; then
  echo "Attempting to download sqoopctl version ${SQOOP_VERSION}"
else
  continue
fi

(
  cd "$tmp"

  echo "Downloading ${filename}..."

  SHA=$(curl -sL "${url}.sha256" | cut -d' ' -f1)
  curl -sLO "${url}"
  echo "Download complete!, validating checksum..."
  checksum=$(openssl dgst -sha256 "${filename}" | awk '{ print $2 }')
  if [ "$checksum" != "$SHA" ]; then
    echo "Checksum validation failed." >&2
    exit 1
  fi
  echo "Checksum valid."
)

(
  cd "$HOME"
  mkdir -p ".sqoop/bin"
  mv "${tmp}/${filename}" ".sqoop/bin/sqoopctl"
  chmod +x ".sqoop/bin/sqoopctl"
)

rm -r "$tmp"

echo "Sqoop was successfully installed 🎉"
echo ""
echo "Add the sqoop CLI to your path with:"
echo "  export PATH=\$HOME/.sqoop/bin:\$PATH"
echo ""
exit 0
done

echo "No versions of sqoopctl found."
exit 1