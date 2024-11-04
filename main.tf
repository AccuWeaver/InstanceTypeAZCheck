provider "aws" {
  profile = "accuweaver-prod-sso"
  region  = "us-east-1"  # Replace with your desired region
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

variable "lambda_function_name" {
  description = "The name of the Lambda function"
  type        = string
  default     = "InstanceTypAZCheck"
}

variable "lambda_role_name" {
  description = "The name of the IAM role for the Lambda function"
  type        = string
  default     = "InstanceTypAZCheck"
}

resource "aws_iam_role" "lambda_role" {
  name = var.lambda_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "lambda_ec2_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2FullAccess"
}

resource "aws_iam_role_policy_attachment" "lambda_basic_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_cfn_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/AWSCloudFormationReadOnlyAccess"
}

resource "null_resource" "build_lambda" {
  provisioner "local-exec" {
    command = <<EOT
    {
      ERROR_FILE=$(mktemp)
      echo "Temporary file created at: $ERROR_FILE"
      OUTPUT=$(GOOS=linux GOARCH=amd64 go build -o InstanceTypAZCheck InstanceTypAZCheck.go 2> $ERROR_FILE)
      RET_CODE=$?
      ERROR=$(cat $ERROR_FILE)
      rm $ERROR_FILE

      if [ $RET_CODE -ne 0 ]; then
        echo '{"stdout": "'$OUTPUT'", "stderr": "'$ERROR'", "return_code": "'$RET_CODE'"}' > build_output.json
        exit 1
      fi

      OUTPUT=$(chmod +x InstanceTypAZCheck bootstrap 2>> $ERROR_FILE)
      RET_CODE=$?
      ERROR=$(cat $ERROR_FILE)
      rm $ERROR_FILE

      if [ $RET_CODE -ne 0 ]; then
        echo '{"stdout": "'$OUTPUT'", "stderr": "'$ERROR'", "return_code": "'$RET_CODE'"}' > build_output.json
        exit 1
      fi

      OUTPUT=$(zip InstanceTypAZCheck.zip InstanceTypAZCheck bootstrap 2>> $ERROR_FILE)
      RET_CODE=$?
      ERROR=$(cat $ERROR_FILE)
      rm $ERROR_FILE

      if [ $RET_CODE -ne 0 ]; then
        echo '{"stdout": "'$OUTPUT'", "stderr": "'$ERROR'", "return_code": "'$RET_CODE'"}' > build_output.json
        exit 1
      fi

      HASH=$(openssl dgst -sha256 -binary InstanceTypAZCheck.zip | openssl enc -base64)
      echo '{"stdout": "'$OUTPUT'", "stderr": "'$ERROR'", "return_code": "'$RET_CODE'", "hash": "'$HASH'", "return_code": "0"}' > build_output.json
    }
  EOT
  }

  triggers = {
    always_run = timestamp()
  }
}

data "external" "build_output" {
  program = ["sh", "-c", "cat build_output.json"]

  depends_on = [null_resource.build_lambda]
}

resource "aws_lambda_function" "InstanceTypAZCheck" {
  function_name = var.lambda_function_name
  role          = aws_iam_role.lambda_role.arn
  handler       = "InstanceTypAZCheck"
  runtime = "provided.al2023"  # Use custom runtime with Amazon Linux 2023
  filename      = "InstanceTypAZCheck.zip"

  source_code_hash = data.external.build_output.result[
  "hash"
  ]

  depends_on = [
    null_resource.build_lambda
  ]
}

output "build_message" {
  value = data.external.build_output.result
}

resource "aws_lambda_function_event_invoke_config" "InstanceTypAZCheck_event_invoke_config" {
  function_name = aws_lambda_function.InstanceTypAZCheck.function_name

  maximum_retry_attempts       = 2
  maximum_event_age_in_seconds = 60
}