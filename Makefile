
function.zip: ./function
	GOOS=linux CGO_ENABLED=0 go build -o main ./function
	zip function.zip main
	rm main

deploy: function.zip
	aws lambda update-function-code \
		--function-name arn:aws:lambda:us-east-1:905986754592:function:proxy \
		--zip-file fileb://function.zip
