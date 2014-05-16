# don't build debug versions
%define debug_package %{nil}

Name:		credulous
Version:	0.1.2
Release:	1%{?dist}
Summary:	Secure AWS credential storage, rotation and redistribution

Group:		Applications/System
License:	MIT
URL:		https://github.com/realestate-com-au/credulous
Source0:	credulous-%{version}.tar.gz

BuildRequires:	golang, mercurial, bzr, git

%description
Credulous securely saves and retrieves AWS credentials, storing
them in an encrypted local repository.

%prep
%setup -n src/github.com/realestate-com-au/credulous

%build
export GOPATH=$RPM_BUILD_DIR
go get -t
go test
go build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir}

cp credulous $RPM_BUILD_ROOT/%{_bindir}
chmod 0755 $RPM_BUILD_ROOT/%{_bindir}/credulous

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
%attr(0755,root,root)		%{_bindir}/credulous

%changelog
* Fri May 16 2014 Colin Panisset <colin.panisset@rea-group.com> 0.1.2-1
- Initial RPM packaging
