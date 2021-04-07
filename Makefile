webapp: *.go
	GOOS=linux go build -o webapp .

.PHONY: deploy-webapp
deploy-webapp: webapp
	scp -i ~/.ec2/hatena-engineer-astj.pem ./webapp ubuntu@3.112.174.97:/home/ishocon/webapp/go/
	scp -i ~/.ec2/hatena-engineer-astj.pem ./templates/* ubuntu@3.112.174.97:/home/ishocon/webapp/go/templates/

.PHONY: deploy-nginx
deploy-nginx: nginx.conf
	scp -i ~/.ec2/hatena-engineer-astj.pem ./nginx.conf ubuntu@3.112.174.97:/etc/nginx/nginx.conf

.PHONY: deploy
deploy: deploy-nginx deploy-webapp
