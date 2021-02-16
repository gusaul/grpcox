# gRPCox
[![Go Report Card](https://goreportcard.com/badge/github.com/gusaul/grpcox)](https://goreportcard.com/report/github.com/gusaul/grpcox)

turn [gRPCurl](https://github.com/fullstorydev/grpcurl) into web based UI, extremely easy to use

## Features
- Recognize and provide list of services and methods inside it as an options.
- Automatically recognize schema input and compose it into JSON based. (ensure your gRPC server supports [server reflection](https://github.com/grpc/grpc/blob/master/src/proto/grpc/reflection/v1alpha/reflection.proto)). Examples for how to set up server reflection can be found [here](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md#known-implementations).
- Save established connection, and reuse it for next invoke/request (also can close/restart connection)

## Installation
### Docker
```shell
docker pull gusaul/grpcox:latest
```
then run
```shell
docker run -p 6969:6969 -v {ABSOLUTE_PATH_TO_LOG}/log:/log -d gusaul/grpcox
```

### Docker Compose
from terminal, move to grpcox directory, then run command
```shell
docker-compose up
```
if you're using docker and want to connect gRPC on your local machine, then use
<br/>`host.docker.internal:<your gRPC port>` instead of `localhost`

### Golang
if you have golang installed on your local machine, just run command
```shell
make start
```
from grpcox directory

configure app preferences by editing `config.env` file

| var             | usage                                       | type   | unit   |
|-----------------|---------------------------------------------|--------|--------|
| MAX_LIFE_CONN   | maximum idle time connection before closed  | number | minute |
| TICK_CLOSE_CONN | ticker interval to sweep expired connection | number | second |
| BIND_ADDR       | ip:port to bind service                     | string |  |

set value `0 (zero)` to disable auto close idle connection.

## Demo
![gRPCox Demo](https://raw.githubusercontent.com/gusaul/grpcox/master/index/img/demogrpcox.gif)
