# HTTP proxy on AWS Lambda

this is basically a reimplementation of [dan-v/awslambdaproxy](https://github.com/dan-v/awslambdaproxy) 
for my own understanding and trimmed down heavily to suit my own simpler use case

## What

*lambdaproxy* is a short-lived anonymous HTTP proxy which draws IP addresses from the large pool of AWS-managed Lambda compute instances,
with the goal of obfuscating network traffic

## How

*lambdaproxy* utilizes the fact that AWS Lambda provides no guarantee as to which of AWS's large fleet of compute instances is assigned to any given (cold start) invocation.
Thus, a Lambda function which simply serves an HTTP proxy will provide a different IP address each time it is invoked.
Since it is not possible to connect to running Lambda functions directly, the Lambda function also establishes a reverse SSH tunnel to the proxy host, 
after which all HTTP traffic is forwarded from the proxy host to the Lambda function's HTTP proxy.

## Usage

`lambdaproxy` must be run on a publicly acessible host (eg. AWS EC2, AWS Lightsail)

1. Install required: 
   - [`go`](https://golang.org/doc/install) (>=1.16)
   - [`serverless`](https://www.serverless.com/framework/docs/getting-started/)
   - [`aws`](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html)
2. configure AWS credentials with `aws configure`
3. `git clone https://github.com/jaswon/lambdaproxy`
4. `cd lambdaproxy`
5. `make`
6. start the proxy with `bin/start`
7. use the proxy by forwarding a local port to `127.0.0.1:6789` on the server via SSH

### Notes
- Each session lasts at most 15 minutes (the maximum execution duration of Lambda functions), but you can just rerun `bin/start`
- AWS keeps Lambda instances alive for [approximately 5-7 minutes after idling](https://mikhail.io/serverless/coldstarts/aws/intervals/), 
which means that you'll receive the same IP address if you start the proxy less than 5-7 minutes after stopping
