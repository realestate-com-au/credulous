# don't build debug versions
%define debug_package %{nil}

Name:		credulous
Version:	0.1.3
Release:	1%{?dist}
Summary:	Secure AWS credential storage, rotation and redistribution

Group:		Applications/System
License:	MIT
URL:		https://github.com/realestate-com-au/credulous
Source0:	credulous-%{version}.tar.gz

BuildRequires:	golang, mercurial, bzr, git, pandoc

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
pandoc -s -w man credulous.md -o credulous.1

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir} \
	$RPM_BUILD_ROOT/%{_sysconfdir}/bash_completion.d \
	$RPM_BUILD_ROOT/%{_sysconfdir}/profile.d \
	$RPM_BUILD_ROOT/%{_mandir}/man1

cp credulous $RPM_BUILD_ROOT/%{_bindir}
cp credulous.bash_completion $RPM_BUILD_ROOT/%{_sysconfdir}/bash_completion.d/credulous.bash_completion
cp credulous.sh $RPM_BUILD_ROOT/%{_sysconfdir}/profile.d/credulous.sh
chmod 0755 $RPM_BUILD_ROOT/%{_bindir}/credulous
cp credulous.1 $RPM_BUILD_ROOT/%{_mandir}/man1/credulous.1

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
%attr(0755,root,root)		%{_bindir}/credulous
%attr(0644,root,root)		%{_mandir}/man1/credulous.1.gz
%attr(0644,root,root)		%{_sysconfdir}/bash_completion.d/credulous.bash_completion
%attr(0644,root,root)		%{_sysconfdir}/profile.d/credulous.sh

%changelog
