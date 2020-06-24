VERSION ?= $(shell git branch --show-current)
GIT_COMMIT ?= $(shell git rev-list -1 HEAD)
BUILD_DATE ?= $(shell date)
LDFLAGS=-ldflags "-X 'github.com/wndhydrnt/autoboard/cmd.BuildDate=${BUILD_DATE}' -X 'github.com/wndhydrnt/autoboard/cmd.BuildHash=${GIT_COMMIT}' -X 'github.com/wndhydrnt/autoboard/cmd.Version=${VERSION}'"

build:
	go build ${LDFLAGS}

build_darwin:
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o autoboard-${VERSION}.darwin-amd64

build_linux:
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o autoboard-${VERSION}.linux-amd64

build_windows:
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o autoboard-${VERSION}.windows-amd64

.PHONY: test
test:
	go test -count=1 ./...

test_bootstrap: test_bootstrap_services test_bootstrap_grafana

test_bootstrap_services:
	cd test && docker-compose up -d grafana prometheus
	sleep 10

test_bootstrap_grafana:
	curl -X"POST" -u"admin:admin" -H'Content-Type: application/json' http://localhost:12959/api/datasources --data '{"name":"test_datasource","type":"prometheus","access":"proxy","url":"http://prometheus:9090","basicAuth":false, "isDefault": true}'
	curl -X"POST" -u"admin:admin" -H'Content-Type: application/json' http://localhost:12959/api/folders --data '{"title":"Test Folder"}'

test_bootstrap_stop:
	cd test && docker-compose down

generate_templates:
	scripts/generate-templates.sh
