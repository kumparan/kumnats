SHELL:=/bin/bash

changelog_args=-o CHANGELOG.md -p '^v'

lint:
	golangci-lint run --concurrency 4 --print-issued-lines=false --exclude-use-default=false --enable=golint --enable=goimports  --enable=unconvert --enable=unparam

changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)