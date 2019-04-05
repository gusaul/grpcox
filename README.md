# gRPCox
[![Go Report Card](https://goreportcard.com/badge/github.com/gusaul/grpcox)](https://goreportcard.com/report/github.com/gusaul/grpcox)

turn [gRPCurl](https://github.com/fullstorydev/grpcurl) into web based UI, extremely easy to use

## Features
- Recognize and provide list of services and methods inside it as an options.
- Automatically recognize schema input and compose it into JSON based. (ensure your gRPC server supports [server reflection](https://github.com/grpc/grpc/blob/master/src/proto/grpc/reflection/v1alpha/reflection.proto)). Examples for how to set up server reflection can be found [here](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md#known-implementations).
- Save established connection, and reuse it for next invoke/request (also can close/restart connection)

## Installation
### Docker Compose
from terminal, move to grpcox directory, then run command
```shell
docker-compose up
```

## Demo
![gRPCox Demo](https://raw.githubusercontent.com/gusaul/grpcox/master/index/img/demogrpcox.gif)
