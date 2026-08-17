build: deps
	CGO_ENABLED=0 go build \
		-ldflags " \
                -s -w \
		-X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD) \
		-X github.com/prometheus/common/version.Revision=$(shell git rev-parse HEAD) \
		-X github.com/prometheus/common/version.Version=$(shell cat VERSION) \
		"
deps:
	go mod download
