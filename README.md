# Message Queue Scalable

## Overview

This project is designed to address the requirements outlined in the provided document. The primary goal is to deploy a scalable message queue system using Kubernetes (k8s), RabbitMQ, and KEDA. The project is structured to facilitate easy deployment and scaling based on message rates.

## Technical 

- Docker : support build container image
- Kubernetes : support orchestration and scaling
- Golang : build example service.

## Assumptions

- Kubernetes (k8s) is already deployed and configured in your environment.

## Project Structure

### `infas` Folder

The `infas` folder contains the necessary configurations to deploy RabbitMQ and KEDA. These tools are essential for managing and scaling the pods based on the message rate.

- **RabbitMQ**: A message broker that enables communication between different services.
- **KEDA**: Kubernetes-based Event Driven Autoscaler that scales pods based on the rate of incoming messages.

The `infas` folder also contains two shell scripts to facilitate the deployment of RabbitMQ and KEDA:

- **deploy-rabbitmq.sh**: This script automates the deployment of RabbitMQ in your Kubernetes cluster.
- **deploy-keda.sh**: This script automates the deployment of KEDA in your Kubernetes cluster.

These scripts simplify the deployment process, ensuring that RabbitMQ and KEDA are set up correctly with minimal manual intervention.

### `services` Folder

The `services` folder is divided into three main components:

1. **Consumer**:
  - Contains example code for the consumer service.
  - Includes a Dockerfile to build the consumer image.
  - Consumer is simulated a batch read 10 messages in queue
  - It handle message using print and sleep 200ms then ack the message.
  - When message is acked, the consumer will receive new message.

2. **Producer**:
  - Contains example code for the producer service.
  - Includes a Dockerfile to build the producer image.
  - Producer publish the message 20 messages per second and lifetime is defined in deployment (default 120 seconds).

3. **.deployment**:
  - Contains deployment configurations for both the consumer and producer services.

The container image has already deploy to my docker hub account and the deployment is using that.

## Deployment

### Setup infrastructures

- Go to folder `infras`
- Run file `rabbitmq.sh` and `keda.sh`

### Service Deployment (Command line)

#### Build Service
**Consumer Service**
- Go to folder `services/consumer`
- Run `docker buildx build --platform linux/amd64,linux/arm64 -t {tagname} .` with `{tagname}` is replaced by your image name. You may need to push the image to a repository if you have remote deployment. In my case, I pushed to docker hub with name `truyet/consumer:1.0`

**Producer Service**
- Go to folder `services/producer`
- Run `docker buildx build --platform linux/amd64,linux/arm64 -t {tagname} .` with `{tagname}` is replaced by your image name. You may need to push the image to a repository if you have remote deployment. In my case, I pushed to docker hub with name `truyet/producer:1.0`

#### Deploy service
**Consumer**
- Go to `services/.deployment`
- Run `kubectl apply -f consumer.yml` to deploy consumer

**Producer**
- Go to `services/.deployment`
- Run `kubectl apply -f producer.yml` to deploy producer.
- Producer is deployed as a cron job run each 5 minutes.



## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [RabbitMQ Documentation](https://www.rabbitmq.com/documentation.html)
- [KEDA Documentation](https://keda.sh/docs/)
- [Docker Documentation](https://docs.docker.com/manuals/)
- [Golang Document](https://go.dev/)
