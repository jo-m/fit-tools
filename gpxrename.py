#!/usr/bin/env python3
"""Rename GPX files based on the track name in their metadata."""

import os
import sys
import xml.etree.ElementTree as ET


def extract_track_name(path):
    tree = ET.parse(path)
    root = tree.getroot()

    # Handle GPX namespace
    ns = ""
    if root.tag.startswith("{"):
        ns = root.tag.split("}")[0] + "}"

    # Try <trk><name>, then <metadata><name>, then <rte><name>
    for parent_tag in ["trk", "metadata", "rte"]:
        name_el = root.find(f".//{ns}{parent_tag}/{ns}name")
        if name_el is not None and name_el.text and name_el.text.strip():
            return name_el.text.strip()

    return None


def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <file.gpx> ...", file=sys.stderr)
        sys.exit(1)

    for path in sys.argv[1:]:
        try:
            name = extract_track_name(path)
        except Exception as e:
            print(f"{path}: error: {e}", file=sys.stderr)
            continue

        if not name:
            print(f"{path}: no track name found, skipping", file=sys.stderr)
            continue

        # Sanitize filename
        name = name.replace("/", "_").replace("\0", "")
        dest = os.path.join(os.path.dirname(path), name + ".gpx")

        if os.path.abspath(path) == os.path.abspath(dest):
            print(f"{path}: already named correctly")
            continue

        if os.path.exists(dest):
            print(f"{path}: target {dest} already exists, skipping", file=sys.stderr)
            continue

        os.rename(path, dest)
        print(f"{path} -> {dest}")


if __name__ == "__main__":
    main()
