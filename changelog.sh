#!/bin/bash

NEW=()
FIXES=()
MISC=()

echo "@ Fetching remote tags ..."

# git fetch --tags > /dev/null

CURTAG=$(git describe --tags --abbrev=0)
OUTPUT=$(git log $CURTAG..HEAD --oneline)
IFS=$'\n' LINES=($OUTPUT)

for LINE in "${LINES[@]}"; do
    LINE=$(echo "$LINE" | sed -E "s/^[[:xdigit:]]+\s+//")
    if [[ $LINE = *"new:"* ]]; then
        LINE=$(echo "$LINE" | sed -E "s/^new: //")
        NEW+=("$LINE")
    elif [[ $LINE = *"fix:"* ]]; then
        LINE=$(echo "$LINE" | sed -E "s/^fix: //")
        FIXES+=("$LINE") 
    elif [[ $LINE != *"i did not bother commenting"* ]] && [[ $LINE != *"Merge "* ]]; then 
        echo "MISC LINE =$LINE"
        LINE=$(echo "$LINE" | sed -E "s/^[a-z]+: //")
        MISC+=("$LINE")
    fi
done

echo
echo "Changelog"
echo "==="

if [ -n "$NEW" ]; then
    echo
    echo "**New Features**"
    echo
    for l in "${NEW[@]}"
    do
        echo "* $l"
    done
fi

if [ -n "$FIXES" ]; then
    echo
    echo "**Fixes**"
    echo
    for l in "${FIXES[@]}"
    do
        echo "* $l"
    done
fi

if [ -n "$MISC" ]; then
    echo
    echo "**Misc**"
    echo
    for l in "${MISC[@]}"
    do
        echo "* $l"
    done
fi

echo
