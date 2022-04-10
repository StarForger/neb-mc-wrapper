#!/usr/bin/env bash

go build -a -o ./bin/neb-mc-wrapper

tar -czvf ./build/neb-mc-wrapper.tgz -C ./bin neb-mc-wrapper --remove-files