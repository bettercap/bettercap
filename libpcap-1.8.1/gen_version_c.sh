#! /bin/sh
echo '#include <pcap/export-defs.h>' > "$2"
echo 'PCAP_API_DEF' >> "$2"
if grep GIT "$1" >/dev/null; then
	read ver <"$1"
	echo $ver | tr -d '\012'
	date +_%Y_%m_%d
else
	cat "$1"
fi | sed -e 's/.*/char pcap_version[] = "&";/' >> "$2"

