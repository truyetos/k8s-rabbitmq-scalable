#!/bin/bash


# Install helm to your machine before running this script
# https://helm.sh/docs/intro/install/

kubectl create ns rabbitmq

helm repo add bitnami https://charts.bitnami.com/bitnami
helm -n rabbitmq install rabbitmq-ops bitnami/rabbitmq-cluster-operator

kubectl -n rabbitmq create configmap definitions --from-file=definitions.json=./rabbitmq/definitions.json

kubectl -n rabbitmq apply -f ./rabbitmq/deployment.yaml