output "s3_bucket_name" {
  description = "S3 bucket where CSV files must be uploaded"
  value       = aws_s3_bucket.transactions.bucket
}

output "rds_endpoint" {
  description = "PostgreSQL endpoint (use as DB_HOST)"
  value       = aws_db_instance.stori.address
}

output "lambda_s3_processor_arn" {
  value = aws_lambda_function.s3_processor.arn
}

output "lambda_api_handler_arn" {
  value = aws_lambda_function.api_handler.arn
}

output "api_gateway_url" {
  description = "Invoke URL for the HTTP API"
  value       = aws_apigatewayv2_api.http_api.api_endpoint
}
