
build/function.zip: ./cmd/function
	cd cmd && GOOS=linux CGO_ENABLED=0 go build -o ../build/main ./function
	cd build && zip function.zip main
	rm build/main

deploy: build/function.zip
	aws lambda update-function-code \
		--function-name arn:aws:lambda:us-east-1:905986754592:function:proxy \
		--zip-file fileb://build/function.zip
