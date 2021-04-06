.PHONY: build
build:
	GOOS=linux go build -o webapp .

.PHONY: deploy
deploy: build
	scp -i ~/.ec2/hatena-engineer-astj.pem ./webapp ubuntu@3.112.174.97:/home/ishocon/webapp/go/
