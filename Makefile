usage: FORCE
	exit 1

FORCE:

include config.env
export $(shell sed 's/=.*//' config.env)

start: FORCE
	@echo " >> building..."
	@mkdir -p log
	@go build
	@./grpcox