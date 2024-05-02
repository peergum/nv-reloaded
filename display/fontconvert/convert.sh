#!/bin/bash

mkdir -p ../Fonts ../fonts-go
sizes="9 12 16 20 24 30"
find ImportedFonts -type f|
grep -E -e '.ttf|.otf'|
while read f; do
	for size in $sizes; do
		of="../Fonts/$(echo "$f"|cut -d. -f1|cut -d/ -f2)${size}pt8b.h"
		./fontconvert "$f" $size 255 > "$of"
    of="../fonts-go/$(echo "$f"|cut -d. -f1|cut -d/ -f2)${size}pt8b.go"
		./fontconvert2go "$f" $size 255 > "$of"
		cp definitions.go ../fonts-go/
	done
done
