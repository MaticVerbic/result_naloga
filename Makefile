dc_run_api := docker-compose run --name result --rm api

api:
	$(dc_run_api) go run cmd/main.go

tidy:
	$(dc_run_api) go mod tidy
