
build/function.zip: ./cmd/function
	GOOS=linux CGO_ENABLED=0 go build -o function ./cmd/function
	zip build/function.zip function
	rm function

deploy: function.zip
	aws lambda update-function-code \
		--function-name arn:aws:lambda:us-east-1:905986754592:function:proxy \
		--zip-file fileb://build/function.zip
