# Twelve Labs Demo: Video Processing

A scalable video processing demo platform built with Go. This project demonstrates a modern microservices architecture using a RESTful API for video uploads and a background worker for processing video data with AWS integration, containerization with kubernetes, and observability using Prometheus and Grafana. This was built purely for demo purposes, and while functional, is not fully ready for a production environment.

## Overview

This Video api and microservice acrhitecture is designed to leverage modern cloud based technologies, and to be fast, scalable and reliable.

-   **RESTful API** built using [Gin](https://github.com/gin-gonic/gin) for handling video uploads and retrieval. Defined by OpenAPI specs.
-   **Background Worker** that processes video files using ffmpeg, integrated with AWS services S3, SQS, and DynamoDB.
-   **Observability & Monitoring** using Prometheus for metrics collection and Grafana for dashboard visualization.
-   **Infrastructure as Code** with Terraform, containerization with Docker, and deployment on AWS EKS.
-   **CI/CD Pipelines** using GitHub Actions.

## Features

-   **Video API:**
    -   POST `/videos` to upload videos.
    -   GET `/videos/:id` to retrieve video metadata.
-   **Worker:**
    -   Polls AWS SQS to process video files.
    -   Uses ffmpeg for video transcoding.
    -   Uses DynamoDB to update video metadata
-   **Monitoring:**
    -   Custom Prometheus metrics for both API and worker.
    -   Grafana dashboards to visualize HTTP request metrics and worker processing performance.
-   **Deployment:**
    -   Terraform to provision AWS infrastructure.
    -   Kubernetes manifests and Helm charts for container orchestration.
    -   CI/CD automation using GitHub Actions.

## Architecture

The system is composed of two main microservices:

1. **API Service:**

    - Handles video uploads and metadata retrieval.
    - Exposes Prometheus metrics at `/metrics`.
    - Accessible via a Kubernetes LoadBalancer service
    - Security group to whitelist ips

2. **Worker Service:**
    - Processes queued video tasks from AWS SQS.
    - Uses ffmpeg for transcoding and updates custom worker metrics.

Both services are containerized and deployed on an EKS cluster. Monitoring is achieved by scraping metrics via ServiceMonitor/PodMonitor configurations, and dashboards are available in Grafana.

Note: given the purely demo nature of this application, security was implemented as a bare-bones measure. With more time, I would implement a more robust security strategy by building an ingress controller and looking into authentication and authorization using OAuth or something similarly secure.

## Prerequisites

-   [Go 1.24+](https://golang.org/)
-   [Docker](https://www.docker.com/) (with Buildx for ARM64 builds)
-   [Terraform 1.10+](https://www.terraform.io/)
-   [kubectl](https://kubernetes.io/docs/tasks/tools/)
-   [Helm](https://helm.sh/)
-   AWS account with proper credentials

## Installation & Local Setup

1. **Clone the Repository:**

```bash
   git clone https://github.com/yourusername/twelve-labs-demo.git
   cd twelve-labs-demo
```

2. **Build API and worker:**

```bash
    cd cmd/api
    go build -o twelve-labs-demo-api .

    cd ../video_processor
    go build -o twelve-labs-video-processor .
```

3. **Run Unit Tests:**

```bash
   go test ./...
```

4. **Run Locally (API):**

```bash
   cd cmd/api
   ./twelve-labs-demo-api
```

## Deployment

### Docker & ECR

Build and push Docker images using Docker Buildx (for ARM64 builds):

```bash
    docker buildx build --platform linux/arm64 -f cmd/api/Dockerfile -t <your-ecr-repo>/twelve-labs/api:latest --push .
    docker buildx build --platform linux/arm64 -f cmd/video_processor/Dockerfile -t <your-ecr-repo>/twelve-labs/video-processor:latest --push .
```

### Infrastructure with Terraform

```bash
    cd infrastructure/terraform
    terraform init --backend-config=backend.s3.tfbackend
    terraform apply --var-file=dev.tfvars
```

note: to deploy to your own aws account, will need to set up an s3 bucket for terraform backend, switch out account specific vars

### Kubernetes & Helm

Deploy your services to an EKS cluster using Helm:

```bash
# Deploy the monitoring stack (Prometheus & Grafana)
helm upgrade --install prometheus-grafana prometheus-community/kube-prometheus-stack --namespace monitoring

# Deploy the application (API and Worker)
helm upgrade --install twelve-labs-demo ./charts/twelve-labs-demo --namespace default
```

Note: given more time I would recommend configuring ArgoCD to deploy the helm charts utilizing a GitOps flow

## Monitoring & Grafana Dashboards

### Prometheus Metrics

-   **API Metrics:**

    -   `myapp_http_requests_total`
    -   `myapp_http_request_duration_seconds`

-   **Worker Metrics:**
    -   `worker_processing_duration_seconds`
    -   `worker_errors_total`

### Grafana Setup

Grafana is deployed as part of the kube-prometheus-stack. To view dashboards, port-forward to Grafana:

```bash
kubectl port-forward svc/prometheus-grafana -n monitoring 3000:80
```

## Future Enhancements/Caveats

-   **More robust unit testing:**

    Without much experience writing unit tests in Go, putting together unit tests that effectively covered base cases in my files was a challenge. Given more time, I would look into unit testing further to ensure robust testing.

-   **Kubernetes Infrastructure**

    For demo purposes I went with a simpler kubernetes setup than I would have if I were configuring for a production environment.

    One of the enhancements I would make is to establish and ingress controller. For simplicities sake, I deployed my API service with a LoadBalancer service, but this isnt as scalable or configurable a solution as an ingress controller.

    I would also instrument my kubernetes cluster with Prometheus and Grafana on cluster rollout rather with a Persistent Volume rather than with a helm chart ad hoc. This is a more robust solution because on cluster restart the current grafana dashboards would get cleared.

    Similarly I would instrument my cluster with ArgoCD to make deploying changes to my helm deployment more conventional

    Currently I have my cluster running with BOTTLEROCKET_ARM_64 spot instances on a minimal t4g.medium instance type. This was configured for cost efficiency for a demo environment. In a production environment I would instrument the cluster with karpenter for increased scalability and reliability.

-   **Decoupling Api from VideoProcessing Worker**

    For simplicities sake, I configured my repo to have both the video processing worker and the api together, but in a production environment I would write these in separate repos. Decoupling is a better approach for a number of reasons including independent development and deployment of the services (when you deploy one you dont have to deploy the other), improved maintainability (codebase becomes unwieldy if it contains multiple services as it grows), and in general better adherence to the principle of separation of concerns.

-   **Real AI Integration**

    For the purposes of this demo, I didnt include any real AI integration into the VideoProcessing service because that would have increased the scope of the project dramatically. AI models are an area of interest for me, and I would be interested to learn more about what it would take to integrate with a real AI model.
