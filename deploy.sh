GOOS=linux go build -o main
zip deployment.zip main
aws lambda update-function-code --function-name sendAWSQuestions --zip-file fileb://deployment.zip
rm main deployment.zip
