# Security policy

The Zabbix container images follow the [Zabbix security policy](https://www.zabbix.com/zabbix_security_policy), which describes the disclosure policy, how security issues are reported and how Zabbix deals with them.

Security fixes are released as new images for the Zabbix versions that are supported according to the [Zabbix life cycle and release policy](https://www.zabbix.com/life_cycle_and_release_policy). This branch builds the images of the current Zabbix development version.

# Reporting a security issue

Report security issues, including issues in the images themselves, by [creating a new issue](https://support.zabbix.com/secure/CreateIssue.jspa) in the restricted Zabbix Security Reports (ZBXSEC) project, as described in the [Zabbix security policy](https://www.zabbix.com/zabbix_security_policy). Please do not report security issues on GitHub.

# Verifying images

The images published on Docker Hub are built by the [GitHub Actions workflows](https://github.com/zabbix/zabbix-docker/tree/trunk/.github/workflows) of this repository.

The examples below use the `zabbix/zabbix-server-pgsql:alpine-trunk` tag. Other images and tags can be checked by adjusting the image reference, workflow and branch where applicable. Tags are updated whenever an image is rebuilt, for example for security updates, while signatures and attestations belong to a specific image digest. For reproducible results, refer to the image by its digest, for example `zabbix/zabbix-server-pgsql@sha256:...`.

The tools used are [cosign](https://github.com/sigstore/cosign), [GitHub CLI](https://cli.github.com/), [Docker Buildx](https://github.com/docker/buildx) and, for vulnerability scanning, [Grype](https://github.com/anchore/grype).

## Signature

Images are signed with [Sigstore](https://www.sigstore.dev/) keyless signing by the workflows of this repository. Scheduled trunk builds run from the current default release branch while checking out the `trunk` branch. The current published image was signed by the workflow from the `7.4` branch:

```bash
cosign verify \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    --certificate-identity 'https://github.com/zabbix/zabbix-docker/.github/workflows/images_build.yml@refs/heads/7.4' \
    zabbix/zabbix-server-pgsql:alpine-trunk
```

## Build provenance

The [SLSA provenance](https://slsa.dev/) attestation records the repository, commit and workflow run that built the image. To verify the current trunk image against the workflow identity described above:

```bash
gh attestation verify oci://docker.io/zabbix/zabbix-server-pgsql:alpine-trunk \
    --repo zabbix/zabbix-docker \
    --signer-workflow zabbix/zabbix-docker/.github/workflows/images_build.yml \
    --source-ref refs/heads/7.4
```

Linux images additionally carry provenance recorded by BuildKit. It can be inspected with:

```bash
docker buildx imagetools inspect zabbix/zabbix-server-pgsql:alpine-trunk \
    --format '{{ json (index .Provenance "linux/amd64").SLSA }}'
```

Verifying the signature or SLSA attestation confirms who built the image and from which source. Inspecting the BuildKit provenance alone does not verify the signer identity. None of these checks mean that the image is free from vulnerabilities.

## Software Bill of Materials (SBOM)

Linux images provide up to three SBOMs. The Zabbix component SBOM is present only in images that contain a Zabbix component; for example, the `snmptraps` image does not contain one.

| SBOM | Format | Content | Platforms |
|------|--------|---------|-----------|
| Image SBOM attestation | CycloneDX, signed | Operating system packages and application dependencies, including the embedded Zabbix component SBOM when present | `linux/amd64` |
| Zabbix component SBOM | CycloneDX, file in `/usr/share/doc/zabbix/sbom/` of the image | Zabbix component with its CPE, third-party code bundled in Zabbix sources (for example Duktape, jQuery, MD5 implementation), PHP and Java libraries, agent 2 plugins | all supported platforms where applicable |
| BuildKit SBOM attestation | SPDX | Packages and dependencies detected in the image filesystem, including embedded SBOM data when present | each platform |

The image SBOM attestation is recommended for vulnerability scanning and tools such as [Dependency-Track](https://dependencytrack.org/) because it provides a normalized CycloneDX document. It describes the `linux/amd64` image only; for other platforms use the BuildKit SBOM attestation.

```bash
gh attestation verify oci://docker.io/zabbix/zabbix-server-pgsql:alpine-trunk \
    --repo zabbix/zabbix-docker \
    --signer-workflow zabbix/zabbix-docker/.github/workflows/images_build.yml \
    --source-ref refs/heads/7.4 \
    --predicate-type https://cyclonedx.org/bom \
    --bundle-from-oci \
    --format json \
    --jq '.[0].verificationResult.statement.predicate' | grype
```

To list only the vulnerabilities that already have a fixed version available, add `--only-fixed`:

```bash
gh attestation verify oci://docker.io/zabbix/zabbix-server-pgsql:alpine-trunk \
    --repo zabbix/zabbix-docker \
    --signer-workflow zabbix/zabbix-docker/.github/workflows/images_build.yml \
    --source-ref refs/heads/7.4 \
    --predicate-type https://cyclonedx.org/bom \
    --bundle-from-oci \
    --format json \
    --jq '.[0].verificationResult.statement.predicate' | grype --only-fixed
```

The BuildKit SBOM attestation is available for each platform of the image, for example `linux/arm64`:

```bash
docker buildx imagetools inspect zabbix/zabbix-server-pgsql:alpine-trunk \
    --format '{{ json (index .SBOM "linux/arm64").SPDX }}' | grype
```

Windows images are signed and come with build provenance, but do not provide SBOMs yet.
