# Renders a CycloneDX 1.6 document for sbom.sh
#   $descriptor[0]  component descriptor
#   $installed[0]   composer vendor/composer/installed.json, when the component has PHP dependencies
#   $manifest[0]    composer.json of the component
#   $vendored       vendored libraries with the detected versions
#   $version        component version

def expand: split("{version}") | join($version);
def licenses: [.[] | {license: {id: .}}];
def composer_ref: "pkg:composer/\(.name)@\(.version)";
def vendored_ref: if .version == "" then .purl else "\(.purl)@\(.version)" end;

$descriptor[0] as $d
| ($installed[0] // {}) as $composer
| ($composer["dev-package-names"] // []) as $dev
| [($composer.packages // [])[] | select(.name as $name | $dev | index($name) | not)] as $php
| ($php | map({key: .name, value: composer_ref}) | from_entries) as $php_refs
| ($d.component.purl | expand) as $root
| {
    "$schema": "http://cyclonedx.org/schema/bom-1.6.schema.json",
    bomFormat: "CycloneDX",
    specVersion: "1.6",
    version: 1,
    metadata: {
      component: {
        "bom-ref": $root,
        type: "application",
        supplier: {name: "Zabbix SIA", url: ["https://www.zabbix.com"]},
        name: $d.component.name,
        version: $version,
        description: $d.component.description,
        licenses: ($d.component.licenses | licenses),
        cpe: ($d.component.cpe | expand),
        purl: $root
      }
    },
    components: (
      [$php[] | {
        "bom-ref": composer_ref,
        type: "library",
        group: (.name | split("/")[0]),
        name: (.name | split("/")[1]),
        version,
        description,
        licenses: ((.license // []) | licenses),
        purl: composer_ref,
        externalReferences: [.source | select(.url) | {type: "vcs", url, comment: "reference: \(.reference)"}]
      }]
      + [$vendored[] | {
        "bom-ref": vendored_ref,
        type: (.type // "library"),
        name,
        licenses: (.licenses | licenses),
        purl: vendored_ref,
        properties: [{name: "zabbix:sbom:file", value: .file}]
      } + (if .version == "" then {} else {version} end)]
    ),
    dependencies: (
      [{
        ref: $root,
        dependsOn: (
          [($manifest[0].require // {}) | keys[] | $php_refs[.] // empty]
          + [$vendored[] | vendored_ref]
        )
      }]
      + [$php[] | {
        ref: composer_ref,
        dependsOn: [(.require // {}) | keys[] | $php_refs[.] // empty]
      }]
      + [$vendored[] | {ref: vendored_ref, dependsOn: []}]
    )
  }
