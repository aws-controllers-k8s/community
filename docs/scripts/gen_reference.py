#!/usr/bin/env python3.8

from __future__ import annotations
import os

import yaml
import argparse
import logging

from pathlib import Path
from dataclasses import dataclass, asdict
from typing import Dict, List, Optional, Union
from jinja2 import Environment
from jinja2.loaders import FileSystemLoader

@dataclass(frozen=True)
class ResourceNames:
    kind: str
    listKind: str
    plural: str
    singular: str

@dataclass(frozen=True)
class Field:
    name: str
    description: str
    type: str
    required: bool
    contains: Optional[Union[str, List[Field]]]
    contains_description: Optional[str]

@dataclass(frozen=True)
class Resource:
    name: str
    service: str
    group: str
    api_version: str
    description: str
    scope: str
    names: ResourceNames
    spec: List[Field]
    status: List[Field]

    def output_path(self, short_name: str = None, extension: str = ".md") -> Path:
        return Path(f"{self.service if short_name is None else short_name}/{self.api_version}/{self.name.lower()}{extension}")


@dataclass(frozen=True)
class ResourceOverview:
    name: str
    api_version: str

@dataclass
class ServiceOverview:
    name: str
    resources: ResourceOverview
    service_metadata: ServiceMetadata

@dataclass(frozen=True)
class OverviewPageMeta:
    services: List[ServiceOverview]

@dataclass(frozen=True)
class ServiceMetadata:
    service: MetadataServiceDescription
    api_versions: List[MetadataAPIVersion]

@dataclass(frozen=True)
class MetadataServiceDescription:
    full_name: str
    short_name: str
    link: str
    documentation: str

@dataclass(frozen=True)
class MetadataAPIVersion:
    api_version: str
    status: str

def load_yaml(path: str) -> Dict:
    with open(path, "r") as stream:
        return yaml.safe_load(stream)

def convert_field(name: str, property: Dict, required=False) -> Field:
    field_type = property['type']
    contains = None
    contains_description = None
    if field_type == "array":
        contains_description = property['items']['description'] if 'description' in property['items'] else ''
        contains = None
        # Check for array of objects
        if 'properties' in property['items']:
            contains = []
            for subfield_name, subfield in property['items']['properties'].items():
                subfield_required  = subfield_name in property['items']['required'] if 'required' in property else False
                contains.append(convert_field(subfield_name, subfield, subfield_required))
        else:
            contains = property['items']['type']

    elif field_type == "object":
        contains = []
        # TODO (RedbackThomson): Handle arbitrary maps
        if 'additionalProperties' in property:
            contains = property['additionalProperties']['type']
        else:
            for subfield_name, subfield in property['properties'].items():
                subfield_required = subfield_name in property['required'] if 'required' in property else False
                contains.append(convert_field(subfield_name, subfield, subfield_required))

    return Field(
        name=name,
        description=property.get('description', ''),
        type=property['type'],
        required=required,
        contains=contains,
        contains_description=contains_description
    )

def convert_crd_to_resource(crd: Dict, service) -> Resource:
    ver = crd['spec']['versions'][0]
    schema = ver['schema']['openAPIV3Schema']
    res_spec = []
    res_status = []

    names = ResourceNames(
        kind=crd['spec']['names']['kind'],
        listKind=crd['spec']['names']['listKind'],
        plural=crd['spec']['names']['plural'],
        singular=crd['spec']['names']['singular']
    )

    spec = schema['properties']['spec']
    spec_properties = spec['properties']
    for field_name, field in spec_properties.items():
        required  = field_name in spec['required'] if 'required' in spec else False
        res_spec.append(convert_field(field_name, field, required))

    status = schema['properties']['status']
    status_properties = status['properties']
    for field_name, field in status_properties.items():
        required  = field_name in status['required'] if 'required' in status else False
        res_status.append(convert_field(field_name, field, required))

    return Resource(
        name=crd['spec']['names']['kind'],
        service=service,
        group=crd['spec']['group'],
        api_version=ver['name'],
        description=spec.get('description', ''),
        scope=crd['spec']['scope'],
        names=names,
        spec=res_spec,
        status=res_status
    )

def write_service_pages(
    service: str,
    metadata: ServiceMetadata,
    service_path: Path,
    page_output_path: Path,
) -> List[ResourceOverview]:
    resources = []

    for crd_path in Path(service_path).rglob('*.yaml'):
        crd = load_yaml(crd_path)

        resource = convert_crd_to_resource(crd, service)
        resources.append(ResourceOverview(name=resource.name,
            api_version=resource.api_version))

        page_path = page_output_path / resource.output_path(service)
        page_path.parent.mkdir(parents=True, exist_ok=True)

        env = Environment(loader=FileSystemLoader(Path(__file__).parent / 'templates'), trim_blocks=True, lstrip_blocks=True)
        template = env.get_template("resource.jinja2")
        with open(page_path, "w") as out:
            print(template.render(asdict(resource)), file=out)
    
    return resources

def write_overview_page(services: List[ServiceOverview], output_path: Path):
    overview_path = output_path / "overview.yaml"

    with open(overview_path, "w") as out:
        page_meta = OverviewPageMeta(services=services)
        yaml.dump(asdict(page_meta), out)

def load_metadata_config(metadata_config_path: Path) -> ServiceMetadata:
    yaml_dict = load_yaml(metadata_config_path)
    description = MetadataServiceDescription(**yaml_dict.get("service", {}))
    api_versions = []
    for version in yaml_dict.get("api_versions", []):
        api_versions.append(MetadataAPIVersion(**version))
    return ServiceMetadata(description, api_versions)

def main(
    gopath: Path,
    metadata_config_path: Path,
    go_src_parent: Path,
    service_bases_path: Path,
    page_output_path: Path,
    data_output_path: Path
):
    overviews = []

    src_parent = gopath / go_src_parent

    for controller_repo in src_parent.glob("*-controller"):
        if not controller_repo.is_dir() or \
            controller_repo.stem.startswith("template"):
            continue

        # Get service name from repository
        service = controller_repo.stem[:-len("-controller")]

        service_bases_path = controller_repo / bases_path

        if not service_bases_path.exists():
            logging.error(f"Service base path {service_bases_path} does not exist")
            continue

        metadata_config = controller_repo / metadata_config_path
        if not metadata_config.exists():
            logging.error(f"Could not find metadata file in {service} repository")
            continue

        metadata = load_metadata_config(metadata_config)

        resources = write_service_pages(service, metadata, service_bases_path, page_output_path)
        overview = ServiceOverview(service, resources, metadata)

        overviews.append(overview)

    write_overview_page(overviews, data_output_path)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate documentation reference pages")
    parser.add_argument("--metadata_config_path", type=str, default="./metadata.yaml",
        help="Relative path to the `metadata` file, relative to the service controller root")
    parser.add_argument("--bases_path", type=str, default="./config/crd/bases",
        help="Relative path to the `bases` directory, relative to the service controller root")
    parser.add_argument("--gopath", type=str, default=os.environ.get('GOPATH'),
        help="Path of the GOPATH - Defaults to $GOPATH")
    parser.add_argument("--go_src_parent", type=str, default="./src/github.com/aws-controllers-k8s",
        help="Relative path to the ACK src path, relative to the GOPATH")
    parser.add_argument("--page_output_path", type=str, default="./content/reference",
        help="Relative path to the documentation output directory, relative to the `docs` directory")
    parser.add_argument("--data_output_path", type=str, default="./data",
        help="Relative path to the documentation output directory, relative to the `docs` directory")

    args = parser.parse_args()

    metadata_config_path = Path(args.metadata_config_path)
    bases_path = Path(args.bases_path)
    gopath = Path(args.gopath)
    go_src_parent = Path(args.go_src_parent)
    page_output_path = Path(args.page_output_path)
    data_output_path = Path(args.data_output_path)

    main(gopath, metadata_config_path, go_src_parent, bases_path, page_output_path, data_output_path)