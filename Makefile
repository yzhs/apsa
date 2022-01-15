all: install

.PHONY: install
install:
	go install ./...

.PHONY: clean
clean:
	-rm ui/cli/cli
	-rm ui/web/web
