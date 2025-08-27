# Makefile

INSTALL_DIR := /usr/local/bin
MAN_DIR := /usr/share/man/man1

build:
	go build -o cdactl

install: build
	install -m 755 cdactl $(INSTALL_DIR)/cdactl
	install -m 644 man/cdactl.1 $(MAN_DIR)/cdactl.1
	gzip -f $(MAN_DIR)/cdactl.1

uninstall:
    rm -f $(INSTALL_DIR)/cdactl
    rm -f $(MAN_DIR)/cdactl.1.gz
