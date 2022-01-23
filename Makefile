.PHONY: build test

bin_dir=build/
commit=$$(git rev-parse HEAD)
time=$$(date +%s)

# Things in my.mk:
# aws-lambda-function=my-lambda-function
# aws-lambda-region=ap-east-1
# aws-lambda-role=arn:aws:iam::xxxxx:....
-include ./my.mk

build:
	@go build \
	-ldflags "-X 'github.com/nextdotid/proof-server/common.Environment=develop' -X 'github.com/nextdotid/proof-server/common.Revision=${commit}' -X 'github.com/nextdotid/proof-server/common.BuildTime=${time}'" \
	-o ${bin_dir} ./cmd/...

test:
	@go test -v ./...

lambda-build:
	@go clean
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
	-v -x \
	-ldflags "-X 'github.com/nextdotid/proof-server/common.Environment=staging' -X 'github.com/nextdotid/proof-server/common.Revision=${commit}' -X 'github.com/nextdotid/proof-server/common.BuildTime=${time}'" \
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
