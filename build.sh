#!/usr/bin/bash

go build -o bin/zistd
go build -o zist
go build -o bin/zistcl github.com/ziscky/zist/zistcl

echo "run ./zist install to install zistd and zistcl."