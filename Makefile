VERSION ?= $(shell git branch --show-current)
GIT_COMMIT ?= $(shell git rev-list -1 HEAD)
BUILD_DATE ?= $(shell date)

build:
	go build -ldflags "-X 'github.com/wndhydrnt/autoboard/cmd.BuildDate=${BUILD_DATE}' -X 'github.com/wndhydrnt/autoboard/cmd.BuildHash=${GIT_COMMIT}' -X 'github.com/wndhydrnt/autoboard/cmd.Version=${VERSION}'"

.PHONY: test
test:
	go test -count=1 ./...

test_bootstrap: test_bootstrap_services test_bootstrap_grafana

test_bootstrap_services:
	cd test && docker-compose up -d
	sleep 10

test_bootstrap_grafana:
	curl -X"POST" -u"admin:admin" -H'Content-Type: application/json' http://localhost:12959/api/datasources --data '{"name":"test_datasource","type":"prometheus","access":"proxy","url":"http://prometheus:9090","basicAuth":false}'
	curl -X"POST" -u"admin:admin" -H'Content-Type: application/json' http://localhost:12959/api/folders --data '{"title":"Test Folder"}'

test_bootstrap_stop:
	cd test && docker-compose down

generate: generate_config generate_templates

generate_config:
	scripts/generate-config.sh

generate_templates:
	scripts/generate-templates.sh
