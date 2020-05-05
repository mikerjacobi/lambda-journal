
APP=journal

sb: clean buildserver
	@docker stack deploy -c sandbox.yaml $(APP)
	@cd server && sls dynamodb start

sb-rm:
	docker stack rm $(APP)

clean:
	rm -rf ./server/bin/*

buildserver:
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o bin/handle_twilio_webhook ./journal/handle_twilio_webhook/...
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o bin/insert_entry ./journal/insert_entry/...
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o bin/get_entry ./journal/get_entry/...

