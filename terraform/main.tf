provider "aws" {
  region = "us-east-1"
}

# S3 Bucket for Video Storage
resource "aws_s3_bucket" "video_storage" {
  bucket = "twelve-labs-video-storage"
}

# S3 Bucket Policy (Allows Public Read for Testing, Change for Production)
resource "aws_s3_bucket_policy" "video_storage_policy" {
  bucket = aws_s3_bucket.video_storage.id
  policy = <<POLICY
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::${aws_s3_bucket.video_storage.id}/*"
      }
    ]
  }
  POLICY
}

# SQS Queue for Processing Events
resource "aws_sqs_queue" "video_processing_queue" {
  name                       = "video-processing-queue"
  delay_seconds              = 0
  visibility_timeout_seconds = 900
  message_retention_seconds  = 86400
  max_message_size           = 262144
}

# ECS Cluster
resource "aws_ecs_cluster" "video_api_cluster" {
  name = "video-api-cluster"
}

# # ECS Task Definition
# resource "aws_ecs_task_definition" "video_api_task" {
#   family                   = "video-api-task"
#   network_mode             = "awsvpc"
#   requires_compatibilities = ["FARGATE"]
#   cpu                      = "512"
#   memory                   = "1024"

#   container_definitions = jsonencode([
#     {
#       name      = "video-api"

#       image     = "your-dockerhub-username/video-api:latest"
#       essential = true
#       environment = [
#         { name = "S3_BUCKET", value = aws_s3_bucket.video_storage.id },
#         { name = "SQS_QUEUE", value = aws_sqs_queue.video_processing_queue.id }
#       ]
#       portMappings = [
#         {
#           containerPort = 8080
#           hostPort      = 8080
#         }
#       ]
#     }
#   ])
# }

# # ECS Service (Runs the API as a Load Balanced Service)
# resource "aws_ecs_service" "video_api_service" {
#   name            = "video-api-service"
#   cluster         = aws_ecs_cluster.video_api_cluster.id
#   task_definition = aws_ecs_task_definition.video_api_task.arn
#   launch_type     = "FARGATE"

#   network_configuration {
#     subnets          = ["subnet-12345678"] # Replace with your actual subnet
#     security_groups  = ["sg-12345678"]     # Replace with your actual security group
#     assign_public_ip = true
#   }
# }

# # API Gateway to Route Requests to ECS
# resource "aws_apigatewayv2_api" "video_api_gateway" {
#   name          = "VideoAPI"
#   protocol_type = "HTTP"
# }

# # API Gateway Integration (Routes Requests to ECS)
# resource "aws_apigatewayv2_integration" "video_api_integration" {
#   api_id                 = aws_apigatewayv2_api.video_api_gateway.id
#   integration_type       = "HTTP_PROXY"
#   integration_uri        = "http://your-load-balancer-url.com" # Replace with ECS Load Balancer URL
#   payload_format_version = "1.0"
# }

# # API Gateway Route for Video Uploads
# resource "aws_apigatewayv2_route" "video_upload_route" {
#   api_id    = aws_apigatewayv2_api.video_api_gateway.id
#   route_key = "POST /videos"
#   target    = "integrations/${aws_apigatewayv2_integration.video_api_integration.id}"
# }

# # API Gateway Deployment
# resource "aws_apigatewayv2_stage" "video_api_stage" {
#   api_id      = aws_apigatewayv2_api.video_api_gateway.id
#   name        = "dev"
#   auto_deploy = true
# }
