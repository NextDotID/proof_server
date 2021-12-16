.PHONY: build test

bin_dir=build/

# Things in my.mk:
# aws-lambda-function=my-lambda-function
# aws-lambda-region=ap-east-1
# aws-lambda-role=arn:aws:iam::xxxxx:....
-include ./my.mk

build:
	@go build -o ${bin_dir} ./cmd/...

test:
	@go test -v ./...

lambda-build:
	@go clean
	@GOOS=linux CGO_ENABLED=0 go build \
	-x \
	-o ./build/lambda \
	./cmd/lambda

lambda-pack: lambda-build
	@cd ./build && zip lambda.zip lambda

lambda-update: lambda-pack
	@aws lambda update-function-code --function-name ${aws-lambda-function} --zip-file 'fileb://./build/lambda.zip'

lambda-create: lambda-pack
	aws lambda create-function \
		--region ${aws-lambda-region} \
		--function-name ${aws-lambda-function} \
		--handler lambda \
		--role ${aws-lambda-role} \
		--runtime go1.x \
		--zip-file "fileb://./build/lambda.zip"

lambda-publish:
	aws lambda update-function-code --function-name ${aws-lambda-function} --publish
