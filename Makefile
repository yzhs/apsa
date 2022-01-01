all: cli web

lib: Bleve.go Datastructures.go markdown_recipe.go modernist_recipe.go Util.go
	go build

cli: lib ui/cli/Cli.go
	cd ui/cli && go build

web: lib ui/web/Web.go
	cd ui/web && go build

.PHONY: install
install: lib web
	cp ui/cli/cli ~/bin/apsa
	systemctl --user stop apsa
	cp ui/web/web ~/bin/apsa-web
	#systemctl --user start apsa

.PHONY: clean
clean:
	-rm ui/cli/cli
	-rm ui/web/web
