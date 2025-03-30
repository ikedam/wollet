#!/bin/sh

set -e
: ${VERSION:=v1.1.0}
: ${WORKDIR:=/tmp/wolnut}
: ${REPEAT:=10}
: ${SLEEP:=10}

if [ ! -d "${WORKDIR}" ]; then
  mkdir -p "${WORKDIR}"
fi

for i in $(seq "${REPEAT}"); do
  wget -q -O "${WORKDIR}/wolnut" "https://github.com/ikedam/wollet/releases/download/${VERSION}/wolnut-linux-mipsle-softfloat" && true
  if [ "${?}" = "0" ]; then
    break
  fi
  if [ "${i}" == "${REPEAT}" ]; then
    echo "give up!" >&2
    exit 1
  fi
  echo "fail(${i})...Retry after ${SLEEP} secs" >&2
  sleep "${SLEEP}"
done

chmod 755 "${WORKDIR}/wolnut"
cp "$(dirname "$0")/wolnut.yaml" "${WORKDIR}"

exec "${WORKDIR}/wolnut" "$@"
