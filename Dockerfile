FROM centos:8

ENV TERM=xterm \
    PATH=/usr/local/go/bin:$PATH

RUN set -eux && \
    dnf -y install \
            --disablerepo "*" \
            --enablerepo "baseos" \
            --enablerepo "appstream" \
        wget && \
    cd /tmp/ && \
    ARCH_SUFFIX="$(arch)"; \
    case "$ARCH_SUFFIX" in \
        x86_64) \
            url='https://dl.google.com/go/go1.17.2.linux-amd64.tar.gz'; \
            sha256='f242a9db6a0ad1846de7b6d94d507915d14062660616a61ef7c808a76e4f1676'; \
            ;; \
        aarch64) \
            url='https://dl.google.com/go/go1.17.2.linux-arm64.tar.gz'; \
            sha256='a5a43c9cdabdb9f371d56951b14290eba8ce2f9b0db48fb5fc657943984fd4fc'; \
            ;; \
        ppc64le) \
            url='https://golang.org/dl/go1.17.2.linux-ppc64le.tar.gz'; \
            sha256='12e2dc7e0ffeebe77083f267ef6705fec1621cdf2ed6489b3af04a13597ed68d'; \
            ;; \
        *) echo "Unknown ARCH_SUFFIX=${ARCH_SUFFIX-}"; exit 1 ;; \
    esac; \
    wget -O go.tgz.asc "$url.asc" && \
    wget -O go.tgz "$url" --progress=dot:giga && \
    echo "$sha256 *go.tgz" | sha256sum -c - && \
    GNUPGHOME="$(mktemp -d)"; export GNUPGHOME; \
    gpg --batch --keyserver keyserver.ubuntu.com --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC EC91 7721 F63B D38B 4796' && \
    gpg --batch --verify go.tgz.asc go.tgz && \
    gpgconf --kill all && \
    rm -rf "$GNUPGHOME" go.tgz.asc && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz

RUN set -eux && \
    dnf -y install \
            --disablerepo "*" \
            --enablerepo "baseos" \
            --enablerepo "appstream" \
            --enablerepo "powertools" \
        rpmdevtools \
        dnf-plugins-core \
        rpmlint && \
    rpm -ivh https://repo.zabbix.com/zabbix/5.4/rhel/8/SRPMS/zabbix-5.4.6-1.el8.src.rpm && \
    cd /root/rpmbuild/ && \
    dnf -y builddep \
            --disablerepo "*" \
            --enablerepo "baseos" \
            --enablerepo "appstream" \
            --enablerepo "powertools" \
        SPECS/zabbix.spec && \
#    rpmlint SPECS/zabbix.spec && \
    rpmbuild -ba SPECS/zabbix.spec
