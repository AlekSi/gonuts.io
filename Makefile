GOPATH:=$(shell pwd)/gopath
SDKROOT:=/usr/local/Cellar/go-app-engine-64/1.7.5/share/go-app-engine-64

GOFILES:=$(shell find . -name *.go)

all: fvb

prepare:
	rm -fr $(GOPATH)
	mkdir -p $(GOPATH)/src
	go get -u github.com/bmizerany/pat
	rm -fr $(GOPATH)/src/github.com/bmizerany/pat/.git
	rm -fr $(GOPATH)/src/github.com/bmizerany/pat/example
	nut get -v aleksi/nut

# format, vet, build
fvb:
	gofmt -e -s -w .
	go tool vet .
	$(SDKROOT)/goroot/bin/go-app-builder -goroot=$(SDKROOT)/goroot -dynamic -unsafe $(GOFILES)

run: fvb
	$(SDKROOT)/dev_appserver.py --skip_sdk_update_check --use_sqlite .

run_clean: fvb
	$(SDKROOT)/dev_appserver.py --skip_sdk_update_check --use_sqlite --clear_datastore .

check_clean:
	git diff-index --exit-code HEAD
	u="$$(git ls-files --others --exclude-standard)" && echo $$u && test -z "$$u"

deploy: fvb check_clean
	$(SDKROOT)/appcfg.py --oauth2 update .
