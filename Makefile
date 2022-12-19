.PHONY: build test

bin_dir=build/
commit=$$(git rev-parse HEAD)
time=$$(date +%s)
aws_registry_uri=${aws-account-id}.dkr.ecr.${aws-lambda-region}.amazonaws.com
aws_docker_image=${aws_registry_uri}/${docker-image-name}:${commit}

# Things in my.mk:
# aws-lambda-function-staging=my-lambda-function-staging
# aws-lambda-function-production=my-lambda-function-production
# aws-lambda-headless-function-staging=my-lambda-headless-function-staging
# aws-lambda-region=ap-east-1
# aws-lambda-role=arn:aws:iam::xxxxx:....
# aws-account-id=xxxxxxxxxx
# docker-image-name=lambda_headless

-include ./my.mk

build:
	@go build \
	-ldflags "-X 'github.com/nextdotid/proof_server/common.Environment=develop' -X 'github.com/nextdotid/proof_server/common.Revision=${commit}' -X 'github.com/nextdotid/proof_server/common.BuildTime=${time}'" \
	-o ${bin_dir} ./cmd/...

test:
	@go test -v ./...

lambda-build-production:
	@go clean
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
	-v -x \
	-ldflags "-X 'github.com/nextdotid/proof_server/common.Environment=production' -X 'github.com/nextdotid/proof_server/common.Revision=${commit}' -X 'github.com/nextdotid/proof_server/common.BuildTime=${time}'" \
	-o ./build/lambda \
	./cmd/lambda

lambda-build-staging:
	@go clean
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
	-v -x \
	-ldflags "-X 'github.com/nextdotid/proof_server/common.Environment=staging' -X 'github.com/nextdotid/proof_server/common.Revision=${commit}' -X 'github.com/nextdotid/proof_server/common.BuildTime=${time}'" \
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

lambda-build-worker-staging:
	@go clean
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
	-v -x \
	-ldflags "-X 'github.com/nextdotid/proof_server/common.Environment=staging' -X 'github.com/nextdotid/proof_server/common.Revision=${commit}' -X 'github.com/nextdotid/proof_server/common.BuildTime=${time}'" \
	-o ./build/lambda \
	./cmd/lambda_worker

lambda-pack-worker-staging: lambda-build-worker-staging
	@cd ./build && zip lambda.zip lambda

lambda-update-worker-staging: lambda-pack-worker-staging
	@aws lambda update-function-code --function-name ${aws-lambda-function-worker-staging} --zip-file 'fileb://./build/lambda.zip'

lamda-create-registry-headless:
	@aws ecr get-login-password \
		--region ${aws-lambda-region} | docker login \
		--username AWS \
		--password-stdin ${aws_registry_uri}
	@aws ecr describe-repositories \
		--repository-names ${docker-image-name} || \
		aws ecr create-repository \
		--repository-name ${docker-image-name} \
		--region ${aws-lambda-region} \
		--image-scanning-configuration scanOnPush=true \
		--image-tag-mutability MUTABLE

lambda-build-headless-staging:
	@docker build -f ./cmd/lambda_headless/Dockerfile -t ${docker-image-name}:${commit} .
	@docker tag ${docker-image-name}:${commit} ${aws_docker_image}

lambda-pack-headless-staging: lamda-create-registry-headless lambda-build-headless-staging
	@docker push ${aws_docker_image}

lambda-create-headless-staging: lambda-pack-headless-staging
	aws lambda create-function \
		--package-type Image \
		--region ${aws-lambda-region} \
		--function-name ${aws-lambda-headless-function-staging} \
		--code ImageUri=${aws_docker_image} \
		--memory-size 1200 \
		--timeout 30 \
		--architectures x86_64 \
		--role ${aws-lambda-role}

lambda-update-headless-staging: lambda-pack-headless-staging
	@aws lambda update-function-code --function-name ${aws-lambda-headless-function-staging} --image-uri ${aws_docker_image}
