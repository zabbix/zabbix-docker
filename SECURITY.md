# Disclosure policy

In Zabbix we use the term "**responsible disclosure**", which means we have a policy on how we are disclosing all security issues that come to our attention, but only after the issues has been resolved and all customers with support contracts are given time to upgrade or patch their installations.

We kindly ask that when you are reporting a security issue, you follow the same guidelines and share the details only with the Zabbix Security team.

# Before reporting the issue:

Make sure that the issue you are submitting is not related to server configuration, 3rd party scripts and utilities. In order to avoid any possible issues with server configuration we advise Zabbix users to read [Best practices for secure Zabbix setup](https://www.zabbix.com/documentation/current/manual/installation/requirements/best_practices).

# How to report a security issue?

[Create a new issue](https://support.zabbix.com/secure/CreateIssue.jspa) in the Zabbix Security Reports (ZBXSEC) section of the public bug tracker describing the problem (and a proposed solution if possible) in detail. This way we can ensure that only Zabbix security team and the reporter have access to the case.

The following information will be helpful for Zabbix Security team:

- Date and time when you identified the security defect.
- Affected Zabbix version range.
- Type of security issue you are reporting, e.g.: XSS, CSRF, SQLi, RCE.
- Affected components, e.g.: Image, Frontend, Server, Agent, API.
- Any details you can provide, e.g. screenshots, screen recordings, http(s) transaction logs, POC exploits (please do not share any evidence via unauthenticated file sharing services and avoid sharing sensitive information, as if Zabbix Security team decides that this issue does not fit Security defect description it might be moved to ZBX project and the issue will be visible to all users).
- Step by step instructions to reproduce the issue as the problem might not be easily identifiable.

# How Zabbix deals with reported security issues:

1. Zabbix Security team reviews the issue and evaluates its potential impact.
2. If the security issue is found not to be related to security then the issue will be moved to ZBX project.
3. Zabbix security team works on the issue to provide a solution and keeps all details on the problem until the next version of Zabbix is out.
4. New images are created and made available for download on [https://www.zabbix.com/container_images](https://www.zabbix.com/container_images), [https://hub.docker.com/u/zabbix](http://hub.docker.com/r/zabbix/) and [Red Hat Certified Container Catalog](https://catalog.redhat.com/software/containers/search?vendor_name=Zabbix%20Sia&p=1)
5. Zabbix requests [CVE identifiers](https://cve.mitre.org/) for the security issue.
6. Clients with valid support agreements are emailed giving a period of time when it is possible to upgrade before the issue becomes known to the public.
7. A public announcement for the community is made.
