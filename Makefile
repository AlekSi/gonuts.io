GOPATH:=$(shell pwd)/gopath
SDKROOT:=/usr/local/Cellar/go-app-engine-64/1.8.3/share/go-app-engine-64

GOFILES:=$(shell cd app && find . -name *.go)

all: fvb

prepare:
	rm -fr $(GOPATH)
	mkdir -p $(GOPATH)/src
	go get -u github.com/bmizerany/pat
	rm -fr $(GOPATH)/src/github.com/bmizerany/pat/.git
	rm -fr $(GOPATH)/src/github.com/bmizerany/pat/example
	go get -u github.com/mjibson/appstats
	rm -fr $(GOPATH)/src/github.com/mjibson/appstats/.git
	rm -fr $(GOPATH)/src/code.google.com
	nut get -v aleksi/nut

# format, vet, build
fvb:
	gofmt -e -s -w .
	go tool vet .
	cd app && $(SDKROOT)/goroot/bin/go-app-builder -goroot=$(SDKROOT)/goroot -dynamic -unsafe $(GOFILES)

run: fvb
	$(SDKROOT)/dev_appserver.py --skip_sdk_update_check=yes app

run_clean: fvb
	$(SDKROOT)/dev_appserver.py --skip_sdk_update_check=yes --clear_datastore=yes app

check_clean:
	git diff-index --exit-code HEAD
	u="$$(git ls-files --others --exclude-standard)" && echo $$u && test -z "$$u"

deploy: fvb check_clean
	$(SDKROOT)/appcfg.py --oauth2 update app
