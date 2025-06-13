#!/bin/bash

for i in $(seq 2 10); do 
  docker start fake-server-$i
done
