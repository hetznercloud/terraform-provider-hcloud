export CGO_ENABLED:=0

VERSION=$(shell ./scripts/git-version.bash)
REPO=github.com/hetznercloud/terraform-provider-hcloud
LD_FLAGS=""

all: build

build: clean bin/terraform-provider-hcloud

bin/%:
	@go build -o bin/$* -ldflags $(LD_FLAGS) $(REPO)

test:
	@./scripts/test.bash

clean:
	@rm -rf bin/*

clean-release:
	@rm -rf _output

release: \
	clean \
	clean-release \
	_output/terraform-provider-hcloud-darwin-amd64.tar.gz \
	_output/terraform-provider-hcloud-linux-amd64.tar.gz \
	_output/terraform-provider-hcloud-linux-386.tar.gz \
	_output/terraform-provider-hcloud-linux-arm.tar.gz \
	_output/terraform-provider-hcloud-freebsd-amd64.tar.gz \
	_output/terraform-provider-hcloud-freebsd-386.tar.gz \
	_output/terraform-provider-hcloud-freebsd-arm.tar.gz \
	_output/terraform-provider-hcloud-openbsd-amd64.tar.gz \
	_output/terraform-provider-hcloud-openbsd-386.tar.gz \
	_output/terraform-provider-hcloud-solaris-amd64.tar.gz \
	_output/terraform-provider-hcloud-windows-amd64.tar.gz \
	_output/terraform-provider-hcloud-windows-386.tar.gz

bin/darwin-amd64/terraform-provider-hcloud: GOARGS = GOOS=darwin GOARCH=amd64
bin/linux-amd64/terraform-provider-hcloud: GOARGS = GOOS=linux GOARCH=amd64
bin/linux-386/terraform-provider-hcloud: GOARGS = GOOS=linux GOARCH=386
bin/linux-arm/terraform-provider-hcloud: GOARGS = GOOS=linux GOARCH=arm
bin/freebsd-amd64/terraform-provider-hcloud: GOARGS = GOOS=freebsd GOARCH=amd64
bin/freebsd-386/terraform-provider-hcloud: GOARGS = GOOS=freebsd GOARCH=386
bin/freebsd-arm/terraform-provider-hcloud: GOARGS = GOOS=freebsd GOARCH=arm
bin/openbsd-amd64/terraform-provider-hcloud: GOARGS = GOOS=openbsd GOARCH=amd64
bin/openbsd-386/terraform-provider-hcloud: GOARGS = GOOS=openbsd GOARCH=386
bin/solaris-amd64/terraform-provider-hcloud: GOARGS = GOOS=solaris GOARCH=amd64
bin/windows-amd64/terraform-provider-hcloud: GOARGS = GOOS=windows GOARCH=amd64
bin/windows-386/terraform-provider-hcloud: GOARGS = GOOS=windows GOARCH=386

bin/%/terraform-provider-hcloud:
	$(GOARGS) go build -o $@ -ldflags $(LD_FLAGS) -a $(REPO)

_output/terraform-provider-hcloud-%.tar.gz: NAME=terraform-provider-hcloud-$(VERSION)-$*
_output/terraform-provider-hcloud-%.tar.gz: DEST=_output/$(NAME)
_output/terraform-provider-hcloud-%.tar.gz: bin/%/terraform-provider-hcloud
	mkdir -p $(DEST)
	cp bin/$*/terraform-provider-hcloud $(DEST)
	cp LICENSE $(DEST)
	cp README.md $(DEST)
	tar zcvf $(DEST).tar.gz -C _output $(NAME)

.PHONY: all build clean test release
.SECONDARY: _output/terraform-provider-hcloud-linux-amd64
