# Makefile

INSTALL_DIR := /usr/local/bin
MAN_DIR := /usr/share/man/man1

install:
    cp src/cdactl $(INSTALL_DIR)/cdactl
    chmod +x $(INSTALL_DIR)/cdactl
    cp man/cdactl.1 $(MAN_DIR)/cdactl.1
    gzip $(MAN_DIR)/cdactl.1

uninstall:
    rm -f $(INSTALL_DIR)/cdactl
    rm -f $(MAN_DIR)/cdactl.1.gz