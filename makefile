tidy:
	go mod tidy
gen:
	sqlc generate
run: tidy gen
	go run cmd/underpass/main.go
run-logs: tidy gen
	go run cmd/underpass/main.go --save_logs
build-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o bin/underpass_linux_amd64 cmd/underpass/main.go

