#!/bin/bash

for i in $(seq 2 10); do 
  docker stop fake-server-$i
done
