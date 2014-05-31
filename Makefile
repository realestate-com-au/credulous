# This option defines which mock configuration to use -- see /etc/mock for
# the available configuration files for your system.
MOCK_CONFIG=epel-6-x86_64
SHELL=/bin/bash
DIST=$(shell grep "config_opts.*dist.*" /etc/mock/$(MOCK_CONFIG).cfg | awk '{ print $$3 }' | cut -f2 -d\' )

SRCS=$(shell ls -1 *.go | grep -v _test.go ) credulous.bash_completion
TESTS=credulous_test.go credentials_test.go crypto_test.go git_test.go \
	testkey testkey.pub credential.json newcreds.json

SPEC=rpm/credulous.spec
SPEC_TMPL=rpm/credulous.spec.tmpl
NAME=$(shell grep '^Name:' $(SPEC_TMPL) | awk '{ print $$2 }' )
VERSION=$(shell git describe --abbrev=0 2>/dev/null )
ifeq ($(strip $(VERSION)), )
	VERSION=$(shell git describe --tags --abbrev=0 )
endif
RELEASE=$(shell grep '^Release:' $(SPEC_TMPL) | awk '{ print $$2 }' | sed -e 's/%{?dist}/.$(DIST)/' )

MOCK_RESULT=/var/lib/mock/$(MOCK_CONFIG)/result

NVR=$(NAME)-$(VERSION)-$(RELEASE)
MOCK_SRPM=$(NVR).src.rpm
RPM=$(NVR).x86_64.rpm
TGZ=$(NAME)-$(VERSION).tar.gz

.DEFAULT: all
.PHONY: debianpkg

all: mock

# This is a dirty hack for building on ubuntu build agents in Travis.
rpmbuild: sources
	@mkdir -p 	$(HOME)/rpmbuild/SOURCES \
			$(HOME)/rpmbuild/SRPMS \
			$(HOME)/rpmbuild/RPMS \
			$(HOME)/rpmbuild/SPECS \
			$(HOME)/rpmbuild/BUILD \
			$(HOME)/rpmbuild/BUILDROOT
	cp $(NAME)-$(VERSION).tar.gz $(HOME)/rpmbuild/SOURCES
	rpmbuild -bs --target x86_64 --nodeps rpm/credulous.spec
	rpmbuild -bb --target x86_64 --nodeps rpm/credulous.spec

# Create the source tarball with N-V prefix to match what the specfile expects
sources:
	@echo "Building for version '$(VERSION)'"
	tar czvf $(TGZ) --transform='s|^|src/github.com/realestate-com-au/credulous/|' $(SRCS) $(TESTS)

debianpkg:
	@echo Build Debian packages
	sed -i -e 's/==VERSION==/$(VERSION)/' debian-pkg/DEBIAN/control
	mkdir -p debian-pkg/DEBIAN/usr/bin
	cp $(HOME)/gopath/bin/credulous debian-pkg/DEBIAN/usr/bin
	chmod 0755 debian-pkg/DEBIAN/usr/bin/credulous
	dpkg-deb --build debian-pkg
	mv debian-pkg.deb $(NAME)_$(VERSION)_amd64.deb

mock: mock-rpm
	@echo "BUILD COMPLETE; RPMS are in ."

mock-rpm: mock-srpm
	mock -r $(MOCK_CONFIG) --rebuild $(MOCK_SRPM)
	cp $(MOCK_RESULT)/$(RPM) .

mock-srpm: sources
	@echo "DIST is $(DIST)"
	@echo "RELEASE is $(RELEASE)"
	# mock -r $(MOCK_CONFIG) --init
	sed -e 's/==VERSION==/$(VERSION)/' $(SPEC_TMPL) > $(SPEC)
	mock -r $(MOCK_CONFIG) --buildsrpm --spec $(SPEC) --sources .
	rm -f $(SPEC)
	cp $(MOCK_RESULT)/$(MOCK_SRPM) .

clean:
	rm -f $(MOCK_SRPM) $(RPM) $(TGZ)

allclean:
	mock -r $(MOCK_CONFIG) --clean
