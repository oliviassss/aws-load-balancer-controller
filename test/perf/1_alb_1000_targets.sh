#!/bin/bash

source ./utils.sh

main() {
  numLBs=1
  numTargets=1000
  create_alb $numLBs $numTargets

  echo "Press any key to proceed with resource deletion"
  read anykey
  for i in $(seq 1 $numLBs); do
    echo "deleting deployment, svc and ingress resources"
    kubectl delete -f alb_$i.yaml
    rm alb_$i.yaml
  done
}

main $@