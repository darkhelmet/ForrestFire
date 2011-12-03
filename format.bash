#!/usr/bin/env bash

for file in `git ls-files | egrep -v '^h(5|tml)' | egrep '.*\.go$'`; do
    gofmt -tabindent=false -tabwidth 4 -w $file
done
