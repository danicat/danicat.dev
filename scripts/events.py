# /// script
# requires-python = ">=3.11"
# dependencies = [
#     "google-genai",
#     "pydantic",
#     "feedparser",
#     "rich",
#     "requests",
# ]
# ///

import argparse
import datetime
import json
import os
import re
import shutil
import sys
from pathlib import Path
from typing import List, Literal, Optional

import feedparser
import requests
from google import genai
from google.genai import types
from pydantic import BaseModel, Field
from rich.console import Console
from rich.prompt import Confirm, Prompt
from rich.table import Table
from rich import box

# Standardized 120-column budget for clean readable terminal tables
console = Console(width=120)

SPEAKERDECK_RSS_URL = "https://speakerdeck.com/danicat.rss"
SESSIONIZE_API_URL = os.environ.get("SESSIONIZE_API_URL")
TALKS_JSON_PATH = Path(__file__).resolve().parent.parent / "data" / "talks.json"
MODEL_NAME = os.environ.get("GEMINI_MODEL", "gemini-3.7-flash")
LOCATION = os.environ.get("GEMINI_LOCATION", "global")
VALID_CATEGORIES = [
    "Agentic Coding",
    "Agent Development",
    "Applied GenAI",
    "Perspectives",
    "Software Engineering",
]


class EventItem(BaseModel):
    name: str = Field(description="Name of the conference, meetup, or event")
    date: str = Field(description="Date of the event in YYYY-MM-DD format")
    location: str = Field(description="City and country of the event or 'Online'")
    url: Optional[str] = Field(default=None, description="Official event website URL if known")


class EventRecord(BaseModel):
    id: str = Field(description="Deterministic slug: YYYY-MM-DD-event-name-talk-title")
    title: str = Field(description="Clean talk title")
    event: Optional[str] = Field(default=None, description="Name of the conference, meetup, or event")
    date: Optional[str] = Field(default=None, description="Date of presentation in YYYY-MM-DD format")
    location: Optional[str] = Field(default=None, description="City and country of the event or 'Online'")
    url: Optional[str] = Field(default=None, description="Official event website URL if known")
    summary: str = Field(description="Concise 1-3 sentence summary of the talk")
    image: Optional[str] = Field(default=None, description="Custom thumbnail or slide preview URL")
    slides: Optional[str] = Field(default=None, description="Slides URL")
    source_code: Optional[str] = Field(default=None, description="GitHub or source code repository URL")
    recording: Optional[str] = Field(default=None, description="Recording or YouTube URL")
    codelab: Optional[str] = Field(default=None, description="Hands-on codelab or workshop tutorial URL")
    materials: Optional[str] = Field(default=None, description="Supporting materials URL")
    language: Optional[str] = Field(default="en", description="Delivery language code, e.g. 'en', 'pt-BR', 'ja'")
    tags: List[str] = Field(description="Lowercase relevant tags")
    categories: List[str] = Field(description="Strictly valid taxonomy categories")


class JudgeVerdict(BaseModel):
    decision: Literal["accept", "auto_corrected", "merge"] = Field(
        description="'accept' if record is pristine; 'auto_corrected' if normalized/taxonomy-fixed; 'merge' if it should enrich an existing record"
    )
    target_id: Optional[str] = Field(
        default=None,
        description="ID/slug of existing talk to merge into if decision is 'merge'"
    )
    auto_fixes_applied: List[str] = Field(
        default_factory=list,
        description="List of automatic corrections applied to ensure quality"
    )
    final_record: EventRecord = Field(
        description="The validated, sanitized, and normalized event record"
    )


BLACKLISTED_EVENT_NAMES = {
    "speaker deck",
    "speaker deck presentation",
    "google cloud online presentation",
    "conference presentation",
    "online presentation",
    "accepted conference session",
}


class SpeakerDeckExtraction(BaseModel):
    title: str = Field(description="Canonical, clean talk title with any bracketed prefixes or event names stripped")
    event_name: Optional[str] = Field(
        default=None,
        description="The real-world conference or meetup name ONLY if explicitly written in the title or description (e.g. 'FOSDEM 2026', 'DevFest Pisa 2026', 'GoLab 2025', 'Python Brasil 14'). If no event name is mentioned, MUST be None. NEVER guess or invent an event name."
    )
    event_date: Optional[str] = Field(
        default=None,
        description="Actual date of presentation in YYYY-MM-DD format if explicitly specified in description text. If not mentioned, MUST be None."
    )
    location: Optional[str] = Field(
        default=None,
        description="City and country of the event if explicitly mentioned in the text. Otherwise None."
    )
    summary: str = Field(description="Concise 1-2 sentence summary of the talk content")


def extract_speakerdeck_llm(client: genai.Client, raw_title: str, raw_desc: str, pub_date: str) -> SpeakerDeckExtraction:
    """Use Gemini LLM to extract clean title and strictly factual metadata from unstructured Speaker Deck notes without guessing."""
    prompt = f"""
Extract strictly factual presentation metadata from this Speaker Deck upload.

Raw Title: {raw_title}
Description: {raw_desc}
Published Upload Date: {pub_date}

ABSOLUTE FACTUALITY RULES:
1. title: Clean the title by stripping leading bracketed tags (e.g. '[DevFest Pisa 2026]' -> 'Build...').
2. event_name: Extract the event name ONLY IF EXPLICITLY STATED in the raw title or description (e.g. 'DevFest Pisa 2026', 'FOSDEM 2026', 'GoLab 2025', 'Python Brasil 14').
   NEVER invent, guess, or synthesize an event name. If no event name is explicitly mentioned in the text, return null.
3. event_date: Extract the presentation date (YYYY-MM-DD) if explicitly mentioned in the description. If no presentation date is stated, return null.
4. location: Extract the city/country ONLY if explicitly mentioned in the description. Otherwise return null.
5. summary: Concise 1-2 sentence summary of the talk content.
"""
    try:
        response = client.models.generate_content(
            model=MODEL_NAME,
            contents=prompt,
            config=types.GenerateContentConfig(
                response_mime_type="application/json",
                response_schema=SpeakerDeckExtraction,
                temperature=0.0,
            ),
        )
        return response.parsed
    except Exception as e:
        clean_title = re.sub(r"^\[.*?\]\s*", "", raw_title).strip()
        return SpeakerDeckExtraction(
            title=clean_title,
            event_name=None,
            event_date=None,
            location=None,
            summary=raw_desc.split("\n")[0] if raw_desc else clean_title,
        )


def slugify(text: str) -> str:
    """Generate a clean URL/ID slug from text."""
    text = text.lower()
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[\s_-]+", "-", text).strip("-")
    return text


def generate_deterministic_id(date_str: str, event_name: str, talk_title: str) -> str:
    """Generate deterministic event ID: YYYY-MM-DD-event-talk."""
    d = date_str if date_str else datetime.date.today().isoformat()
    return f"{slugify(d)}-{slugify(event_name)}-{slugify(talk_title)}"


def normalize_url(url: Optional[str]) -> str:
    """Normalize URL by stripping query parameters and trailing slashes."""
    if not url:
        return ""
    return url.split("?")[0].rstrip("/")


def fetch_speakerdeck_preview(slide_url: Optional[str]) -> Optional[str]:
    """Fetch exact presentation preview image from Speaker Deck HTML/player."""
    if not slide_url or "speakerdeck.com" not in slide_url:
        return None
    try:
        r = requests.get(slide_url, timeout=10)
        if r.status_code != 200:
            return None
        m = re.search(r'class="speakerdeck-embed"[^>]+data-id="([a-f0-9]+)"', r.text)
        if not m:
            m = re.search(r'data-id="([a-f0-9]+)"', r.text)
        if not m:
            return None
        deck_id = m.group(1)
        r2 = requests.get(f"https://speakerdeck.com/player/{deck_id}", timeout=10)
        if r2.status_code != 200:
            return None
        m2 = re.search(r'(https://files\.speakerdeck\.com/presentations/' + deck_id + r'/preview_slide_0\.jpg[^\s"\'<>]*)', r2.text)
        if m2:
            return m2.group(1)
    except Exception:
        pass
    return None


def extract_thumbnail_url(entry) -> Optional[str]:
    """Extract slide preview image from Speaker Deck RSS media tags or player."""
    if "media_content" in entry and len(entry["media_content"]) > 0:
        url = entry["media_content"][0].get("url")
        if url:
            return url
    if "media_thumbnail" in entry and len(entry["media_thumbnail"]) > 0:
        url = entry["media_thumbnail"][0].get("url")
        if url:
            return url
    for link in entry.get("links", []):
        if link.get("type", "").startswith("image/"):
            return link.get("href")
    link_url = entry.get("link")
    if link_url and "speakerdeck.com" in link_url:
        return fetch_speakerdeck_preview(link_url)
    return None


def load_talks() -> dict:
    if not TALKS_JSON_PATH.exists():
        console.print(f"[red]ERROR:[/red] Talks data file not found at {TALKS_JSON_PATH}")
        sys.exit(1)
    with open(TALKS_JSON_PATH, "r", encoding="utf-8") as f:
        return json.load(f)


def save_talks(data: dict):
    with open(TALKS_JSON_PATH, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        f.write("\n")


def evaluate_and_heal_record(
    client: genai.Client,
    raw_title: str,
    raw_description: str,
    pub_date: str,
    slide_url: Optional[str],
    thumbnail: Optional[str],
    existing_talks: List[dict],
) -> JudgeVerdict:
    """LLM-as-a-Judge: Extracts, validates, auto-heals, and deduplicates the talk entry."""
    existing_summaries = [
        {"id": t.get("id") or t.get("title"), "title": t.get("title"), "events": [e.get("name") for e in t.get("events", [])]}
        for t in existing_talks[:15]
    ]

    prompt = f"""
You are an expert technical editor and automated quality judge for a developer portfolio.
Evaluate and structure this incoming talk/presentation entry.

Incoming Data:
- Raw Title: {raw_title}
- Description: {raw_description}
- Published/Date: {pub_date}
- Slide URL: {slide_url}

Existing Recent Talks in System:
{json.dumps(existing_summaries, indent=2)}

Allowed Taxonomy Categories (Must choose 1 or 2 strictly from this list):
{json.dumps(VALID_CATEGORIES)}

Quality Gate & Factuality Rules:
1. Title: Strip bracketed prefixes (e.g. '[DevFest Pisa 2026]').
2. Event & Date: Extract ONLY event names explicitly stated in the input (e.g. 'FOSDEM 2026', 'DevFest Pisa 2026', 'GoLab 2025', 'Python Brasil 14').
   CRITICAL: NEVER invent or guess a conference/event name. If no event name is explicitly mentioned in the text, use 'Unspecified'.
3. Slug ID: Construct deterministic id = "YYYY-MM-DD-eventname-talktitle" (slugified).
4. Taxonomy Auto-Correction: Map any informal categories (e.g. 'Golang', 'AI Agents') to the allowed categories.
5. Deduplication / Merge: Check if this is an update to an existing talk in the system. If it matches an existing talk, set decision = 'merge' and target_id to the existing talk's id.
6. If everything is clean or auto-corrected, set decision = 'auto_corrected' or 'accept'.
"""

    try:
        response = client.models.generate_content(
            model=MODEL_NAME,
            contents=prompt,
            config=types.GenerateContentConfig(
                response_mime_type="application/json",
                response_schema=JudgeVerdict,
                temperature=0.0,
            ),
        )
        verdict: JudgeVerdict = response.parsed
        if thumbnail and not verdict.final_record.image:
            verdict.final_record.image = thumbnail
        if slide_url and not verdict.final_record.slides:
            verdict.final_record.slides = slide_url
        return verdict

    except Exception as e:
        console.print(f"[yellow]Judge fallback triggered ({e}). Applying rule-based self-healing.[/yellow]")
        clean_title = re.sub(r"^\[.*?\]\s*", "", raw_title).strip()
        safe_date = pub_date[:10] if pub_date and len(pub_date) >= 10 else datetime.date.today().isoformat()
        safe_id = generate_deterministic_id(safe_date, "Unspecified", clean_title)

        fallback_record = EventRecord(
            id=safe_id,
            title=clean_title,
            summary=raw_description.split("\n")[0] if raw_description else clean_title,
            image=thumbnail,
            events=[
                EventItem(
                    name="Unspecified",
                    date=safe_date,
                    location="-",
                    url=None,
                )
            ],
            slides=slide_url,
            source_code=None,
            recording=None,
            tags=["tech", "development"],
            categories=["AI & Development"],
        )
        return JudgeVerdict(
            decision="auto_corrected",
            target_id=None,
            auto_fixes_applied=[f"Applied emergency fallback due to LLM error: {e}"],
            final_record=fallback_record,
        )


def sync_speakerdeck(data: dict, api_key: str) -> int:
    """Sync talks from Speaker Deck RSS with self-healing judge."""
    existing_talks = data.get("talks", [])
    existing_slide_urls = {
        normalize_url(t.get("slides")) for t in existing_talks if t.get("slides")
    }

    console.print(f"\n[bold cyan]Fetching Speaker Deck RSS:[/bold cyan] [blue]{SPEAKERDECK_RSS_URL}[/blue]...")
    feed = feedparser.parse(SPEAKERDECK_RSS_URL)

    if not feed.entries:
        console.print("[yellow]No entries found in Speaker Deck feed.[/yellow]")
        return 0

    new_entries = [
        e for e in feed.entries
        if normalize_url(e.get("link", "")) and normalize_url(e.get("link", "")) not in existing_slide_urls
        and (e.get("published_parsed") is None or e.get("published_parsed").tm_year >= 2024)
    ]

    if not new_entries:
        console.print("[green]✓ talks.json is up to date with Speaker Deck![/green]")
        return 0

    console.print(f"[bold]Found {len(new_entries)} new deck(s) on Speaker Deck to evaluate:[/bold]")
    client = genai.Client(api_key=api_key)
    added_count = 0

    for entry in new_entries:
        title = entry.get("title", "")
        description = entry.get("description", "") or entry.get("summary", "")
        pub_date = entry.get("published", "")
        slide_url = entry.get("link", "")
        thumbnail = extract_thumbnail_url(entry)

        console.print(f"\nEvaluating '[bold]{title}[/bold]' with {MODEL_NAME} Judge...")
        verdict = evaluate_and_heal_record(
            client=client,
            raw_title=title,
            raw_description=description,
            pub_date=pub_date,
            slide_url=slide_url,
            thumbnail=thumbnail,
            existing_talks=existing_talks,
        )

        record_dict = verdict.final_record.model_dump()

        if verdict.decision == "merge" and verdict.target_id:
            # Merge into existing talk
            merged = False
            for t in data["talks"]:
                if t.get("id") == verdict.target_id or t.get("title") == verdict.final_record.title:
                    if slide_url:
                        t["slides"] = slide_url
                    if thumbnail and not t.get("image"):
                        t["image"] = thumbnail
                    merged = True
                    break
            if merged:
                console.print(f"[green]✓ Merged slides into existing talk: {verdict.target_id}[/green]")
                added_count += 1
                continue

        # Insert new self-healed record
        data["talks"].insert(0, record_dict)
        added_count += 1
        console.print(f"[green]✓ [{verdict.decision.upper()}] Approved & Added:[/green] {record_dict['title']} (ID: {record_dict['id']})")
        if verdict.auto_fixes_applied:
            for fix in verdict.auto_fixes_applied:
                console.print(f"  [dim]↳ Auto-fix: {fix}[/dim]")

    return added_count


def sync_sessionize(data: dict, api_key: Optional[str]) -> int:
    """Sync speaker bio and sessions from Sessionize JSON API endpoint."""
    if not SESSIONIZE_API_URL:
        console.print("\n[yellow]Notice: SESSIONIZE_API_URL is not set. Skipping Sessionize sync.[/yellow]")
        return 0

    console.print(f"\n[bold cyan]Fetching Sessionize Profile & Sessions...[/bold cyan]")
    try:
        resp = requests.get(SESSIONIZE_API_URL, timeout=10)
        resp.raise_for_status()
        sess_data = resp.json()
    except Exception as e:
        console.print(f"[red]Failed to fetch Sessionize data:[/red] {e}")
        return 0

    # 1. Update Speaker Bio
    if "speaker" in sess_data:
        sp = sess_data["speaker"]
        data["speaker"] = {
            "firstName": sp.get("firstName", ""),
            "lastName": sp.get("lastName", ""),
            "tagline": sp.get("tagline", ""),
            "bio": sp.get("bio", ""),
            "speakerProfileUrl": sp.get("speakerProfileUrl", "https://sessionize.com/daniela"),
            "photoUrl": sp.get("photoUrl", ""),
            "photoLargeUrl": sp.get("photoLargeUrl", ""),
        }
        console.print(f"[green]✓ Updated speaker bio & profile photo for {sp.get('firstName')} {sp.get('lastName')}[/green]")

    # 2. Check Sessions
    sessions = sess_data.get("sessions", [])
    if not sessions:
        return 0

    existing_talks = data.get("talks", [])
    existing_titles = {slugify(t.get("title", "")) for t in existing_talks}
    added_count = 0

    client = genai.Client(api_key=api_key) if api_key else None

    for sess in sessions:
        title = sess.get("title", "")
        desc = sess.get("description", "").strip()
        slug_title = slugify(title)

        if slug_title not in existing_titles:
            console.print(f"Found new Sessionize session: [bold]{title}[/bold]")
            if client:
                verdict = evaluate_and_heal_record(
                    client=client,
                    raw_title=title,
                    raw_description=desc,
                    pub_date=datetime.date.today().isoformat(),
                    slide_url=None,
                    thumbnail=None,
                    existing_talks=existing_talks,
                )
                rec = verdict.final_record.model_dump()
                rec["events"] = None
                rec["id"] = slugify(rec["title"])
            else:
                rec = {
                    "id": slugify(title),
                    "title": title,
                    "summary": desc.split("\n")[0] if desc else title,
                    "image": None,
                    "events": None,
                    "slides": None,
                    "source_code": None,
                    "recording": None,
                    "tags": ["sessionize", "talk"],
                    "categories": ["AI & Development"],
                }

            data["talks"].insert(0, rec)
            existing_titles.add(slug_title)
            added_count += 1
            console.print(f"[green]✓ Added session:[/green] {title}")

    return added_count


def cmd_import(args):
    """Import talks from Speaker Deck and/or Sessionize."""
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        console.print("[red]ERROR:[/red] GEMINI_API_KEY environment variable is not set.", file=sys.stderr)
        sys.exit(1)

    data = load_talks()
    run_all = not args.speakerdeck and not args.sessionize
    deck_count = 0
    sess_count = 0

    if run_all or args.speakerdeck:
        deck_count = sync_speakerdeck(data, api_key)

    if run_all or args.sessionize:
        sess_count = sync_sessionize(data, api_key)

    total_added = deck_count + sess_count
    if total_added > 0 or run_all or args.sessionize:
        # Backfill deterministic IDs if missing on existing items
        for t in data.get("talks", []):
            if "id" not in t or not t["id"]:
                first_event = t.get("events", [{}])[0]
                t_date = first_event.get("date", "2026-01-01")
                t_name = first_event.get("name", "Event")
                t["id"] = generate_deterministic_id(t_date, t_name, t.get("title", "talk"))

        save_talks(data)
        console.print(f"\n[bold green]✓ talks.json updated! Added {total_added} item(s).[/bold green]")


def cmd_list(args):
    """List all events across local and live remote sources with Date, Event, Location, Title, Description, Data Source, and Status."""
    data = load_talks()
    talks = data.get("talks", [])

    local_slide_urls = {normalize_url(t.get("slides")) for t in talks if t.get("slides")}
    local_titles = {slugify(t.get("title", "")) for t in talks if t.get("title")}

    # 1. Fetch live remote data (unless --local is passed)
    remote_speakerdeck = []
    remote_sessionize = []

    if not getattr(args, "local", False):
        # Speaker Deck
        try:
            feed = feedparser.parse(SPEAKERDECK_RSS_URL)
            if feed.entries:
                remote_speakerdeck = feed.entries
        except Exception as e:
            console.print(f"[yellow]Warning: Could not fetch live Speaker Deck feed ({e})[/yellow]")

        # Sessionize
        if SESSIONIZE_API_URL:
            try:
                resp = requests.get(SESSIONIZE_API_URL, timeout=8)
                if resp.status_code == 200:
                    remote_sessionize = resp.json().get("sessions", [])
            except Exception as e:
                console.print(f"[yellow]Warning: Could not fetch live Sessionize API ({e})[/yellow]")

    rows = []
    today = datetime.date.today().isoformat()

    # 2. Build entries
    # Local entries
    for talk in talks:
        talk_title = talk.get("title", "")
        summary = (talk.get("summary") or "").strip().replace("\r", " ").replace("\n", " ")
        talk_slides = normalize_url(talk.get("slides"))

        # Check if local talk has remote counterparts
        has_remote_deck = talk_slides in {normalize_url(d.get("link")) for d in remote_speakerdeck} or slugify(talk_title) in {slugify(d.get("title", "")) for d in remote_speakerdeck}
        has_remote_sess = slugify(talk_title) in {slugify(s.get("title", "")) for s in remote_sessionize}

        e_date = talk.get("date") or "-"
        e_name = talk.get("event") or "-"
        e_loc = talk.get("location") or "-"

        rows.append({
            "date": e_date,
            "event": e_name,
            "location": e_loc,
            "title": talk_title,
            "description": summary,
            "data_source": "local",
            "status": "ok",
            "has_remote_deck": has_remote_deck,
            "has_remote_sess": has_remote_sess,
            "slug": slugify(f"{e_name}-{talk_title}" if e_name != "-" else talk_title),
        })

    local_talks_by_slide = {normalize_url(t.get("slides")): t for t in talks if t.get("slides")}
    local_talks_by_slug = {slugify(t.get("title", "")): t for t in talks if t.get("title")}
    api_key = os.environ.get("GEMINI_API_KEY")
    client = genai.Client(api_key=api_key) if api_key else None

    # Speaker Deck entries
    for deck in remote_speakerdeck:
        d_url = normalize_url(deck.get("link", ""))
        raw_title = deck.get("title", "")
        raw_desc = (deck.get("description") or deck.get("summary") or "").strip().replace("\r", " ").replace("\n", " ")
        pub_date = deck.get("published", "")

        # Skip pre-2024 items
        if "published_parsed" in deck and deck["published_parsed"] and deck["published_parsed"].tm_year < 2024:
            continue

        # Check if already matched locally
        matched_local = local_talks_by_slide.get(d_url) or local_talks_by_slug.get(slugify(re.sub(r"^\[.*?\]\s*", "", raw_title)))
        if matched_local:
            d_title = matched_local.get("title", raw_title)
            d_event = matched_local.get("event", "-") or "-"
            d_loc = matched_local.get("location", "-") or "-"
            d_date = matched_local.get("date", pub_date[:10] if pub_date else today) or "-"
            d_desc = matched_local.get("summary", raw_desc)
            d_slug = slugify(f"{d_event}-{d_title}" if d_event != "-" else d_title)
            status = "ok"
        else:
            # Genuinely new deck: use Gemini LLM extraction
            if client:
                extracted = extract_speakerdeck_llm(client, raw_title, raw_desc, pub_date)
                d_title = extracted.title
                d_event = extracted.event_name or "-"
                d_loc = extracted.location or "-"
                d_date = extracted.event_date or (pub_date[:10] if pub_date else today)
                d_desc = extracted.summary
                d_slug = slugify(extracted.title)
            else:
                d_title = re.sub(r"^\[.*?\]\s*", "", raw_title).strip()
                d_event = "-"
                d_loc = "-"
                d_date = pub_date[:10] if pub_date else today
                d_desc = raw_desc
                d_slug = slugify(d_title)
            status = "needs sync"

        rows.append({
            "date": d_date,
            "event": d_event,
            "location": d_loc,
            "title": d_title,
            "description": d_desc,
            "data_source": "speakerdeck",
            "status": status,
            "slug": d_slug,
        })

    # Sessionize entries
    for sess in remote_sessionize:
        s_title = sess.get("title", "")
        s_slug = slugify(s_title)
        s_desc = (sess.get("description") or "").strip().replace("\r", " ").replace("\n", " ")

        # Check if synced locally
        is_synced = s_slug in local_titles
        status = "ok" if is_synced else "needs sync"

        rows.append({
            "date": "-",
            "event": "Sessionize",
            "location": "-",
            "title": s_title,
            "description": s_desc,
            "data_source": "sessionize",
            "status": status,
            "slug": s_slug,
        })

    # 3. Deduplicate if --dedup is passed
    if getattr(args, "dedup", False):
        dedup_map = {}
        for r in rows:
            key = r["slug"]
            if key not in dedup_map:
                dedup_map[key] = {
                    "date": r["date"],
                    "event": r["event"],
                    "location": r["location"],
                    "title": r["title"],
                    "description": r["description"],
                    "sources": {r["data_source"]},
                    "status": r["status"],
                }
            else:
                existing = dedup_map[key]
                existing["sources"].add(r["data_source"])
                # Prefer named event over placeholder/unspecified
                if existing["event"] in ("-", "Speaker Deck", "Sessionize", "Unspecified") and r["event"] not in ("-", "Speaker Deck", "Sessionize", "Unspecified"):
                    existing["event"] = r["event"]
                # Prefer known location over placeholder
                if existing["location"] in ("-", "TBD") and r["location"] not in ("-", "TBD"):
                    existing["location"] = r["location"]
                # Prefer earlier/actual presentation date over upload date
                if existing["date"] == "-" or (r["date"] != "-" and r["data_source"] == "local"):
                    existing["date"] = r["date"]
                if "local" in existing["sources"]:
                    existing["status"] = "ok"

        deduped_rows = []
        for v in dedup_map.values():
            deduped_rows.append({
                "date": v["date"],
                "event": v["event"],
                "location": v["location"],
                "title": v["title"],
                "description": v["description"],
                "data_source": ", ".join(sorted(v["sources"])),
                "status": v["status"],
            })
        display_rows = deduped_rows
    else:
        display_rows = rows

    # Sort descending by date
    display_rows.sort(key=lambda x: x["date"] if x["date"] != "-" else "1970-01-01", reverse=True)

    table = Table(
        title=f"\n[bold cyan]Events & Presentations[/bold cyan] [dim]({len(display_rows)} entries)[/dim]\n",
        header_style="bold magenta",
        box=box.ROUNDED,
    )
    table.add_column("Date", style="cyan", no_wrap=True, min_width=11)
    table.add_column("Event", style="magenta", max_width=20, overflow="ellipsis")
    table.add_column("Location", style="dim", max_width=13, overflow="ellipsis")
    table.add_column("Title", style="bold green", max_width=26, overflow="ellipsis")
    table.add_column("Description", style="white", max_width=32, overflow="ellipsis")
    table.add_column("Source", style="blue", no_wrap=True, min_width=18)
    table.add_column("Status", no_wrap=True, min_width=8)

    for r in display_rows:
        status_styled = "[bold green]ok[/bold green]" if r["status"] == "ok" else "[bold yellow]needs sync[/bold yellow]"
        clean_desc = (r["description"] or "").strip().replace("\n", " ").replace("\r", " ")
        if len(clean_desc) > 65:
            clean_desc = clean_desc[:62].rstrip() + "..."

        table.add_row(
            r["date"],
            r["event"],
            r["location"],
            r["title"],
            clean_desc or "-",
            r["data_source"],
            status_styled,
        )

    console.print(table)
    console.print()


def cmd_validate(args):
    """Validate talks.json structure, taxonomy, and deterministic IDs."""
    data = load_talks()
    talks = data.get("talks", [])
    errors = 0
    warnings = 0

    console.print(f"[bold]Validating {len(talks)} presentations in talks.json...[/bold]\n")

    for i, talk in enumerate(talks):
        title = talk.get("title", f"Presentation #{i}")
        categories = talk.get("categories", [])
        if len(categories) != 1:
            console.print(f"[red]Error in '{title}':[/red] Each presentation must be associated with exactly one category, found {categories}")
            errors += 1
        for cat in categories:
            if cat not in VALID_CATEGORIES:
                console.print(f"[red]Error in '{title}':[/red] Invalid category '{cat}' (must be one of {VALID_CATEGORIES})")
                errors += 1

        ev_name = (talk.get("event") or "").strip()
        if ev_name:
            if ev_name.lower() in BLACKLISTED_EVENT_NAMES or "speaker deck" in ev_name.lower():
                console.print(f"[red]Error in '{title}':[/red] Invalid generic event name '{ev_name}' (Event cannot be named after hosting platforms or generic filler)")
                errors += 1

        d = talk.get("date")
        if d:
            try:
                datetime.date.fromisoformat(d)
            except Exception:
                console.print(f"[red]Error in '{title}':[/red] Invalid date format '{d}' (expected YYYY-MM-DD)")
                errors += 1

    if errors == 0 and warnings == 0:
        console.print("[green]✓ All presentations are strictly valid![/green]")
    else:
        console.print(f"\nCompleted validation: [red]{errors} error(s)[/red], [yellow]{warnings} warning(s)[/yellow]")


def cmd_add(args):
    """Interactive wizard to add a talk or workshop with LLM Judge."""
    api_key = os.environ.get("GEMINI_API_KEY")
    data = load_talks()

    console.print("\n[bold cyan]=== Add New Event / Talk Wizard ===[/bold cyan]\n")
    input_text = Prompt.ask("Paste URL (Google Slides, YouTube, Sessionize) or brief description", default="")

    if input_text and api_key:
        console.print(f"\n[dim]Evaluating input with {MODEL_NAME} Judge...[/dim]")
        client = genai.Client(api_key=api_key)
        verdict = evaluate_and_heal_record(
            client=client,
            raw_title=input_text,
            raw_description=input_text,
            pub_date=datetime.date.today().isoformat(),
            slide_url=input_text if "speakerdeck" in input_text or "docs.google.com" in input_text else None,
            thumbnail=None,
            existing_talks=data.get("talks", []),
        )
        rec = verdict.final_record.model_dump()
    else:
        title = Prompt.ask("Talk Title")
        summary = Prompt.ask("Summary")
        event_name = Prompt.ask("Event Name")
        event_date = Prompt.ask("Event Date (YYYY-MM-DD)", default=datetime.date.today().isoformat())
        event_loc = Prompt.ask("Location", default="London, UK")
        slides = Prompt.ask("Slides URL", default="")
        recording = Prompt.ask("Recording URL", default="")
        tags_str = Prompt.ask("Tags (comma-separated)", default="golang, ai")
        tags = [t.strip().lower() for t in tags_str.split(",") if t.strip()]

        rec = {
            "id": generate_deterministic_id(event_date, event_name, title),
            "title": title,
            "summary": summary,
            "image": None,
            "events": [{"name": event_name, "date": event_date, "location": event_loc, "url": None}],
            "slides": slides or None,
            "source_code": None,
            "recording": recording or None,
            "tags": tags,
            "categories": ["AI & Development"],
        }

    console.print("\n[bold]Validated Entry:[/bold]")
    console.print(json.dumps(rec, indent=2, ensure_ascii=False))

    if Confirm.ask("\nAdd this talk to talks.json?", default=True):
        data["talks"].insert(0, rec)
        save_talks(data)
        console.print(f"[green]✓ Successfully added '{rec['title']}' to talks.json![/green]")


def configure_import_parser(parser: argparse.ArgumentParser):
    parser.add_argument("--speakerdeck", action="store_true", help="Import only from Speaker Deck")
    parser.add_argument("--sessionize", action="store_true", help="Import only from Sessionize")
    parser.set_defaults(func=cmd_import)


def main():
    parser = argparse.ArgumentParser(description="Events Management CLI for danicat.dev")
    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    import_parser = subparsers.add_parser("import", help="Import/sync talks from Speaker Deck and Sessionize (no flags = all)")
    configure_import_parser(import_parser)

    sync_parser = subparsers.add_parser("sync", help="Alias for 'import'")
    configure_import_parser(sync_parser)

    add_parser = subparsers.add_parser("add", help="Add a new talk or workshop interactively")
    add_parser.set_defaults(func=cmd_add)

    list_parser = subparsers.add_parser("list", help="List all events across sources (date | event | location | title | description | data source | status)")
    list_parser.add_argument("--dedup", action="store_true", help="Deduplicate identical events across sources into single combined rows")
    list_parser.add_argument("--local", action="store_true", help="Offline mode: list only local talks.json without fetching remote sources")
    list_parser.set_defaults(func=cmd_list)

    validate_parser = subparsers.add_parser("validate", help="Validate talks.json structure and taxonomy")
    validate_parser.set_defaults(func=cmd_validate)

    args = parser.parse_args()
    if hasattr(args, "func"):
        args.func(args)
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
