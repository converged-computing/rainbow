#!/bin/bash

for color in red blue yellow
  do
    docker exec -it cluster-${color} rainbow register --host scheduler:8080 --secret peanutbuttajellay --cluster-name cluster-${color}
done
