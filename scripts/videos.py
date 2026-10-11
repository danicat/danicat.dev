#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "requests",
#     "rich",
# ]
# ///

"""
Synchronize YouTube channel uploads (@danikopaizen), the curated Conference Talks
playlist (PL9dBIQfOJu-6uny3OQLlG5MbJ_hYWCJjr), and recorded talks in data/talks.json
into data/videos.json.
"""

import argparse
import json
import re
import sys
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import Any, Dict, List

import requests
from rich.console import Console
from rich.table import Table

console = Console(width=120)

ROOT_DIR = Path(__file__).resolve().parent.parent
VIDEOS_JSON_PATH = ROOT_DIR / "data" / "videos.json"
TALKS_JSON_PATH = ROOT_DIR / "data" / "talks.json"

CHANNEL_ID = "UCRa5xBK_o3-HPAqph86z_LQ"
PLAYLIST_ID = "PL9dBIQfOJu-6uny3OQLlG5MbJ_hYWCJjr"
CHANNEL_VIDEOS_URL = "https://www.youtube.com/@danikopaizen/videos"
PLAYLIST_URL = f"https://www.youtube.com/playlist?list={PLAYLIST_ID}"

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
    "Accept-Language": "en-US,en;q=0.9",
    "Cookie": "SOCS=CAISNQgDEitib3FfaWRlbnRpdHlmcm9udGVuZHVpc2VydmVyXzIwMjMwODI5LjA3X3AxGgJlbiACGgYIgJnPpwY",
}


def extract_youtube_id(url: str) -> str:
    if not url:
        return ""
    if "youtu.be/" in url:
        return url.split("youtu.be/")[1].split("?")[0].split("&")[0]
    if "watch?v=" in url:
        return url.split("watch?v=")[1].split("&")[0].split("?")[0]
    return ""


def fetch_channel_videos() -> List[Dict[str, Any]]:
    resp = requests.get(CHANNEL_VIDEOS_URL, headers=HEADERS, timeout=15)
    resp.raise_for_status()
    seen: set[str] = set()
    items: List[Dict[str, Any]] = []
    for vid in re.findall(r'"videoId":"([a-zA-Z0-9_-]{11})"', resp.text):
        if vid in seen:
            continue
        seen.add(vid)
        title = fetch_oembed_title(vid)
        items.append(
            {
                "id": vid,
                "title": title,
                "date": "",
                "duration": "",
                "channelName": "Daniela Petruzalek",
                "summary": "",
                "kind": "demo",
            }
        )
    return items


def fetch_playlist_items() -> List[Dict[str, Any]]:
    resp = requests.get(PLAYLIST_URL, headers=HEADERS, timeout=15)
    resp.raise_for_status()
    match = re.search(r"var ytInitialData = (\{.*?\});</script>", resp.text)
    if not match:
        raise RuntimeError(f"Failed to locate ytInitialData on playlist page {PLAYLIST_URL}")
    data = json.loads(match.group(1))
    discovered: List[Dict[str, Any]] = []

    def walk(obj: Any) -> None:
        if isinstance(obj, dict):
            if "lockupViewModel" in obj:
                lvm = obj["lockupViewModel"]
                cid = lvm.get("contentId")
                if cid:
                    meta = lvm.get("metadata", {}).get("lockupMetadataViewModel", {})
                    title = meta.get("title", {}).get("content", "")
                    rows = meta.get("metadata", {}).get("contentMetadataViewModel", {}).get("metadataRows", [])
                    parts: List[str] = []
                    for r in rows:
                        for p in r.get("metadataParts", []):
                            t = p.get("text", {}).get("content", "")
                            if t:
                                parts.append(t)
                    raw_str = json.dumps(lvm)
                    dur_m = re.search(r'"text":\s*"(\d+:\d+(?::\d+)?)"', raw_str)
                    duration = dur_m.group(1) if dur_m else ""
                    discovered.append(
                        {
                            "id": cid,
                            "title": title,
                            "channelName": parts[0] if parts else "",
                            "duration": duration,
                            "kind": "talk",
                        }
                    )
            else:
                for v in obj.values():
                    walk(v)
        elif isinstance(obj, list):
            for v in obj:
                walk(v)

    walk(data)
    return discovered


def fetch_video_watch_metadata(vid: str) -> Dict[str, str]:
    url = f"https://www.youtube.com/watch?v={vid}"
    resp = requests.get(url, headers=HEADERS, timeout=15)
    resp.raise_for_status()
    html = resp.text
    pub_m = re.search(r'"publishDate":"([^"]+)"', html)
    dur_m = re.search(r'"lengthSeconds":"(\d+)"', html)
    desc_m = re.search(r'"shortDescription":"((?:\\.|[^"\\])*)"', html)
    duration = ""
    if dur_m:
        secs = int(dur_m.group(1))
        h, rem = divmod(secs, 3600)
        m, s = divmod(rem, 60)
        duration = f"{h}:{m:02d}:{s:02d}" if h else f"{m}:{s:02d}"
    summary = ""
    if desc_m:
        raw_desc = json.loads('"' + desc_m.group(1) + '"')
        summary = raw_desc.strip().split("\n\n")[0].replace("\n", " ")[:280]
    return {
        "date": pub_m.group(1)[:10] if pub_m else "",
        "duration": duration,
        "summary": summary,
    }


def fetch_oembed_title(vid: str) -> str:
    url = f"https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v={vid}&format=json"
    resp = requests.get(url, headers=HEADERS, timeout=15)
    resp.raise_for_status()
    return (resp.json().get("title") or "").strip()


def main() -> None:
    parser = argparse.ArgumentParser(description="Sync YouTube channel & playlist videos into data/videos.json")
    parser.add_argument("--dry-run", action="store_true", help="Preview new videos without writing data/videos.json")
    args = parser.parse_args()

    if not VIDEOS_JSON_PATH.exists():
        raise FileNotFoundError(f"Missing {VIDEOS_JSON_PATH}")
    if not TALKS_JSON_PATH.exists():
        raise FileNotFoundError(f"Missing {TALKS_JSON_PATH}")

    videos_doc = json.loads(VIDEOS_JSON_PATH.read_text(encoding="utf-8"))
    talks_doc = json.loads(TALKS_JSON_PATH.read_text(encoding="utf-8"))

    existing_by_id: Dict[str, Dict[str, Any]] = {v["id"]: v for v in videos_doc.get("videos", [])}
    talks_by_vid: Dict[str, Dict[str, Any]] = {}
    for t in talks_doc.get("talks", []):
        rec = t.get("recording") or ""
        yt_id = extract_youtube_id(rec)
        if yt_id:
            talks_by_vid[yt_id] = t

    console.print("[bold cyan]Fetching @danikopaizen channel videos and Conference Talks playlist...[/bold cyan]")
    channel_items = fetch_channel_videos()
    playlist_items = fetch_playlist_items()

    added_count = 0
    updated_count = 0
    for src in channel_items + playlist_items:
        vid = src["id"]
        if vid in existing_by_id:
            if src["kind"] == "demo" and vid not in talks_by_vid:
                live_title = fetch_oembed_title(vid) or src["title"]
                if live_title and existing_by_id[vid].get("title") != live_title:
                    console.print(
                        f"[yellow]Updated title for {vid}:[/yellow] '{existing_by_id[vid].get('title')}' -> '{live_title}'"
                    )
                    existing_by_id[vid]["title"] = live_title
                    updated_count += 1
            continue
        console.print(f"[yellow]New video discovered:[/yellow] {vid} ({src['title']})")
        watch_meta = fetch_video_watch_metadata(vid)
        talk_match = talks_by_vid.get(vid, {})
        record = {
            "id": vid,
            "title": talk_match.get("title") or src["title"],
            "date": talk_match.get("date") or src.get("date") or watch_meta["date"],
            "duration": src.get("duration") or watch_meta["duration"],
            "kind": src["kind"],
            "channelName": src.get("channelName") or "YouTube",
            "event": talk_match.get("event"),
            "url": f"https://www.youtube.com/watch?v={vid}",
            "thumbnail": f"https://img.youtube.com/vi/{vid}/hqdefault.jpg",
            "summary": talk_match.get("summary") or src.get("summary") or watch_meta["summary"],
            "language": talk_match.get("language") or "en",
            "slides": talk_match.get("slides"),
            "source_code": talk_match.get("source_code"),
            "post": None,
            "tags": talk_match.get("tags") or ["golang"],
            "categories": talk_match.get("categories") or ["Software Engineering"],
        }
        existing_by_id[vid] = record
        added_count += 1

    sorted_videos = sorted(existing_by_id.values(), key=lambda x: x.get("date") or "", reverse=True)
    videos_doc["videos"] = sorted_videos

    table = Table(title=f"Videos Catalog ({len(sorted_videos)} total, {added_count} added, {updated_count} updated)")
    table.add_column("Date", style="cyan", no_wrap=True)
    table.add_column("Kind", style="magenta", no_wrap=True)
    table.add_column("ID", style="dim", no_wrap=True)
    table.add_column("Title", style="bold")
    table.add_column("Channel / Event", style="green")
    for v in sorted_videos:
        table.add_row(
            v.get("date") or "-",
            v.get("kind") or "talk",
            v["id"],
            v["title"][:55],
            v.get("event") or v.get("channelName") or "-",
        )
    console.print(table)

    if not args.dry_run and (added_count > 0 or updated_count > 0):
        VIDEOS_JSON_PATH.write_text(json.dumps(videos_doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        console.print(
            f"[bold green]Updated {VIDEOS_JSON_PATH} ({added_count} new, {updated_count} updated).[/bold green]"
        )
    elif args.dry_run:
        console.print("[dim]Dry run complete; no files written.[/dim]")
    else:
        console.print("[bold green]All channel & playlist videos are already in sync![/bold green]")


if __name__ == "__main__":
    main()
