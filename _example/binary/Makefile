.PHONY: all

all: test

test:
	@env $$(xargs < comptest.env) go test ./ -p 4 -v -count=1
