terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  profile = "personal"
}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default_vpc_subnets" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

resource "aws_s3_bucket" "transactions" {
  bucket = var.s3_bucket_name
}

resource "aws_s3_bucket_versioning" "transactions" {
  bucket = aws_s3_bucket.transactions.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_security_group" "rds" {
  name        = "stori-rds-sg"
  description = "Allow Postgres from anywhere (demo only!)"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "stori" {
  name       = "stori-db-subnets"
  subnet_ids = data.aws_subnets.default_vpc_subnets.ids
}

resource "aws_db_instance" "stori" {
  identifier        = "stori-db"
  engine            = "postgres"
  engine_version    = "16.2"
  instance_class    = "db.t4g.micro"
  allocated_storage = 20

  username = var.db_username
  password = var.db_password
  db_name  = var.db_name
  port     = 5432

  publicly_accessible    = true
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.stori.name
}

resource "aws_iam_role" "lambda_exec" {
  name = "stori-lambda-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_s3" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
}

resource "aws_iam_role_policy_attachment" "lambda_ses" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSESFullAccess"
}

resource "aws_lambda_function" "s3_processor" {
  function_name = "stori-s3-processor"
  package_type  = "Image"
  image_uri     = var.ecr_s3_processor_image
  role          = aws_iam_role.lambda_exec.arn
  timeout       = 30
  memory_size   = 512

  environment {
    variables = {
      DB_HOST     = aws_db_instance.stori.address
      DB_PORT     = "5432"
      DB_USER     = var.db_username
      DB_PASSWORD = var.db_password
      DB_NAME     = var.db_name
      DB_SCHEMA   = "public"

      S3_BUCKET_NAME = aws_s3_bucket.transactions.bucket
      S3_REGION      = var.aws_region

      SES_FROM      = var.email_from
      EMAIL_DEFAULT = var.email_default

      AWS_ENDPOINT_URL      = ""
      AWS_S3_USE_PATH_STYLE = "false"
      STORI_LOGO_URL        = var.stori_logo_url
    }
  }
}

resource "aws_lambda_function" "api_handler" {
  function_name = "stori-api-handler"
  package_type  = "Image"
  image_uri     = var.ecr_api_handler_image
  role          = aws_iam_role.lambda_exec.arn
  timeout       = 15
  memory_size   = 256

  environment {
    variables = {
      DB_HOST     = aws_db_instance.stori.address
      DB_PORT     = "5432"
      DB_USER     = var.db_username
      DB_PASSWORD = var.db_password
      DB_NAME     = var.db_name
      DB_SCHEMA   = "public"

      S3_BUCKET_NAME = aws_s3_bucket.transactions.bucket
      S3_REGION      = var.aws_region

      SES_FROM      = var.email_from
      EMAIL_DEFAULT = var.email_default

      AWS_ENDPOINT_URL      = ""
      AWS_S3_USE_PATH_STYLE = "false"
      STORI_LOGO_URL        = var.stori_logo_url
    }
  }
}

resource "aws_lambda_permission" "allow_s3_invoke" {
  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.s3_processor.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.transactions.arn
}

resource "aws_s3_bucket_notification" "s3_to_lambda" {
  bucket = aws_s3_bucket.transactions.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.s3_processor.arn
    events              = ["s3:ObjectCreated:*"]
    filter_suffix       = ".csv"
  }

  depends_on = [aws_lambda_permission.allow_s3_invoke]
}

resource "aws_apigatewayv2_api" "http_api" {
  name          = "stori-http-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id                 = aws_apigatewayv2_api.http_api.id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.api_handler.invoke_arn
  integration_method     = "POST"
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "default_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_lambda_permission" "allow_apigw_invoke" {
  statement_id  = "AllowAPIGWInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_handler.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}
