.PHONY: build install clean

BIN_DIR = bin
TARGET = $(BIN_DIR)/stx

build:
	go build -o $(TARGET) ./cmd/stx

install: build
	cp $(TARGET) /usr/local/bin/stx

clean:
	rm -rf $(BIN_DIR)
