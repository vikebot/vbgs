# Go parameters
GOCMD=go
GOBUILD=${GOCMD} build
GOCLEAN=${GOCMD} clean
GOTEST=${GOCMD} test
GOGET=${GOCMD} get
GOMOD=${GOCMD} mod

# Set binary names
REPOSITORY=vikebot/vbgs
BINARY=vbgs

# Extract git information of current build
COMMIT_HASH := ${shell git rev-parse --verify HEAD}
COMMIT_HASH_SHORT := ${shell echo ${COMMIT_HASH} | cut -c1-7}
COMMIT_TAG := ${shell git show-ref --tags --dereference | grep ${COMMIT_HASH} | awk 'BEGIN{FS=$$0}{ for(i=1;i<=NF;i++){ if($$i=="/"){ p=i } }}END{ print substr($$0, p+1, 6) }'}

# Define version information of build
VERSION=${COMMIT_HASH}
VERSION_SHORT=${COMMIT_HASH_SHORT}
ifeq (${COMMIT_TAG},)
else
	VERSION=${COMMIT_TAG}@${COMMIT_HASH_SHORT}
	VERSION_SHORT=${COMMIT_TAG}
endif

# Working directory
WORKING_DIRECTORY := ${shell pwd}


# Commands section
all: clean mod test test-sec build
clean:
	${GOCLEAN}
	rm -f ${BINARY}
	rm -rf release/
mod-download:
	${GOMOD} download
mod: mod-verify mod-vendor
mod-verify:
	${GOMOD} verify
mod-vendor:
	${GOMOD} vendor
test: mod test-sec
	${GOTEST} -cover -covermode=atomic -race ./...
build:
	${GOBUILD} -o ${BINARY}
test-sec:
	# hack for getting gosec working with go modules
	# from https://github.com/securego/gosec/issues/234#issuecomment-427463106
	docker run -it --rm -v ${WORKING_DIRECTORY}:/go/src/github.com/vikebot/vbgs -e "GOPATH=/go" securego/gosec -quiet -severity medium /go/src/github.com/vikebot/vbgs/...


# Release compilation:
RELEASE_DIR=release/
BINARY_RELEASE=${RELEASE_DIR}${BINARY}
GOBUILDPARAMS=-mod vendor
LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION}"

# shortcut for building everything
release-binaries: release-linux-amd64 release-linux-386 release-windows-amd64 release-windows-386 release-darwin-amd64 release-darwin-386
release: release-binaries release-docker
release-all: release


#
# NORMAL BINARY BUILDS
#
define release-binary
	GOOS=${1} GOARCH=${2} ${GOBUILD} ${GOBUILDPARAMS} ${LDFLAGS} -o ${BINARY_RELEASE}${3}
	cd ${RELEASE_DIR} && \
		tar -zcvf ${BINARY}-${VERSION_SHORT}-${1}-${2}.tar.gz ${BINARY}${3}
	cd ${RELEASE_DIR} && \
		sha256sum ${BINARY}-${VERSION_SHORT}-${1}-${2}.tar.gz >> ${BINARY}-${VERSION_SHORT}-checksums.txt
	rm -rf ${BINARY_RELEASE}${3}
endef

# LINUX
release-linux-amd64: clean mod test
	${call release-binary,linux,amd64,}
release-linux-386: clean mod test
	${call release-binary,linux,386,}
# WINDOWS
release-windows-amd64: clean mod test
	${call release-binary,windows,amd64,.exe}
release-windows-386: clean mod test
	${call release-binary,windows,386,.exe}
# DARWIN
release-darwin-amd64: clean mod test
	${call release-binary,darwin,amd64,}
release-darwin-386: clean mod test
	${call release-binary,darwin,386,}

#
# DOCKER BUILDS
#
ifeq (${COMMIT_TAG},)
release-docker: clean mod test
	docker build --compress --build-arg VERSION=${VERSION} -t ${REPOSITORY} .
else
release-docker: clean mod test
	docker build --compress --build-arg VERSION=${VERSION} -t ${REPOSITORY}:${COMMIT_TAG} .
endif
