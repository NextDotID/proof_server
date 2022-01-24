.PHONY: build test

bin_dir=build/
commit=$$(git rev-parse HEAD)
time=$$(date +%s)

# Things in my.mk:
# aws-lambda-function-staging=my-lambda-function-staging
# aws-lambda-function-production=my-lambda-function-production
# aws-lambda-region=ap-east-1
# aws-lambda-role=arn:aws:iam::xxxxx:....
-include ./my.mk

build:
	@go build \
	-ldflags "-X 'github.com/nextdotid/proof-server/common.Environment=develop' -X 'github.com/nextdotid/proof-server/common.Revision=${commit}' -X 'github.com/nextdotid/proof-server/common.BuildTime=${time}'" \
	-o ${bin_dir} ./cmd/...

test:
	@go test -v ./...

lambda-build-production:
	@go clean
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
	-v -x \
	-ldflags "-X 'github.com/nextdotid/proof-server/common.Environment=production' -X 'github.com/nextdotid/proof-server/common.Revision=${commit}' -X 'github.com/nextdotid/proof-server/common.BuildTime=${time}'" \
	-o ./build/lambda \
	./cmd/lambda

lambda-build-staging:
	@go clean
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
	-v -x \
	-ldflags "-X 'github.com/nextdotid/proof-server/common.Environment=staging' -X 'github.com/nextdotid/proof-server/common.Revision=${commit}' -X 'github.com/nextdotid/proof-server/common.BuildTime=${time}'" \
	-o ./build/lambda \
	./cmd/lambda

lambda-pack-staging: lambda-build-staging
	@cd ./build && zip lambda.zip lambda

lambda-pack-production: lambda-build-production
	@cd ./build && zip lambda.zip lambda

lambda-update-staging: lambda-pack-staging
	@aws lambda update-function-code --function-name ${aws-lambda-function-staging} --zip-file 'fileb://./build/lambda.zip'

lambda-update-production: lambda-pack-production
	@aws lambda update-function-code --function-name ${aws-lambda-function-production} --zip-file 'fileb://./build/lambda.zip'

lambda-create-staging: lambda-pack-staging
	aws lambda create-function \
		--region ${aws-lambda-region} \
		--function-name ${aws-lambda-function-staging} \
		--handler lambda \
		--role ${aws-lambda-role} \
		--runtime go1.x \
		--zip-file "fileb://./build/lambda.zip"

lambda-create-production: lambda-pack-production
	aws lambda create-function \
		--region ${aws-lambda-region} \
		--function-name ${aws-lambda-function-production} \
		--handler lambda \
		--role ${aws-lambda-role} \
		--runtime go1.x \
		--zip-file "fileb://./build/lambda.zip"
