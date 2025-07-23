#!/bin/bash

docker build -t fake-server-image .

./create_fake_server.sh