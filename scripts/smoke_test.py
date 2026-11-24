#!/usr/bin/env python3
"""Post-build smoke test for danicat.dev (GitHub Pages artifact in ./public)."""

import json
import os
import re
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

PUBLIC = Path("public")


def check_feeds() -> list[str]:
    errors = []
    feeds = [
        PUBLIC / "index.xml",
        PUBLIC / "en" / "index.xml",
        PUBLIC / "ja" / "index.xml",
        PUBLIC / "pt-br" / "index.xml",
        PUBLIC / "posts" / "index.xml",
    ]
    for feed in feeds:
        if not feed.exists():
            errors.append(f"Missing RSS feed: {feed}")
            continue
        try:
            tree = ET.parse(feed)
            root = tree.getroot()
        except ET.ParseError as exc:
            errors.append(f"Invalid XML in {feed}: {exc}")
            continue
        items = root.findall(".//item")
        if not items:
            errors.append(f"RSS feed has 0 items: {feed}")
            continue
        for idx, item in enumerate(items):
            link = (item.findtext("link") or "").strip()
            guid = (item.findtext("guid") or "").strip()
            pub_date = (item.findtext("pubDate") or "").strip()
            if not link:
                errors.append(f"{feed}: item #{idx} has empty <link>")
            if not guid:
                errors.append(f"{feed}: item #{idx} has empty <guid>")
            if "0001" in pub_date or not pub_date:
                errors.append(f"{feed}: item #{idx} ({link}) has invalid pubDate '{pub_date}'")
            if link.rstrip("/").endswith("/about") or link.rstrip("/").endswith("/chibi"):
                errors.append(f"{feed}: unexpected non-content page in feed: {link}")
    return errors


def check_index_json() -> list[str]:
    errors = []
    indices = [
        PUBLIC / "index.json",
        PUBLIC / "ja" / "index.json",
        PUBLIC / "pt-br" / "index.json",
    ]
    for idx_path in indices:
        if not idx_path.exists():
            errors.append(f"Missing search index: {idx_path}")
            continue
        try:
            data = json.loads(idx_path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            errors.append(f"Invalid JSON in {idx_path}: {exc}")
            continue
        if not isinstance(data, list) or len(data) == 0:
            errors.append(f"Empty or non-list index: {idx_path}")
            continue
        for i, item in enumerate(data):
            if not item.get("url"):
                errors.append(f"{idx_path}: item #{i} ({item.get('title')}) has empty url")
            if not item.get("lang"):
                errors.append(f"{idx_path}: item #{i} ({item.get('title')}) missing lang")
            if item.get("markdown_url") == "index.md":
                errors.append(f"{idx_path}: item #{i} ({item.get('title')}) has broken markdown_url 'index.md'")
    return errors


def check_html_pages() -> list[str]:
    errors = []
    ld_json_re = re.compile(
        r'<script[^>]*type=["\']application/ld\+json["\'][^>]*>(.*?)</script>',
        re.IGNORECASE | re.DOTALL,
    )
    noreferrer_re = re.compile(r'rel=["\'][^"\']*\bnoreferrer\b[^"\']*["\']', re.IGNORECASE)
    for html_path in PUBLIC.rglob("*.html"):
        text = html_path.read_text(encoding="utf-8", errors="replace")
        for m in ld_json_re.finditer(text):
            raw = m.group(1).strip()
            if not raw:
                continue
            try:
                json.loads(raw)
            except json.JSONDecodeError as exc:
                errors.append(f"Invalid JSON-LD in {html_path}: {exc}")
        if noreferrer_re.search(text):
            errors.append(f"Found rel='noreferrer' in {html_path}")
    return errors


def main() -> int:
    if not PUBLIC.exists():
        print("ERROR: ./public directory does not exist. Run hugo first.", file=sys.stderr)
        return 1
    errors = []
    errors.extend(check_feeds())
    errors.extend(check_index_json())
    errors.extend(check_html_pages())
    if errors:
        print(f"SMOKE TEST FAILED ({len(errors)} issues):", file=sys.stderr)
        for err in errors[:50]:
            print(f"  - {err}", file=sys.stderr)
        if len(errors) > 50:
            print(f"  ... and {len(errors) - 50} more", file=sys.stderr)
        return 1
    print("SMOKE TEST PASSED: RSS feeds, index.json, JSON-LD, and noreferrer checks all clean.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
