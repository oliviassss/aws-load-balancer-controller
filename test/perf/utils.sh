#!/bin/bash
#set -o errexit
#set -o nounset
#set -o pipefail

lbName=""
lbArn=""
tgtGrpArn=""
tgHealth=""

echo_time() {
    date +"%D %T $*"
}

generate_alb_manifest() {
echo "generating manifest $1..."
#echo "deployment name $1, service name $2, number of targets $3"
cat <<EOF > alb_$1.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-$1
spec:
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
    type: RollingUpdate
  replicas: $2
  selector:
    matchLabels:
      app: my-app-$1
  template:
    metadata:
      labels:
        app: my-app-$1
        ports: multi
    spec:
      containers:
        - name: multi
          imagePullPolicy: Always
          image: "kishorj/hello-multi:v1"
          ports:
            - name: http
              containerPort: 80
EOF

cat <<EOF >> alb_$1.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: my-svc-$1
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb-ip"
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
spec:
  selector:
    app: my-app-$1
  type: NodePort
  ports:
EOF

for i in $(seq 1 1); do
  cat <<EOF >> alb_$1.yaml
  - name: port
    port: $((80 + i))
    targetPort: $((8080 + i - 1))
    protocol: TCP
EOF
done

cat <<EOF >> alb_$1.yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress-$1
  annotations:
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
spec:
  ingressClassName: alb
  rules:
    - http:
        paths:
        - path: /abc
          pathType: Prefix
          backend:
            service:
              name: my-svc-$1
              port:
                name: port
EOF

#for i in $(seq 1 $2); do
#  pathPrefix="path"
#  portPrefix="port"
#  pathName="/$pathPrefix"$i
#  portName="$portPrefix-"$i
#  cat <<EOF >> manifest_$1.yaml
#        - path: $pathName
#          pathType: Prefix
#          backend:
#            service:
#              name: my-svc-$1
#              port:
#                name: $portName
#EOF
#done
}

generate_nlb_manifest() {
echo "generating nlb manifest $1..."
cat <<EOF > nlb_$1.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app-$1
spec:
  strategy:
    rollingUpdate:
      maxSurge: 100%
      maxUnavailable: 25%
    type: RollingUpdate
  replicas: $2
  selector:
    matchLabels:
      app: my-app-$1
  template:
    metadata:
      labels:
        app: my-app-$1
        ports: multi
    spec:
      containers:
        - name: multi
          imagePullPolicy: Always
          image: "kishorj/hello-multi:v1"
          ports:
            - name: http
              containerPort: 80
EOF

cat <<EOF >> nlb_$1.yaml
---
apiVersion: v1
kind: Service
metadata:
  name: my-svc-$1
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb-ip"
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
spec:
  selector:
    app: my-app-$1
  type: LoadBalancer
  ports:
    - name: port-$1
      port: 80
      targetPort: 8080
      protocol: TCP
EOF

#for i in $(seq 1 $2); do
#  cat <<EOF >> nlb_$1.yaml
#    - name: port-$i
#      port: 80
#      targetPort: $((8080 + i - 1))
#      protocol: TCP
#EOF
#done
}

create_alb(){
  numLBs=$1
  numTargets=$2
  echo_time "Starting Test, create $numLBs ALB with $numTargets Targets each"
  startTime=$(date +%T)
  for j in $(seq 1 $numLBs); do
    generate_alb_manifest $j $numTargets
    echo "creating deployment, svc, ingress $j"
    kubectl apply -f alb_$j.yaml
  done

  for i in $(seq 1 $numLBs); do
    get_alb_name my-ingress-$i
    echo_time "Deployed loadbalancer $lbName"
    query_lb_arn $lbName

    get_target_group_arn $lbArn
    echo_time "TargetGroup ARN $tgtGrpArn"

    check_target_group_health
  done

  endTime=$(date +%T)
  diff=$(datediff $startTime $endTime -f "%H hours, %M minutes, and %S seconds")
  echo "Total time to provision $numLBs ALB with $numTGs Targets each: $diff"

}

create_nlb(){
  numLBs=$1
  numTargets=$2
  echo_time "Starting Test, create $numLBs NLB with $numTargets targets each"
  startTime=$(date +%T)
  for j in $(seq 1 $numLBs); do
    echo $j
    generate_nlb_manifest $j $numTargets
    kubectl apply -f nlb_$1.yaml
  done

  # check creation
  for i in $(seq 1 $numLBs); do
    get_nlb_name my-svc-$j
    echo_time "Deployed loadbalancer $lbName"
    query_lb_arn $lbName

    get_target_group_arn $lbArn
    echo_time "TargetGroup ARN $tgtGrpArn"

    check_target_group_health
  done
  endTime=$(date +%T)
  diff=$(datediff $startTime $endTime -f "%H hours, %M minutes, and %S seconds")
  echo "Total time to provision $numLBs NLB with $numTargets Targets each: $diff"
}

get_alb_name() {
	NS=""
	if [ ! -z $2 ]; then
		NS="-n $2"
	fi
	echo_time "Looking up ingress $1 $NS"
	for i in $(seq 1 10); do
		lbName=$(kubectl get ingress $1 $NS -ojsonpath='{.status.loadBalancer.ingress[0].hostname}' |awk -F- {'print $1"-"$2"-"$3"-"$4'})
		if [ "$lbName" != "" ]; then
			break
		fi
		sleep 2
		echo -n "."
	done
	echo
	echo_time "$lbName"

}

get_nlb_name() {
	NS=""
	if [ ! -z $2 ]; then
		NS="-n $2"
	fi
	echo_time "Looking up service $1 $NS"
	for i in $(seq 1 10); do
		lbName=$(kubectl get svc $1 $NS -ojsonpath='{.status.loadBalancer.ingress[0].hostname}' |awk -F- {'print $1"-"$2"-"$3"-"$4'})
		if [ "$lbName" != "" ]; then
			break
		fi
		sleep 2
		echo -n "."
	done
	echo
	echo_time "$lbName"

}

query_lb_arn() {
	lbArn=$(aws elbv2 describe-load-balancers --names $1 | jq -j '.LoadBalancers | .[] | .LoadBalancerArn')
	echo_time "LoadBalancer ARN $lbArn"
}

get_target_group_arn() {
	tgtGrpArn=$(aws elbv2 describe-target-groups --load-balancer-arn $lbArn | jq -j ".TargetGroups[0].TargetGroupArn")
}

get_target_group_health() {
	tgHealth=$(aws elbv2 describe-target-health --target-group-arn $tgtGrpArn | jq -r '.TargetHealthDescriptions[] | [.TargetHealth.State][]')
}

check_target_group_health() {
	echo_time "Checking target group health "
	numreplicas=$1
	lastcount=0
	for i in  $(seq 1 60); do
		count=0
		echo -n "."
		get_target_group_health
		for status in $tgHealth; do
			let "count+=1"
		done
		if [ $count -ne $lastcount ]; then
			lastcount=$count
			echo_time "Got $count targets"
		fi

		something_else=0
		for status in $tgHealth; do
			if [[ "$status" != "healthy" && "$status" != "unhealthy" ]];then
				something_else=1
				break
			fi
		done
		if [ $something_else -eq 0 ]; then
			if [ -z $numreplicas ]; then
				break
			fi
			if [ "$numreplicas" -eq "$count" ]; then
				break
			fi
		fi
		sleep 10
	done
	echo
	echo_time "Got $lastcount targets"
	echo_time "Target health"
	echo $tgHealth
}
