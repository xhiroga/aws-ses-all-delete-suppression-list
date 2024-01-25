BIN := /usr/local/bin

install: build;
	cp -f aws-ses-all-suppression-list $(BIN)

build:
	go build

uninstall:
	rm -f $(BIN)/aws-ses-all-suppression-list
