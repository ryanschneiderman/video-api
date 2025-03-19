provider "aws" {
  region = "us-east-1"

  default_tags {
    tags = {
      Environment = "dev"
      Project     = "video-api"
    }
  }
}

locals {
  service_account_name = "twelve-labs-sa"
}

# S3 Bucket for Video Storage
resource "aws_s3_bucket" "video_storage" {
  bucket = "${var.aws_account}-twelve-labs-video-storage"
}

resource "aws_ecr_repository" "twelve_labs_api" {
  name                 = "twelve-labs/api"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "twelve_labs_video_processor" {
  name                 = "twelve-labs/video-processor"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}


resource "aws_dynamodb_table" "videos" {
  name         = "twelve-labs-videos"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "video_id"

  attribute {
    name = "video_id"
    type = "S"
  }
}

# Create a dead-letter queue
resource "aws_sqs_queue" "dlq" {
  name = "video-processing-dlq"
}


# SQS Queue for Processing Events
resource "aws_sqs_queue" "video_processing_queue" {
  name                       = "video-processing-queue"
  delay_seconds              = 0
  visibility_timeout_seconds = 900
  message_retention_seconds  = 86400
  max_message_size           = 262144

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn,
    maxReceiveCount     = 5 # After 5 receives, message goes to DLQ
  })
}
