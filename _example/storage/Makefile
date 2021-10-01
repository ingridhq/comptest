.PHONY: all

all: setup test clean

setup:
	@docker-compose -f docker-compose.yaml --env-file comptest.env up --force-recreate -d

test:
	@env $$(xargs < comptest.env) go test ./ -p 4 -v -count=1

clean:
	@docker-compose -f docker-compose.yaml --env-file comptest.env down > /dev/null&
