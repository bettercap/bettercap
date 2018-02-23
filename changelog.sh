#!/bin/bash

NEW=()
FIXES=()
MISC=()

echo "@ Fetching remote tags ..."

git fetch --tags > /dev/null

CURTAG=$(git describe --tags --abbrev=0)
OUTPUT=$(git log $CURTAG..HEAD --oneline)
# https://stackoverflow.com/questions/19771965/split-bash-string-by-newline-characters
IFS=$'\n' LINES=($OUTPUT)

for LINE in "${LINES[@]}"; do
    LINE=$(echo "$LINE" | sed -E "s/^[[:xdigit:]]+\s+//")
    if [[ $LINE = *"new:"* ]]; then
        NEW+=("$LINE")
    elif [[ $LINE = *"fix:"* ]]; then
        FIXES+=("$LINE") 
    elif [[ $LINE != *"i did not bother commenting"* ]]; then 
        echo "MISC LINE =$LINE"
        MISC+=("$LINE")
    fi
done

echo
echo "Changelog"
echo "==="
echo

if [ -n "$NEW" ]; then
    echo "**New Features**"
    echo
    for l in "${NEW[@]}"
    do
        echo "* $l"
    done
fi

if [ -n "$FIXES" ]; then
    echo "**Fixes**"
    echo
    for l in "${FIXES[@]}"
    do
        echo "* $l"
    done
fi

if [ -n "$MISC" ]; then
    echo "**Misc**"
    echo
    for l in "${MISC[@]}"
    do
        echo "* $l"
    done
fi

echo
