# result_naloga

For swagger docs use [http://localhost/docs/swagger/](http://localhost/docs/swagger/). 
To run the task use `http://localhost/result?workers=x`, where x represents the amount of workers wanted. 

Task demands, between 1 and 4 workers, but for better generalization, this requirement has been omitted and API allows for more workers, since the amount of urls to be hit can also be changed in .env file. 

Docker and docker-compose support has been added, but due to issues on personal machine (hypver-v issues) the last optional task has not been fully completed. 

Run instructions
* `cp .env.example .env`
* `go run cmd/main.go`

Tasks: 
Main part: 
Simply run the service and use `result` with url query `workers` to fetch results. 

Tests: 
To run tests simply use `go test ./...` or `go test -v ./...` for verbose testing. 

Swagger: 
Package used to interface swagger is [swaggo/swag](https://github.com/swaggo/swag). To access the documentations use [http://localhost/docs/swagger/index.html](http://localhost/docs/swagger/index.html). 

Docker: 
Added Dockerfile and docker-compose file, but could not test or deploy due to issues with hyper-v on my machine. 
