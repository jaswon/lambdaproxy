
all: bin/start deploy

bin/start: ./host
	GOOS=linux CGO_ENABLED=0 go build -o bin/start ./host

bin/proxy: ./function
	GOOS=linux CGO_ENABLED=0 go build -o bin/proxy ./function

deploy: bin/proxy
	sls deploy
