variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "aws_profile" {
  type        = string
  description = "AWS CLI profile to use for this deployment"
  default     = "personal"
}

variable "s3_bucket_name" {
  description = "Globally unique S3 bucket name for transaction CSVs"
  type        = string
}

variable "db_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "stori"
}

variable "db_username" {
  description = "PostgreSQL username"
  type        = string
  default     = "app"
}

variable "db_password" {
  description = "PostgreSQL password"
  type        = string
  sensitive   = true
}

variable "email_from" {
  description = "SES-verified FROM email address"
  type        = string
}

variable "email_default" {
  description = "Destination email for summaries"
  type        = string
}

variable "stori_logo_url" {
  description = "Public URL of Stori logo"
  type        = string
  default     = "https://media.licdn.com/dms/image/v2/D4E0BAQHuxJutLmsBFQ/company-logo_200_200/company-logo_200_200/0/1700583469952?e=1764201600&v=beta&t=yAwe1j0mbzSEM19MZSGWYt1RWiD9l7rPcgjSxGZSp_Q"
}

variable "ecr_s3_processor_image" {
  description = "Full ECR image URI for the S3 processor Lambda"
  type        = string
}

variable "ecr_api_handler_image" {
  description = "Full ECR image URI for the API handler Lambda"
  type        = string
}
