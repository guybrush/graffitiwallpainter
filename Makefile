GITCOMMIT=`git describe --always`
VERSION=$$(git describe 2>/dev/null || echo "0.0.0-${GITCOMMIT}")
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
GITDATESHORT=$$(TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD | sed 's/[-T:]//g' | sed 's/\(+.*\)$$//g')
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
BUILDDATESHORT=`date -u +"%Y%m%d%H%M%S"`
PACKAGE=github.com/guybrush/graffitiwallpainter
LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILDDATE} -X main.GitCommit=${GITCOMMIT} -X main.GitDate=${GITDATE}"
DOCKERIMAGE="guybrush/graffitiwallpainter"
BINARY=bin/graffitipainter

all: test build

test:
	go test -v ./...

clean:
	rm -rf bin

build:
	go build --ldflags=${LDFLAGS} -o ${BINARY}

dockerimage:
	docker build -t ${DOCKERIMAGE} -t ${DOCKERIMAGE}:${GITDATESHORT}-${GITCOMMIT} .

dockerimage-push:
	docker push ${DOCKERIMAGE}
	docker push ${DOCKERIMAGE}:${GITDATESHORT}-${GITCOMMIT}

