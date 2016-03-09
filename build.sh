#!/usr/bin/bash

go get github.com/gorilla/mux
go get github.com/BurntSushi/toml

go build -o bin/zistd
go build -o zist
go build -o bin/zistcl github.com/ziscky/zist/zistcl

echo "Run ./zist install to install zistd and zistcl."