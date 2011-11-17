#!/bin/bash
make install;
6g -o test.out test.go && 6l -o test test.out;
