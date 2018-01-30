TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
VERSION=$(shell ./scripts/git-version.sh)

default: build

build: fmtcheck
	go install

clean:
	@rm -rf bin

clean-release:
	@rm -rf _output

release: \
	clean \
	clean-release \
	_output/terraform-provider-hcloud_linux_amd64.zip \
	_output/terraform-provider-hcloud_darwin_amd64.zip \
	_output/terraform-provider-hcloud_freebsd_amd64.zip \
	_output/terraform-provider-hcloud_freebsd_386.zip \
	_output/terraform-provider-hcloud_freebsd_arm.zip \
	_output/terraform-provider-hcloud_linux_amd64.zip \
	_output/terraform-provider-hcloud_linux_386.zip \
	_output/terraform-provider-hcloud_linux_arm.zip \
	_output/terraform-provider-hcloud_openbsd_amd64.zip \
	_output/terraform-provider-hcloud_openbsd_386.zip \
	_output/terraform-provider-hcloud_solaris_amd64.zip \
	_output/terraform-provider-hcloud_windows_amd64.zip \
	_output/terraform-provider-hcloud_windows_386.zip

bin/darwin_amd64/terraform-provider-hcloud:  GOARGS = GOOS=darwin GOARCH=amd64
bin/freebsd_amd64/terraform-provider-hcloud:  GOARGS = GOOS=freebsd GOARCH=amd64
bin/freebsd_386/terraform-provider-hcloud:  GOARGS = GOOS=freebsd GOARCH=386
bin/freebsd_arm/terraform-provider-hcloud:  GOARGS = GOOS=freebsd GOARCH=arm
bin/linux_amd64/terraform-provider-hcloud:  GOARGS = GOOS=linux GOARCH=amd64
bin/linux_386/terraform-provider-hcloud:  GOARGS = GOOS=linux GOARCH=386
bin/linux_arm/terraform-provider-hcloud:  GOARGS = GOOS=linux GOARCH=arm
bin/openbsd_amd64/terraform-provider-hcloud:  GOARGS = GOOS=openbsd GOARCH=amd64
bin/openbsd_386/terraform-provider-hcloud:  GOARGS = GOOS=openbsd GOARCH=386
bin/solaris_amd64/terraform-provider-hcloud:  GOARGS = GOOS=solaris GOARCH=amd64
bin/windows_amd64/terraform-provider-hcloud:  GOARGS = GOOS=windows GOARCH=amd64
bin/windows_386/terraform-provider-hcloud:  GOARGS = GOOS=windows GOARCH=386

bin/%/terraform-provider-hcloud: clean
	$(GOARGS) go build -o $@ -a .

_output/terraform-provider-hcloud_%.zip: NAME=terraform-provider-hcloud_$(VERSION)_$*
_output/terraform-provider-hcloud_%.zip: DEST=_output/$(NAME)
_output/terraform-provider-hcloud_%.zip: bin/%/terraform-provider-hcloud
	mkdir -p $(DEST)
	cp bin/$*/terraform-provider-hcloud README.md LICENSE $(DEST)
	cd $(DEST) && zip -r ../$(NAME).zip .

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./aws"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

.PHONY: build test testacc vet fmt fmtcheck errcheck vendor-status test-compile release clean clean-release
