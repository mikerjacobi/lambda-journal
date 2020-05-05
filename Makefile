
APP=journal

sb: clean buildserver
	@docker stack deploy -c sandbox.yaml $(APP)
	@cd server && sls dynamodb start

sb-rm:
	docker stack rm $(APP)

clean:
	rm -rf ./server/journal/bin/*

buildserver:
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o journal/bin/handle_twilio_webhook ./journal/handle_twilio_webhook/...
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o journal/bin/insert_journal ./journal/insert_journal/...
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o journal/bin/get_journal ./journal/get_journal/...

deploydynamo: 
	yamllint server/serverless.yaml
	@cd server && sls deploy --aws-s3-accelerate --force --verbose --stage production

deploylambdas: clean buildserver
	@yamllint server/journal/serverless.yaml
	@cd server/journal/bin && zip -r handle_twilio_webhook handle_twilio_webhook
	@cd server/journal/bin && zip -r insert_journal insert_journal
	@cd server/journal/bin && zip -r get_journal get_journal
	@cd server/journal && serverless package --stage production
	@cd server/journal && sls deploy --aws-s3-accelerate --force --verbose --stage production
