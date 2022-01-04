tidy:
	go mod tidy
gen:
	sqlc generate
run: tidy gen
	go run cmd/underpass/main.go
run_logs: tidy gen
	go run cmd/underpass/main.go --save_logs
build-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o underpass cmd/underpass/main.go

