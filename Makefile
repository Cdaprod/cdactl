# Makefile

INSTALL_DIR := /usr/local/bin
MAN_DIR := /usr/share/man/man1
LIB_DIR := /usr/local/lib

install:
	cp src/cdactl $(INSTALL_DIR)/cdactl
	chmod +x $(INSTALL_DIR)/cdactl
	cp man/cdactl.1 $(MAN_DIR)/cdactl.1
	gzip -f $(MAN_DIR)/cdactl.1
	cp src/cda-common.sh $(LIB_DIR)/cda-common.sh
	chmod +x $(LIB_DIR)/cda-common.sh

uninstall:
	rm -f $(INSTALL_DIR)/cdactl
	rm -f $(MAN_DIR)/cdactl.1.gz
	rm -f $(LIB_DIR)/cda-common.sh