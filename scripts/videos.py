#!/usr/bin/env python3
"""
Synchronize YouTube channel uploads, the curated Conference Talks playlist,
and recorded talks in data/talks.json into data/videos.json using only
Python standard library and YouTube's ytInitialData payload.
"""

from concurrent.futures import ThreadPoolExecutor
from datetime import datetime
import json
from pathlib import Path
import re
from typing import Any
import urllib.request

ROOT_DIR = Path(__file__).resolve().parent.parent
VIDEOS_JSON_PATH = ROOT_DIR / "data" / "videos.json"
TALKS_JSON_PATH = ROOT_DIR / "data" / "talks.json"

HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
        "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
    ),
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


def fetch_html(url: str) -> str:
    req = urllib.request.Request(url, headers=HEADERS)
    with urllib.request.urlopen(req, timeout=15) as resp:
        return resp.read().decode("utf-8")


def discover_videos(url: str, kind: str) -> list[dict[str, str]]:
    html = fetch_html(url)
    vids = list(dict.fromkeys(re.findall(r'"(?:videoId|contentId)":"([a-zA-Z0-9_-]{11})"', html)))
    durs = re.findall(r'"thumbnailBadgeViewModel":\{"text":"(\d+:\d+(?::\d+)?)"', html)
    if not vids:
        raise RuntimeError(f"Failed to discover any video IDs from {url}")
    return [
        {"id": vid, "duration": dur, "kind": kind}
        for vid, dur in zip(vids, durs + [""] * len(vids))
    ]


def fetch_watch_metadata(vid: str) -> dict[str, Any]:
    url = f"https://www.youtube.com/watch?v={vid}"
    html = fetch_html(url)
    vc_m = re.search(r'"videoViewCountRenderer":\{"viewCount":\{"simpleText":"([^"]+)"', html)
    if not vc_m:
        raise RuntimeError(f"Failed to extract viewCount from ytInitialData for video '{vid}' ({url})")
    digits = re.sub(r"\D", "", vc_m.group(1))
    views = int(digits) if digits else 0

    title_m = re.search(r'"videoPrimaryInfoRenderer":\{"title":\{"runs":\[\{"text":"((?:\\.|[^"\\])*)"', html)
    owner_m = re.search(r'"videoOwnerRenderer":\{.*?"title":\{"runs":\[\{"text":"((?:\\.|[^"\\])*)"', html)
    dt_m = re.search(r'"dateText":\{"simpleText":"([^"]+)"\}', html)
    desc_m = re.search(r'"attributedDescription":\{"content":"((?:\\.|[^"\\])*)"', html)

    date_str = ""
    if dt_m and (dm := re.search(r"([A-Z][a-z]{2} \d{1,2}, \d{4})", dt_m.group(1))):
        date_str = datetime.strptime(dm.group(1), "%b %d, %Y").strftime("%Y-%m-%d")

    summary = ""
    if desc_m:
        raw_desc = json.loads('"' + desc_m.group(1) + '"')
        summary = raw_desc.strip().split("\n\n")[0].replace("\n", " ")[:280]

    return {
        "title": json.loads('"' + title_m.group(1) + '"').strip() if title_m else "",
        "channelName": json.loads('"' + owner_m.group(1) + '"').strip() if owner_m else "YouTube",
        "date": date_str,
        "summary": summary,
        "views": views,
    }


def main() -> None:
    videos_doc = json.loads(VIDEOS_JSON_PATH.read_text(encoding="utf-8"))
    talks_doc = json.loads(TALKS_JSON_PATH.read_text(encoding="utf-8"))

    channel_cfg = videos_doc["channel"]
    existing_by_id: dict[str, dict[str, Any]] = {v["id"]: v for v in videos_doc.get("videos", [])}
    talks_by_vid = {
        yt_id: t
        for t in talks_doc.get("talks", [])
        if (yt_id := extract_youtube_id(t.get("recording") or ""))
    }

    discovered = (
        discover_videos(f"{channel_cfg['url'].rstrip('/')}/videos", "demo")
        + discover_videos(channel_cfg["playlistUrl"], "talk")
    )

    added_ids: set[str] = set()
    for src in discovered:
        vid = src["id"]
        if vid not in existing_by_id:
            added_ids.add(vid)
            existing_by_id[vid] = {
                "id": vid,
                "duration": src["duration"],
                "kind": src["kind"],
                "url": f"https://www.youtube.com/watch?v={vid}",
            }

    yt_video_ids = [vid for vid, rec in existing_by_id.items() if extract_youtube_id(rec.get("url", ""))]
    with ThreadPoolExecutor(max_workers=8) as pool:
        watch_metas = dict(zip(yt_video_ids, pool.map(fetch_watch_metadata, yt_video_ids)))

    updated_count = 0
    for vid, rec in existing_by_id.items():
        if vid not in watch_metas:
            rec.setdefault("views", None)
            continue
        meta = watch_metas[vid]
        if vid in added_ids:
            talk = talks_by_vid.get(vid, {})
            existing_by_id[vid] = {
                "id": vid,
                "title": talk.get("title") or meta["title"],
                "date": talk.get("date") or meta["date"],
                "duration": rec.get("duration") or "",
                "views": meta["views"],
                "kind": rec["kind"],
                "channelName": meta["channelName"],
                "event": talk.get("event"),
                "url": f"https://www.youtube.com/watch?v={vid}",
                "thumbnail": f"https://img.youtube.com/vi/{vid}/hqdefault.jpg",
                "summary": talk.get("summary") or meta["summary"],
                "language": talk.get("language") or "en",
                "slides": talk.get("slides"),
                "source_code": talk.get("source_code"),
                "post": None,
                "tags": talk.get("tags") or ["golang"],
                "categories": talk.get("categories") or ["Software Engineering"],
            }
        else:
            if rec.get("kind") == "demo" and vid not in talks_by_vid and meta["title"] and rec.get("title") != meta["title"]:
                rec["title"] = meta["title"]
                updated_count += 1
            if rec.get("views") != meta["views"]:
                rec["views"] = meta["views"]
                updated_count += 1

    videos_doc["videos"] = sorted(existing_by_id.values(), key=lambda x: x.get("date") or "", reverse=True)
    if added_ids or updated_count:
        VIDEOS_JSON_PATH.write_text(json.dumps(videos_doc, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(f"Synced {len(videos_doc['videos'])} videos ({len(added_ids)} new, {updated_count} updated).")


if __name__ == "__main__":
    main()
