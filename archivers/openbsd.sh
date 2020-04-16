#!/bin/sh

if [ -z "$1" ] || [ -z "$2" ]; then
	echo "usage: $0 <url> <output path>"
	exit 1
fi

curl -s -o"$2" "$1"
