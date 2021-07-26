clean:
	rm -rf tplink2mqtt.*
lint:
	golangci-lint run ./internal/... ./cmd/... ./pkg/...
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tplink2mqtt.linux_amd64 ./cmd/tplink2mqtt
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o tplink2mqtt.darwin_amd64 ./cmd/tplink2mqtt
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o tplink2mqtt.windows_amd64.exe ./cmd/tplink2mqtt

docker:
	docker build -f ./cmd/tplink2mqtt/Dockerfile -t shauncampbell/tplink2mqtt:local .