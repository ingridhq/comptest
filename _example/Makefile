.PHONY: all

all: setup test clean

setup:
	@docker-compose -f comptest/docker-compose.yaml --env-file comptest/comptest.env up --force-recreate -d

test:
	@env $$(xargs < comptest/comptest.env) go test ./comptest -p 4 -v -count=1

clean:
	@docker-compose -f comptest/docker-compose.yaml --env-file comptest/comptest.env down &
