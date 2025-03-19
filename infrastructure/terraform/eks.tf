
data "aws_eks_cluster" "this" {
  provider = aws
  name     = "cluster1"
}

data "aws_eks_cluster_auth" "this" {
  name = data.aws_eks_cluster.this.name
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.this.endpoint
  token                  = data.aws_eks_cluster_auth.this.token
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.this.certificate_authority[0].data)
}

data "aws_iam_openid_connect_provider" "oidc" {
  url = data.aws_eks_cluster.this.identity[0].oidc[0].issuer
}

resource "aws_iam_role" "eks_pod_role" {
  name = "eks-pod-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = data.aws_iam_openid_connect_provider.oidc.arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${data.aws_iam_openid_connect_provider.oidc.url}:sub" = "system:serviceaccount:default:${local.service_account_name}"
          }
        }
      }
    ]
  })
}

data "aws_iam_policy_document" "s3_access" {
  statement {
    effect = "Allow"

    actions = [
      "s3:PutObject",
      "s3:GetObject"
    ]

    resources = [
      "arn:aws:s3:::498061775412-twelve-labs-video-storage/*"
    ]
  }
}

data "aws_iam_policy_document" "dynamodb_access" {
  statement {
    effect = "Allow"

    actions = [
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan"
    ]

    resources = [aws_dynamodb_table.videos.arn]
  }
}

data "aws_iam_policy_document" "sqs_access" {
  statement {
    effect = "Allow"

    actions = [
      "sqs:SendMessage",
      "sqs:GetQueueUrl",
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:ChangeMessageVisibility"
    ]

    resources = [aws_sqs_queue.video_processing_queue.arn]
  }
}

resource "aws_iam_role_policy" "s3_access" {
  role   = aws_iam_role.eks_pod_role.name
  policy = data.aws_iam_policy_document.s3_access.json
}

resource "aws_iam_role_policy" "dynamodb_access" {
  role   = aws_iam_role.eks_pod_role.name
  policy = data.aws_iam_policy_document.dynamodb_access.json
}

resource "aws_iam_role_policy" "sqs_access" {
  role   = aws_iam_role.eks_pod_role.name
  policy = data.aws_iam_policy_document.sqs_access.json
}

resource "kubernetes_service_account" "twelve_labs_sa" {
  provider = kubernetes
  metadata {
    name      = local.service_account_name
    namespace = "default"
    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.eks_pod_role.arn
    }
  }
}
