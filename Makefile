GIT_DATE := $(shell git log -1 --date=short --pretty='%cd' | tr -d -)
GIT_HASH := $(shell git rev-parse HEAD)

BUILD_FLAGS := -ldflags "-X main.git_date=$(GIT_DATE) -X main.git_hash=$(GIT_HASH)"

PLATFORMS := linux/amd64 linux/386 linux/arm darwin/amd64 windows/amd64 windows/386 openbsd/amd64
SOURCES := $(shell find . -maxdepth 1 -type f -name "*.go")
ALL_SOURCES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
ext = $(shell if [ "$(os)" = "windows" ]; then echo ".exe"; fi)

all: certgraph

release: $(PLATFORMS)

certgraph: $(SOURCES) $(ALL_SOURCES)
	go build $(BUILD_FLAGS) -o $@ $(SOURCES)

$(PLATFORMS): $(SOURCES)
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) go build $(BUILD_FLAGS) -o 'build/$(os)/$(arch)/certgraph$(ext)' $(SOURCES)
	cd build/$(os)/$(arch)/; zip -r ../../certgraph-$(os)-$(arch)-$(GIT_DATE).zip .; cd ../../../

fmt:
	gofmt -s -w -l .

clean:
	rm -r certgraph build/

serv:
	(cd docs; python -m SimpleHTTPServer)

.PHONY: all fmt clean release serv $(PLATFORMS)
