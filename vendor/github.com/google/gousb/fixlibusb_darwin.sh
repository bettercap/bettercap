#!/bin/bash

function die {
  echo "$@"
  exit 1
}

FILE="$1"
if [[ -z "$FILE" ]]; then
  die "Usage: $0 <path to libusb.h>"
fi

if [[ $(gcc --version | grep -i "llvm") == "" ]]; then
  die "Error: This change is unnecessary unless your gcc uses llvm"
fi

BACKUP="${FILE}.orig"
if [[ -f "$BACKUP" ]]; then
  die "It looks like you've already run this script ($BACKUP exists)"
fi

cp $FILE $BACKUP || die "Could not create backup"

{
  echo 'H'                                  # Turn on error printing
  echo 'g/\[0\].*non-standard/s/\[0\]/[1]/' # Use [1] instead of [0] so the size is unambiguous
  echo 'g/\[.\].*non-standard/p'            # Print the lines changed
  echo 'w'                                  # Write output
} | ed $FILE
