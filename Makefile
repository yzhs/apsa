all: install

.PHONY: install
install:
	systemctl --user stop apsa
	go install ./...
	systemctl --user start apsa

.PHONY: clean
clean:
	-rm ui/cli/cli
	-rm ui/web/web
