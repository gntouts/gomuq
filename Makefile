.PHONY: default
default: build ;

define UAMQ_CONFIG
# uamq sample configuration file

mqtt:
  host: localhost
  port: 1883

redis:
  host: localhost
  port: 6379

uart:
  name: arduino
  baudrate: 115200

log:
  target: /home/hass/log/uamq/uamq.log
endef
export UAMQ_CONFIG

LIBUSB := $(shell dpkg -s libusb-1.0-0-dev 2> /dev/null | grep Status | grep -o ok)
GOLANG := $(shell go version 2> /dev/null | grep -o go | uniq)

libusb_ok:
ifeq ($(LIBUSB), ok)
else
	$(info libusb-1.0-0-dev is not installed, consider doing apt-get install libusb-1.0-0-dev)
	$(error dependencies missing)
endif

go_ok:
ifeq ($(GOLANG), go)
else
	$(info go is not installed, for more info see https://go.dev/doc/install)
	$(error dependencies missing)
endif

dependencies: libusb_ok go_ok

prepare:
	@mkdir -p dist
	@go mod tidy
	@go mod vendor

nt: dependencies prepare
	@GOOS=linux GOARCH=$$(dpkg --print-architecture) go build -o dist/nt_$$(dpkg --print-architecture) ./cmd/nt 

uamq: dependencies prepare
	@GOOS=linux GOARCH=$$(dpkg --print-architecture) go build -o dist/uamq_$$(dpkg --print-architecture) ./cmd/uamq

config: prepare
	@echo "$$UAMQ_CONFIG" > dist/uamq.sample.conf

build: nt uamq config

install:
	@mkdir -p dist/install
	@cp dist/nt_$$(dpkg --print-architecture) dist/install/nt
	@mv dist/install/nt /usr/local/bin
	@cp dist/uamq_$$(dpkg --print-architecture) dist/install/uamq
	@mv dist/install/uamq /usr/local/bin
	@rm -fr dist/install

uninstall:
	@rm /usr/local/bin/uamq
	@rm /usr/local/bin/nt

clean:
	@rm -fr dist release
