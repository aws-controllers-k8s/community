#!/usr/bin/env python3.8

import argparse
import collections
import os
import pathlib
import pprint
import sys

import github
import prettytable
from dataclasses import dataclass, asdict
from jinja2 import Environment
from jinja2.loaders import FileSystemLoader
from pathlib import Path
from typing import List, Mapping, NewType

from ackdiscover import awssdkgo, controller, ecrpublic, printer, service, project_stages, maintenance_phases

DEFAULT_CACHE_DIR = os.path.join(pathlib.Path.home(), ".cache", "ack-discover")

DisplayNameMap = Mapping[str, str]

@dataclass
class TemplateArgs:
    summary_table: str # Rendered service summary Markdown table
    services_table: str # Rendered service overview Markdown table
    service_display_names: DisplayNameMap # Maps from service package name to display name
    controllers: Mapping[str, controller.Controller] # Maps from service package name to controller

def write_services_page(args: TemplateArgs, page_output_path: Path):
    env = Environment(loader=FileSystemLoader(Path(__file__).parent / 'templates'), trim_blocks=True, lstrip_blocks=True)
    template = env.get_template("services.jinja2")
    with open(page_output_path, "w") as out:
        print(template.render(asdict(args)), file=out)

def filter_actively_developed_controllers(controllers: Mapping[str, controller.Controller]):
    active = {}

    for k, controller in controllers.items():
        if controller.project_stage != project_stages.NONE:
            active[k] = controller

    return active

def build_controller_table(controllers: List[controller.Controller], display_names: DisplayNameMap):
    t = prettytable.PrettyTable()
    t.set_style(prettytable.MARKDOWN)

    t.field_names = [
        "AWS Service",
        "Project Stage",
        "Maintenance Phase",
        "Latest Version"
    ]
    t.align = "r"
    t.align["Service"] = "l"
    for key, c in controllers.items():
        display_name = display_names[key]
        display_name_split = display_name.split(' ', 1)

        doc_anchor_link = display_name.lower().replace(" ", "-")
        service_name = f"{display_name_split[0]} [{display_name_split[1]}](#{doc_anchor_link})"

        proj_stage = f"`{c.project_stage}`"
        maint_phase = f"`{c.maintenance_phase}`"
        con_version = "n/a"
        if c.latest_release is not None and c.latest_release.controller_version is not None:
            con_version = f"[{c.latest_release.controller_version}]({c.latest_release.release_url})"

        t.add_row([
            service_name,
            proj_stage,
            maint_phase,
            con_version,
        ])
    return t


def build_summary_table(controllers: List[controller.Controller]):
    t = prettytable.PrettyTable()
    t.set_style(prettytable.MARKDOWN)

    counts = collections.defaultdict(int)

    t.field_names = [
        "Maintenance Phase",
        "# Services",
    ]
    t.align = "l"
    t.align["# Services"] = "r"
    for key, c in controllers.items():
        if c.maintenance_phase == maintenance_phases.NONE:
            continue
        maint_phase = f"`{c.maintenance_phase}`"
        counts[maint_phase] += 1

    for maint_phase, count in counts.items():
        t.add_row([
            maint_phase,
            count,
        ])
    return t


def create_display_names(controllers) -> DisplayNameMap:
    names = {}

    for key, c in controllers.items():
        proper_name = c.service.abbrev_name if c.service.abbrev_name else c.service.full_name

        # Some abbreviations don't prepend an AWS identifier
        if proper_name.startswith("AWS ") or proper_name.startswith("Amazon "):
            names[key] = proper_name
            continue

        # Some full names start with "AWS" but not "AWS " or "Amazon" but not
        # "Amazon " (e.g. "AmazonApiGatewayV2"). Here, we handle these cases to
        # ensure that the proper name is always "AWS" or "Amazon", a space, and
        # the abbreviated or full name stripped of "AWS" or "Amazon"
        # prefixes...
        if proper_name.startswith("AWS"):
            proper_name = proper_name[4:]
            names[key] = f"AWS {proper_name}"
            continue

        if proper_name.startswith("Amazon"):
            proper_name = proper_name[6:]

        # Default service names' proper name to "Amazon {service}"
        names[key] = f"Amazon {proper_name}"

    return names

def main(
    cache_dir: str,
    page_output_path: Path,
    gh_token: str,
    debug: bool = False
):
    os.makedirs(cache_dir, exist_ok=True)

    gh = github.Github(gh_token)
    writer = printer.Writer(printer.WriterArgs(debug=debug))

    try:
        ep_client = ecrpublic.get_client(writer)
    except Exception as e:
        print("ERROR: failed to get client for ECR Public:", str(e))
        sys.exit(255)

    repo = awssdkgo.get_repo(writer, gh_token, cache_dir)
    services = service.collect_all(writer, repo)

    controllers = controller.collect_all(writer, gh, ep_client, services)
    active_controllers = filter_actively_developed_controllers(controllers)

    display_names = create_display_names(controllers)

    summary_table = build_summary_table(active_controllers)
    table = build_controller_table(active_controllers, display_names)

    args = TemplateArgs(
        summary_table=summary_table,
        services_table=table,
        service_display_names=display_names,
        controllers=active_controllers
    )
    write_services_page(args, page_output_path / "services.md")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate documentation service release phase pages")
    parser.add_argument("--debug", action=argparse.BooleanOptionalAction,
        help="Enables debugging logging")
    parser.add_argument("--page_output_path", type=str, default="./content/docs/community",
        help="Relative path to the documentation output directory, relative to the `docs` directory")

    gh_token = os.environ.get("GITHUB_TOKEN")

    if not gh_token:
        print("ERROR: Please ensure GITHUB_TOKEN environment variable is set to the Github Personal Access Token (PAT) the script will use to query the Github API.")
        sys.exit(1)

    args = parser.parse_args()
    cache_dir = os.environ.get("CACHE_DIR", DEFAULT_CACHE_DIR)

    ret = main(
        cache_dir,
        Path(args.page_output_path),
        gh_token,
        args.debug
    )
    sys.exit(ret)
