# Description: Makefile for building the project

.PHONY: build clean

app_name = bulk

reset: clean build uninstall install
	@echo "Reset complete"

build:
	@echo "Building..."
	@go build -o ./gh-$(app_name) main.go

clean:
	@echo "Cleaning..."
	@rm -f gh-$(app_name)

install: 
	@echo "Installing..."
	@gh extension install .

uninstall:
	@echo "Uninstalling..."
	@gh extension remove $(app_name) 
