#!/usr/bin/env bash

function assert() {
    local -r expected=${1}
    shift
    echo -e ""
    eval "${@}"
    local actual=$?
    if [[ "${actual}" != "${expected}" ]]; then
        echo "test failed: expected ${expected}, got ${actual}"
        exit 1
    fi
    echo "test passed"
}


assert 0 go run main.go -h
assert 0 go run main.go -v ./test/env.sh
assert 0 go run main.go test/sleep.sh
assert 0 go run main.go test/exit.sh 0

assert 1 go run main.go
assert 1 go run main.go -x
# assert 1 go run main.go -shell=sh ./test/bash.sh
assert 1 go run main.go test/exit.sh 1