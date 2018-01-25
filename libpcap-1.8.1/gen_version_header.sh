#! /bin/sh
print_version_string()
{
	if grep GIT "$1" >/dev/null
	then
		read ver <"$1"
		echo $ver | tr -d '\012'
		date +_%Y_%m_%d
	else
		cat "$1"
	fi
}
if test $# != 3
then
	echo "Usage: gen_version_header.sh <version file> <template> <output file>" 1>&2
	exit 1
fi
version_string=`print_version_string "$1"`
sed "s/%%LIBPCAP_VERSION%%/$version_string/" "$2" >"$3"
