#!/usr/bin/env bash

for file in `git ls-files | egrep '.*\.go$'`; do
    gofmt -tabs=false -tabwidth 4 -w $file
done
