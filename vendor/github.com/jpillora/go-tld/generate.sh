#!/bin/sh

#pull the official TLD list, remove comments and blanks, reverse each line, then sort
words=$(curl -# https://publicsuffix.org/list/public_suffix_list.dat \
	| grep -v "^//" \
	| grep -v "^\$" \
	| grep -v "^!" \
	| grep -v "^*" \
	| rev \
	| sort)

#convert each line into Go strings
strings=$(for w in $words; do
	echo "	\"$w\","
done)

#output the generated file
echo "package tld
//generated on '$(date -u)'

//list contains all TLDs reversed, then sorted
var list = []string{
$strings
}

var count = len(list)
" >parse_list.go
