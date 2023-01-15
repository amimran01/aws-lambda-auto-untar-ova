# aws-lambda-auto-untar-ova

Lambda function written in Golang which untar .ova files pushed on an S3 bucket

To automate the execution of this function, add a trigger on the lambda

Env parameters :
- DESTINATION_S3_BUCKET (if unset, destination bucket = source bucket)