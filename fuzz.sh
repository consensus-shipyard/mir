#!/bin/bash

set -e

fuzzTime=${1:-10}

files=$(grep -r --include='**_fuzz_test.go' --files-with-matches 'func Fuzz' .)

for file in ${files}
do
	funcs=$(grep -oE "Fuzz[a-zA-Z]+" $file)
	for func in ${funcs}
	do
		echo "Fuzzing $func in $file"
		parentDir=$(dirname $file)
		go test $parentDir -run=$func -fuzz=$func -fuzztime=${fuzzTime}s
	done
done