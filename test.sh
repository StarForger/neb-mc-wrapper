#!/usr/bin/env bash

function codeTest() {
    local -r actual=$1
    local -r expected=$2

    if [[ "${actual}" != ${expected} ]]; then
        echo "expected = ${expected}, actual = ${actual}"
        exit 1
    fi
}

go run main.go
go run main.go -h
# codeTest $? 0
go run main.go -x
codeTest $? 1
go run main.go -v ./test/env.sh
codeTest $? 0
go run main.go -shell=sh ./test/bash.sh
codeTest $? 1
# go run main.go test/bash.sh
# go run main.go test/exit.sh 1