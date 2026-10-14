# Renders a CycloneDX 1.6 document for sbom.sh
#   $descriptor[0]  component descriptor
#   $installed[0]   composer vendor/composer/installed.json, when the component has PHP dependencies
#   $manifest[0]    composer.json of the component
#   $vendored       vendored libraries with the detected versions
#   $version        component version

def expand: split("{version}") | join($version);
# SPDX identifiers are given as strings, SPDX expressions as {"expression": ...},
# other licenses as {"name": ...}
def licenses: [.[] | if type == "string" then {license: {id: .}} elif .expression then . else {license: .} end];
def composer_ref: "pkg:composer/\(.name)@\(.version)";
# A purl with {version} sets where the version goes, for example a git tag v{version}
def vendored_ref:
  if (.purl | contains("{version}")) then (.version as $v | .purl | split("{version}") | join($v))
  elif .version == "" then .purl
  else "\(.purl)@\(.version)" end;

def root_component($component; $ref):
  {
    "bom-ref": $ref,
    type: "application",
    supplier: {name: "Zabbix SIA", url: ["https://www.zabbix.com"]},
    name: $component.name,
    version: $version,
    description: $component.description,
    licenses: ($component.licenses | licenses),
    cpe: ($component.cpe | expand),
    purl: $ref
  };

def composer_component:
  {
    "bom-ref": composer_ref,
    type: "library",
    group: (.name | split("/")[0]),
    name: (.name | split("/")[1]),
    version,
    description,
    licenses: ((.license // []) | licenses),
    purl: composer_ref,
    externalReferences: [.source | select(.url) | {type: "vcs", url, comment: "reference: \(.reference)"}]
  };

def vendored_component:
  {
    "bom-ref": vendored_ref,
    type: (.type // "library"),
    name,
    licenses: (.licenses | licenses),
    purl: vendored_ref
  }
  + (if .version == "" then {} else {version} end)
  + (if .cpe then {cpe: (.version as $v | .cpe | split("{version}") | join($v))} else {} end)
  + (if .file then {properties: [{name: "zabbix:vendored-file", value: .file}]} else {} end);

def root_dependency($ref; $composer_json; $php_refs; $vendored):
  {
    ref: $ref,
    dependsOn: (
      [($composer_json.require // {}) | keys[] | $php_refs[.] // empty]
      + [$vendored[] | vendored_ref]
    )
  };

def composer_dependency($php_refs):
  {
    ref: composer_ref,
    dependsOn: [(.require // {}) | keys[] | $php_refs[.] // empty]
  };

def vendored_dependency:
  {ref: vendored_ref, dependsOn: []};

$descriptor[0] as $d
| ($installed[0] // {}) as $composer
| ($manifest[0] // {}) as $composer_json
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
      component: root_component($d.component; $root)
    },
    components: (
      [$php[] | composer_component]
      + [$vendored[] | vendored_component]
    ),
    dependencies: (
      [root_dependency($root; $composer_json; $php_refs; $vendored)]
      + [$php[] | composer_dependency($php_refs)]
      + [$vendored[] | vendored_dependency]
    )
  }
