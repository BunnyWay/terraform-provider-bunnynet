#!/bin/bash

VERSION="$1"
PLATFORMS=("darwin/amd64" "darwin/arm64" "freebsd/386" "freebsd/amd64" "freebsd/arm" "freebsd/arm64" "linux/386" "linux/amd64" "linux/arm" "linux/arm64" "windows/386" "windows/amd64" "windows/arm" "windows/arm64")

# upload artifacts
for FILE in terraform-provider-bunny_*; do
  echo "Uploading \"$FILE\""
  curl -H "AccessKey: ${STORAGE_PASSWORD}" "https://${STORAGE_HOST}/${STORAGE_ZONE}/terraform-provider-bunny/${VERSION}/${FILE}" -X PUT --data-binary "@${PWD}/${FILE}"
done

PLATFORM_VERSIONS="["
# generate JSONs
for PLATFORM in "${PLATFORMS[@]}"; do
  OS=$(echo "$PLATFORM" | cut -d '/' -f 1)
  ARCH=$(echo "$PLATFORM" | cut -d '/' -f 2)
  FILE="terraform-provider-bunny_${VERSION}_${OS}_${ARCH}.zip"
  SHASUM=$(shasum -a 256 $FILE | awk '{print $1}')

  CONTENTS=$(cat .github/scripts/platform.json.template | sed "s/{{VERSION}}/${VERSION}/g" | sed "s/{{OS}}/${OS}/g" | sed "s/{{ARCH}}/${ARCH}/g" | sed "s/{{SHASUM}}/${SHASUM}/g")
  curl -H "AccessKey: ${STORAGE_PASSWORD}" "https://${STORAGE_HOST}/${STORAGE_ZONE}/v1/providers/bunny/bunny/${VERSION}/download/${OS}/${ARCH}" -X PUT --data-binary "${CONTENTS}"

  PLATFORM_VERSIONS="${PLATFORM_VERSIONS}{\"os\":\"${OS}\",\"arch\":\"${ARCH}\"},"
done

# /v1/providers/bunny/bunny/versions
PLATFORM_VERSIONS="${PLATFORM_VERSIONS%,}]"
CONTENTS=$(cat .github/scripts/versions.json.template | sed "s/{{VERSION}}/${VERSION}/g" | sed "s/{{PLATFORMS}}/${PLATFORM_VERSIONS}/g")
curl -H "AccessKey: ${STORAGE_PASSWORD}" "https://${STORAGE_HOST}/${STORAGE_ZONE}/v1/providers/bunny/bunny/versions" -X PUT --data-binary "${CONTENTS}"

# /v1/providers/
CONTENTS=$(cat .github/scripts/providers.json.template | sed "s/{{VERSION}}/${VERSION}/g")
curl -H "AccessKey: ${STORAGE_PASSWORD}" "https://${STORAGE_HOST}/${STORAGE_ZONE}/v1/providers/index.html" -H 'Override-Content-Type: application/json' -X PUT --data-binary "${CONTENTS}"
