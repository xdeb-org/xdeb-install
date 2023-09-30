import concurrent.futures
import gzip
import lzma
import requests
from dataclasses import dataclass
from pathlib import Path
from typing import Iterator

import yaml


def quoted_presenter(dumper, data):
    if data.split(".")[0].isdigit():
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style="'")

    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


yaml.add_representer(str, quoted_presenter)
yaml.representer.SafeRepresenter.add_representer(str, quoted_presenter)


@dataclass
class Package:
    name: str
    version: str
    url: str
    sha256: str


def parse_packages_file(url_prefix: str, packages_file: str) -> Iterator[Package]:
    packages = packages_file.split("\n\n")

    for package_data in packages:
        if not package_data:
            continue

        name = None
        version = None
        url = None
        sha256 = None

        for line in package_data.split("\n"):
            if line.startswith("Package:"):
                name = line.split(": ")[1]
                continue

            if line.startswith("Version:"):
                version = f'{line.split("Version: ")[1]}'
                continue

            if line.startswith("Filename:"):
                suffix = line.split("Filename: ")[1]
                url = f"{url_prefix}/{suffix}"
                continue

            if line.startswith("SHA256:"):
                sha256 = line.split("SHA256: ")[1]
                continue

        yield Package(name, version, url, sha256)


def get_packages(url: str, dist: str, component: str, architecture: str) -> Iterator[Package]:
    packages_file_url = f"{url}/dists/{dist}/{component}/binary-{architecture}/Packages"
    packages_file = requests.get(packages_file_url)

    # try Packages.xz, then Packages.gz
    if packages_file.status_code != 200:
        packages_file = requests.get(f"{packages_file_url}.xz")

        if packages_file.status_code != 200:
            packages_file = requests.get(f"{packages_file_url}.gz")

    if packages_file.status_code == 200:
        if packages_file.url.endswith("Packages.xz"):
            packages_file = lzma.decompress(packages_file.content, lzma.FORMAT_XZ).decode("utf-8")
        elif packages_file.url.endswith("Packages.gz"):
            packages_file = gzip.decompress(packages_file.content).decode("utf-8")
        else:
            packages_file = packages_file.text

        for package in parse_packages_file(url, packages_file):
            yield package


def dump_packages(url: str, dist: str, component: str, architecture: str, void_architecture: str) -> None:
    packages = get_packages(url, dist, component, architecture)

    doc = {
        "xdeb": {
            "packages": [{
                "name": package.name,
                "version": str(package.version),
                "url": package.url,
                "sha256": package.sha256
            } for package in packages]
        }
    }

    if doc["xdeb"]["packages"]:
        local_file = Path(f"repositories/{void_architecture}/{directory}/{dist}/{component}.yaml")
        local_file.parent.mkdir(parents=True, exist_ok=True)

        with local_file.open("w") as f:
            yaml.dump(doc, f, Dumper=yaml.CSafeDumper, default_flow_style=False, sort_keys=False)


void_architectures = Path("repositories").iterdir()

for void_architecture in void_architectures:
    if not void_architecture.is_dir():
        continue

    lists = void_architecture.joinpath("lists.yaml")

    if not lists.exists():
        continue

    with lists.open() as f:
        package_lists: dict = yaml.load(f, Loader=yaml.CSafeLoader)

        for directory, data in package_lists.items():
            with concurrent.futures.ThreadPoolExecutor(max_workers=20) as executor:
                future_map = [
                    executor.submit(dump_packages, data["url"], dist, component, data["architecture"], void_architecture.name)
                    for dist in data["dists"] for component in data["components"]
                ]

                for future in concurrent.futures.as_completed(future_map):
                    future.result()
