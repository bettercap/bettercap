#!/usr/bin/env bash

set -eu

PROGRAM="${1}"
shift
COMMAND="${*}"

IMAGE="https://downloads.raspberrypi.org/raspbian_lite/images/raspbian_lite-2020-02-14/2020-02-13-raspbian-buster-lite.zip"
GOLANG="https://golang.org/dl/go1.16.2.linux-armv6l.tar.gz"

REPO_DIR="${PWD}"
TMP_DIR="/tmp/builder"
MNT_DIR="${TMP_DIR}/mnt"

if ! systemctl is-active systemd-binfmt.service >/dev/null 2>&1; then
  mkdir -p "/lib/binfmt.d"
  echo ':qemu-arm:M::\x7fELF\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x28\x00:\xff\xff\xff\xff\xff\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00\x00\xfe\xff\xff\xff:/usr/bin/qemu-arm-static:F' > /lib/binfmt.d/qemu-arm-static.conf
  systemctl restart systemd-binfmt.service
fi

mkdir -p "${TMP_DIR}"
wget --show-progress -qcO "${TMP_DIR}/raspbian.zip" "${IMAGE}"
gunzip -c "${TMP_DIR}/raspbian.zip" > "${TMP_DIR}/raspbian.img"
truncate "${TMP_DIR}/raspbian.img" --size=+2G
parted --script "${TMP_DIR}/raspbian.img" resizepart 2 100%

LOOP_PATH="$(losetup --find --partscan --show "${TMP_DIR}/raspbian.img")"
e2fsck -y -f "${LOOP_PATH}p2"
resize2fs "${LOOP_PATH}p2"
partprobe "${LOOP_PATH}"

mkdir -p "${MNT_DIR}"
mountpoint -q "${MNT_DIR}" && umount -R "${MNT_DIR}"
mount -o rw "${LOOP_PATH}p2" "${MNT_DIR}"
mount -o rw "${LOOP_PATH}p1" "${MNT_DIR}/boot"

mount --bind /dev "${MNT_DIR}/dev/"
mount --bind /sys "${MNT_DIR}/sys/"
mount --bind /proc "${MNT_DIR}/proc/"
mount --bind /dev/pts "${MNT_DIR}/dev/pts"
mount | grep "${MNT_DIR}"
df -h

cp /usr/bin/qemu-arm-static "${MNT_DIR}/usr/bin"
cp /etc/resolv.conf "${MNT_DIR}/etc/resolv.conf"

mkdir -p "${MNT_DIR}/root/src/${PROGRAM}"
mount --bind "${REPO_DIR}" "${MNT_DIR}/root/src/${PROGRAM}"

cp "${MNT_DIR}/etc/ld.so.preload" "${MNT_DIR}/etc/_ld.so.preload"
touch "${MNT_DIR}/etc/ld.so.preload"

chroot "${MNT_DIR}" bin/bash -x <<EOF
set -eu

export LANG="C"
export LC_ALL="C"
export LC_CTYPE="C"
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/root/bin"

wget --show-progress -qcO /tmp/golang.tar.gz "${GOLANG}"
tar -C /usr/local -xzf /tmp/golang.tar.gz
export GOROOT="/usr/local/go"
export GOPATH="/root"

apt-get -y update
apt-get install wget libpcap-dev libusb-1.0-0-dev libnetfilter-queue-dev build-essential git

cd "/root/src/${PROGRAM}"
${COMMAND}
EOF
echo "Build finished"
