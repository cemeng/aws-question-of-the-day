# aws-question-of-the-day

Serverless proof of concept using Lambda, DynamoDB, SES using Golang

Deployment (manual):
```
GOOS=linux go build -o main
zip deployment.zip main
```

Upload deployment.zip using AWS Lambda console
