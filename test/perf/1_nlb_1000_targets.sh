#!/bin/bash

source ./utils.sh

main(){
  numLBs=1
  numTargets=1000
  create_nlb $numLBs $numTargets

  echo "Press any key to proceed with resource deletion"
  read anykey
  for i in $(seq 1 $numLBs); do
    echo "deleting deployment and nlb resources"
    kubectl delete -f nlb_$i.yaml
    rm nlb_$i.yaml
  done
}

main $@
