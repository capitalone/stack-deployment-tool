VERSION:=0.0.1
export GO15VENDOREXPERIMENT=1

default: build


PACKAGES:=`go list ./... | grep -v old | grep -v /vendor/|grep -v plugins`

filename:=$(basename $(notdir ${CURDIR}))
bin_name:=sdt
MAIN_PACKAGE:=`go list -f "{{.Name}}|{{.ImportPath}}" ../...|grep -v "/vendor/" | grep -v "build" |grep "main|" | sed s:"main|":"":g`



.PHONY: test
test:
	go test -covermode=count  -v ${PACKAGES}

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out  -v ${PACKAGES}

.PHONY: record
record:
	RECORDING=1 go test -v `glide novendor |grep -Ev "^\.$$"|grep -v "plugins"`

.PHONY: install-deps
install-deps:
	glide install --strip-vcs --strip-vendor

.PHONY: deps
deps: 
	glide up  --strip-vcs --strip-vendor
	find . -type f -name ".gitmodules" -exec rm -f {} \;

.PHONY: vendor
vendor: deps

.PHONY: fmt
fmt:
	go fmt `glide nv`


BUILD_VERSION:=${VERSION}-$(shell git log -n 1 --pretty=format:'%H')
VER_FLAG:=--ldflags '-X github.com/capitalone/stack-deployment-tool/sdt.Version=$(BUILD_VERSION)'

.PHONY: quick
quick: build_darwin

.PHONY: build_linux
build_linux:
	export GOOS=$(subst build_,,$@) && export GOARCH="amd64" && mkdir -p build/$${GOOS}_$${GOARCH} && cd build &&\
		go build ${VER_FLAG} -o ${bin_name}_$${GOOS}_$${GOARCH} ${MAIN_PACKAGE} && cp ${bin_name}_$${GOOS}_$${GOARCH} $${GOOS}_$${GOARCH}/${bin_name}

.PHONY: build_darwin
build_darwin:
	export GOOS=$(subst build_,,$@) && export GOARCH="amd64" && mkdir -p build/$${GOOS}_$${GOARCH} && cd build &&\
		go build ${VER_FLAG} -o ${bin_name}_$${GOOS}_$${GOARCH} ${MAIN_PACKAGE} && cp ${bin_name}_$${GOOS}_$${GOARCH} $${GOOS}_$${GOARCH}/${bin_name}


.PHONY: build
build: clean test build_darwin build_linux

.PHONY: release
release: build

.PHONY: clean
clean:
	rm -rf build




