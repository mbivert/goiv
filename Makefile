# Default installation directory.
#	make install dir=$HOME/bin
dir ?= /bin/
mandir ?= /usr/share/man/man1/
root ?= root
group ?= root

.PHONY: all
all: goiv update-doc

goiv: goiv.go
	@echo Building goiv...
	@go build goiv.go

.PHONY: update-doc
update-doc: goiv.1
	@echo Updating README.md...
	@(cat README.md.base; echo; echo '# goiv(1) - Go Image Viewer';echo; COLUMNS=80 man ./goiv.1 | sed 's/^/    /') > README.md

.PHONY: install
install: goiv
	@echo Installing goiv to ${dir}/goiv...
	@install -o ${root} -g ${group} -m 755 goiv ${dir}/goiv
	@echo Installing goiv.1 to ${mandir}/goiv.1...
	@install -o ${root} -g ${group} -m 644 goiv.1 ${mandir}/goiv.1

.PHONY: uninstall
uninstall:
	@echo Removing ${dir}/goiv...
	@rm -f ${dir}/goiv
	@echo Removing ${mandir}/goiv.1...
	@rm -f ${mandir}/goiv.1
