################################################################################

%define _posixroot        /
%define _root             /root
%define _bin              /bin
%define _sbin             /sbin
%define _srv              /srv
%define _home             /home
%define _opt              /opt
%define _lib32            %{_posixroot}lib
%define _lib64            %{_posixroot}lib64
%define _libdir32         %{_prefix}%{_lib32}
%define _libdir64         %{_prefix}%{_lib64}
%define _logdir           %{_localstatedir}/log
%define _rundir           %{_localstatedir}/run
%define _lockdir          %{_localstatedir}/lock/subsys
%define _cachedir         %{_localstatedir}/cache
%define _spooldir         %{_localstatedir}/spool
%define _crondir          %{_sysconfdir}/cron.d
%define _loc_prefix       %{_prefix}/local
%define _loc_exec_prefix  %{_loc_prefix}
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_libdir       %{_loc_exec_prefix}/%{_lib}
%define _loc_libdir32     %{_loc_exec_prefix}/%{_lib32}
%define _loc_libdir64     %{_loc_exec_prefix}/%{_lib64}
%define _loc_libexecdir   %{_loc_exec_prefix}/libexec
%define _loc_sbindir      %{_loc_exec_prefix}/sbin
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_datarootdir  %{_loc_prefix}/share
%define _loc_includedir   %{_loc_prefix}/include
%define _loc_mandir       %{_loc_datarootdir}/man
%define _rpmstatedir      %{_sharedstatedir}/rpm-state
%define _pkgconfigdir     %{_libdir}/pkgconfig

################################################################################

%define __ln              %{_bin}/ln
%define __touch           %{_bin}/touch
%define __service         %{_sbin}/service
%define __chkconfig       %{_sbin}/chkconfig
%define __useradd         %{_sbindir}/useradd
%define __groupadd        %{_sbindir}/groupadd
%define __getent          %{_bindir}/getent
%define __systemctl       %{_bindir}/systemctl

################################################################################

%define debug_package     %{nil}
%define pkg_name          bacula_exporter

################################################################################

Summary:         Bacula Exporter for Prometheus
Name:            bacula-exporter
Version:         1.0.0
Release:         0%{?dist}
Group:           Applications/System
License:         MIT
URL:             https://github.com/funbox/bacula_exporter

Source0:         https://github.com/funbox/%{pkg_name}/archive/v%{version}.tar.gz

BuildRoot:       %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:   golang >= 1.12

Requires:        kaosv >= 2.15

%if 0%{?rhel} >= 7
Requires:        systemd
%endif

Provides:        %{name} = %{version}-%{release}

################################################################################

%description
Bacula Exporter for Prometheus

################################################################################

%prep
%setup -q

%build
export GOPATH=$(pwd)

pushd src/github.com/funbox/%{pkg_name}
  %{__make} %{?_smp_mflags} deps
  %{__make} %{?_smp_mflags} all
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/%{pkg_name}
install -dm 755 %{buildroot}%{_sysconfdir}/logrotate.d
install -dm 755 %{buildroot}%{_initddir}
install -dm 755 %{buildroot}%{_unitdir}
install -dm 755 %{buildroot}%{_logdir}/%{pkg_name}
install -dm 755 %{buildroot}%{_rundir}/%{pkg_name}

install -pm 755 src/github.com/funbox/%{pkg_name}/%{pkg_name} \
                %{buildroot}%{_bindir}/

install -pm 644 src/github.com/funbox/%{pkg_name}/common/%{pkg_name}.knf \
                %{buildroot}%{_sysconfdir}/%{pkg_name}/%{pkg_name}.knf

install -pm 755 src/github.com/funbox/%{pkg_name}/common/%{name}.init \
                %{buildroot}%{_initddir}/%{name}

install -pm 644 src/github.com/funbox/%{pkg_name}/common/%{pkg_name}.logrotate \
                %{buildroot}%{_sysconfdir}/logrotate.d/%{pkg_name}

%if 0%{?rhel} >= 7
install -pDm 644 src/github.com/funbox/%{pkg_name}/common/%{name}.service \
                 %{buildroot}%{_unitdir}/
%endif
popd

%clean
rm -rf %{buildroot}

################################################################################

%pre
getent group %{name} >/dev/null || groupadd -r %{name}
getent passwd %{name} >/dev/null || useradd -r -M -g %{name} -s /sbin/nologin %{name}
exit 0

%post
if [[ $1 -eq 1 ]] ; then
    %if 0%{?rhel} <= 6
    %{__chkconfig} --add %{name}
    %else
    %{__systemctl} enable %{name}.service &>/dev/null || :
    %endif
fi

%preun
if [[ $1 -eq 0 ]] ; then
    %if 0%{?rhel} <= 6
    %{__service} %{name} stop &>/dev/null || :
    %{__chkconfig} --del %{service_name}
    %else
    %{__systemctl} --no-reload disable %{name}.service &>/dev/null || :
    %{__systemctl} stop %{name}.service &>/dev/null || :
    %endif
fi

%postun
%if 0%{?rhel} >= 7
if [[ $1 -ge 1 ]] ; then
    %{__systemctl} daemon-reload &>/dev/null || :
fi
%endif

################################################################################

%files
%defattr(-,root,root,-)
%attr(-,%{name},%{name}) %dir %{_logdir}/%{pkg_name}
%attr(-,%{name},%{name}) %dir %{_rundir}/%{pkg_name}
%config(noreplace) %{_sysconfdir}/%{pkg_name}/%{pkg_name}.knf
%config(noreplace) %{_sysconfdir}/logrotate.d/%{pkg_name}
%if 0%{?rhel} >= 7
%{_unitdir}/%{name}.service
%endif
%{_initddir}/%{name}
%{_bindir}/%{pkg_name}

################################################################################

%changelog
* Mon Jun 29 2020 Gleb Goncharov <inbox@funbox.ru> - 1.0.0-0
- Initial release
