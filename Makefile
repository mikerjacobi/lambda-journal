
APP=journal

sb: clean buildserver
	docker stack deploy -c sandbox.yaml $(APP)

sb-rm:
	docker stack rm $(APP)

clean:
	rm -rf ./server/*/bin/*

buildserver:
	@cd server && env GOOS=linux go build -ldflags="-s -w" -o bin/handle_twilio_webhook ./journal/handle_twilio_webhook/...

