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
import sys
from pathlib import Path
from typing import Dict, List, Literal, Optional, Set, Tuple

import feedparser
import requests
from google import genai
from google.genai import types
from pydantic import BaseModel, Field, model_validator
from rich import box
from rich.console import Console
from rich.prompt import Confirm, Prompt
from rich.table import Table

console = Console(width=130)

SPEAKER_CANONICAL_NAME = "Daniela Petruzalek"
SPEAKERDECK_RSS_URL = "https://speakerdeck.com/danicat.rss"
SESSIONIZE_API_URL = os.environ.get("SESSIONIZE_API_URL")
TALKS_JSON_PATH = Path(__file__).resolve().parent.parent / "data" / "talks.json"
MODEL_NAME = os.environ.get("GEMINI_MODEL", "gemini-2.5-flash")

VALID_CATEGORIES = [
    "Agentic Coding",
    "Agent Development",
    "Applied GenAI",
    "Perspectives",
    "Software Engineering",
]


class EventRecord(BaseModel):
    id: str = Field(description="Deterministic slug: YYYY-MM-DD-event-name-talk-title or talk-title if undated")
    title: str = Field(description="Clean canonical talk title")
    event: Optional[str] = Field(default=None, description="Verified conference, meetup, or event name (null if none)")
    date: Optional[str] = Field(default=None, description="Verified date of presentation in YYYY-MM-DD format (null if unverified)")
    location: Optional[str] = Field(default=None, description="Verified City, Country or 'Online' (null if unverified)")
    url: Optional[str] = Field(default=None, description="Official event schedule/website URL (null if none)")
    summary: str = Field(description="Concise 1-3 sentence summary of the talk")
    image: Optional[str] = Field(default=None, description="Custom thumbnail or slide preview URL")
    slides: Optional[str] = Field(default=None, description="Slides URL")
    source_code: Optional[str] = Field(default=None, description="GitHub or source code repository URL")
    recording: Optional[str] = Field(default=None, description="Recording or YouTube URL")
    codelab: Optional[str] = Field(default=None, description="Hands-on codelab or workshop tutorial URL")
    materials: Optional[str] = Field(default=None, description="Supporting materials URL")
    language: Optional[str] = Field(default="en", description="Delivery language code, e.g. 'en', 'pt-BR', 'ja'")
    tags: List[str] = Field(description="Lowercase kebab-case relevant tags, strictly sorted in alphabetical order, must never repeat category name")
    categories: List[str] = Field(description="Strictly ONE valid taxonomy category from VALID_CATEGORIES")

    @model_validator(mode="after")
    def clean_and_sort_tags(self) -> "EventRecord":
        cat_slugs = {slugify(c) for c in self.categories}
        cleaned = [
            t.strip().lower() for t in self.tags
            if t.strip() and slugify(t) not in cat_slugs
        ]
        self.tags = sorted(list(dict.fromkeys(cleaned)))
        return self


class JudgeVerdict(BaseModel):
    decision: Literal["accept", "auto_corrected", "merge"] = Field(
        description="'accept' if pristine; 'auto_corrected' if normalized; 'merge' if it enriches an existing record"
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


def slugify(text: str) -> str:
    """Generate a clean URL/ID slug from text."""
    text = text.lower()
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[\s_-]+", "-", text).strip("-")
    return text


def generate_deterministic_id(date_str: Optional[str], event_name: Optional[str], talk_title: str) -> str:
    """Generate deterministic event ID: YYYY-MM-DD-event-talk or talk-title if undated."""
    parts = []
    if date_str:
        parts.append(slugify(date_str))
    if event_name:
        parts.append(slugify(event_name))
    parts.append(slugify(talk_title))
    return "-".join(parts)


def normalize_url(url: Optional[str]) -> str:
    """Normalize URL by stripping query parameters and trailing slashes."""
    if not url:
        return ""
    return url.split("?")[0].rstrip("/")


def fetch_speakerdeck_preview(slide_url: Optional[str]) -> Optional[str]:
    """Fetch presentation preview image from Speaker Deck HTML/player."""
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


def research_event_with_google_search(
    client: genai.Client,
    talk_title: str,
    raw_description: str = "",
    event_hint: str = "",
) -> str:
    """Stage 1: Use Gemini with Google Search tool to find verified conference and presentation details for Daniela Petruzalek."""
    prompt = f"""
Search Google to find verified conference, meetup, or presentation event details for speaker "{SPEAKER_CANONICAL_NAME}".

Target Presentation Title: "{talk_title}"
Event Hint (if any): "{event_hint}"
Description Snippet: "{raw_description}"

Query instructions:
Search for "{SPEAKER_CANONICAL_NAME}" together with "{talk_title}" and any event hints.
Find and report:
1. Conference / Meetup / Event Name (e.g., 'DevFest Pisa 2026', 'GoLab 2025', 'GopherCon UK 2025', 'FOSDEM 2026', 'TDC São Paulo 2025')
2. Date of presentation (YYYY-MM-DD format). If only month and year are available, note it.
3. Location (City, Country or 'Online')
4. Official Schedule / Event URL

If this talk has not been presented at an actual event or you cannot find verified conference details, state: NO_VERIFIED_EVENT_FOUND.
"""
    try:
        response = client.models.generate_content(
            model=MODEL_NAME,
            contents=prompt,
            config=types.GenerateContentConfig(
                tools=[types.Tool(google_search=types.GoogleSearch())],
                temperature=0.0,
            ),
        )
        return response.text or ""
    except Exception as e:
        console.print(f"[yellow]Google Search grounding note: {e}[/yellow]")
        return ""


def evaluate_and_heal_record(
    client: genai.Client,
    raw_title: str,
    raw_description: str,
    slide_url: Optional[str],
    thumbnail: Optional[str],
    existing_talks: List[dict],
    event_hint: str = "",
) -> JudgeVerdict:
    """Stage 2: Structured Output normalization using Gemini native response_schema."""
    bracket_match = re.match(r"^\[(.*?)\]\s*(.*)", raw_title)
    if bracket_match:
        if not event_hint:
            event_hint = bracket_match.group(1).strip()
        clean_title_cand = bracket_match.group(2).strip()
    else:
        clean_title_cand = raw_title.strip()

    # Stage 1: Google Search Grounding for Daniela Petruzalek
    grounded_research = research_event_with_google_search(
        client=client,
        talk_title=clean_title_cand,
        raw_description=raw_description,
        event_hint=event_hint,
    )

    existing_summaries = [
        {"id": t.get("id") or t.get("title"), "title": t.get("title"), "event": t.get("event"), "date": t.get("date")}
        for t in existing_talks[:25]
    ]

    prompt = f"""
You are an expert technical editor for {SPEAKER_CANONICAL_NAME}'s developer portfolio.
Structure and normalize this presentation entry using the verified web research findings and input data.

Input Presentation:
- Raw Title: {raw_title}
- Event Hint: {event_hint}
- Description: {raw_description}
- Slide URL: {slide_url}

Verified Web Research Findings (from Google Search for {SPEAKER_CANONICAL_NAME}):
{grounded_research if grounded_research else "NO_VERIFIED_EVENT_FOUND"}

Existing Recent Talks in Portfolio:
{json.dumps(existing_summaries, indent=2)}

Allowed Taxonomy Categories (Must choose EXACTLY ONE strictly from this list):
{json.dumps(VALID_CATEGORIES)}

Factuality & Structuring Rules:
1. Title: Clean canonical title (strip any bracketed event prefix like '[DevFest Pisa 2026]').
2. Event & Presentation Date:
   - Use the verified conference name, date (YYYY-MM-DD), location, and event URL discovered in the Web Research Findings or explicitly written in the notes.
   - If Web Research Findings say NO_VERIFIED_EVENT_FOUND or no event was verified, set event = null, date = null, location = null, url = null.
   - NEVER use the slide deck upload date as an event date.
   - NEVER use hosting platforms like Speaker Deck or Sessionize as event names.
3. Summary:
   - Provide a concise 1-3 sentence summary. Prioritize the detailed description provided.
4. Slug ID:
   - If date and event are known: deterministic id = "YYYY-MM-DD-eventname-talktitle" (slugified).
   - If undated: id = "talktitle" (slugified).
5. Category: Assign EXACTLY ONE valid category from Allowed Taxonomy Categories.
6. Tags: Lowercase, kebab-case (e.g. 'go', 'gemini-api', 'mcp', 'agentic-coding').
7. Deduplication / Merge:
   - Check if this matches an existing talk in the portfolio. If so, set decision = 'merge' and target_id to the existing talk's id.
   - Otherwise, set decision = 'accept' or 'auto_corrected'.
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
        console.print(f"[yellow]Structured judge fallback ({e}). Applying rule-based normalization.[/yellow]")
        clean_title = clean_title_cand or raw_title
        safe_id = generate_deterministic_id(None, None, clean_title)

        fallback_record = EventRecord(
            id=safe_id,
            title=clean_title,
            event=None,
            date=None,
            location=None,
            url=None,
            summary=raw_description.split("\n")[0] if raw_description else clean_title,
            image=thumbnail,
            slides=slide_url,
            source_code=None,
            recording=None,
            codelab=None,
            materials=None,
            language="en",
            tags=["tech", "software-engineering"],
            categories=["Software Engineering"],
        )
        return JudgeVerdict(
            decision="auto_corrected",
            target_id=None,
            auto_fixes_applied=[f"Applied fallback normalization: {e}"],
            final_record=fallback_record,
        )


def load_talks() -> dict:
    if not TALKS_JSON_PATH.exists():
        console.print(f"[red]ERROR:[/red] Talks data file not found at {TALKS_JSON_PATH}")
        sys.exit(1)
    with open(TALKS_JSON_PATH, "r", encoding="utf-8") as f:
        return json.load(f)


def save_talks(data: dict):
    # Ensure tags for all talks are strictly sorted alphabetically, deduplicated, and do not repeat category names
    for t in data.get("talks", []):
        if "tags" in t and isinstance(t["tags"], list):
            cat_slugs = {slugify(c) for c in t.get("categories", [])}
            t["tags"] = sorted(list(dict.fromkeys(
                tag.strip().lower() for tag in t["tags"]
                if tag.strip() and slugify(tag) not in cat_slugs
            )))
    with open(TALKS_JSON_PATH, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        f.write("\n")


def fetch_remote_sources(sessionize_url: Optional[str] = None) -> Tuple[List[dict], List[dict], Optional[dict]]:
    """Fetch live Speaker Deck RSS entries and Sessionize profile/sessions."""
    remote_speakerdeck = []
    remote_sessionize = []
    sessionize_speaker = None

    # 1. Fetch Speaker Deck
    try:
        feed = feedparser.parse(SPEAKERDECK_RSS_URL)
        if feed.entries:
            remote_speakerdeck = feed.entries
    except Exception as e:
        console.print(f"[yellow]Warning: Could not fetch Speaker Deck feed ({e})[/yellow]")

    # 2. Fetch Sessionize
    sess_endpoint = sessionize_url or SESSIONIZE_API_URL
    if sess_endpoint:
        try:
            resp = requests.get(sess_endpoint, timeout=10)
            if resp.status_code == 200:
                sess_json = resp.json()
                remote_sessionize = sess_json.get("sessions", [])
                sessionize_speaker = sess_json.get("speaker")
        except Exception as e:
            console.print(f"[yellow]Warning: Could not fetch Sessionize API ({e})[/yellow]")

    return remote_speakerdeck, remote_sessionize, sessionize_speaker


def process_discovery_pipeline(
    data: dict,
    api_key: Optional[str],
    local_only: bool = False,
    dry_run: bool = False,
) -> Tuple[List[dict], int, int]:
    """
    Unified discovery and sync engine:
    1. Cross-references Speaker Deck (slides/previews) and Sessionize (descriptions).
    2. Identifies matching local entries vs new candidates.
    3. Runs Stage 1 (Google Search Grounding for Daniela Petruzalek) & Stage 2 (Gemini Structured Output) for new items.
    4. Detects field-level diffs (+slides, +image, +date, +event) for local entries.
    5. Updates talks.json if dry_run is False.
    """
    talks = data.get("talks", [])
    client = genai.Client(api_key=api_key) if api_key else None

    if local_only:
        # Simple offline view
        display_rows = []
        for t in talks:
            display_rows.append({
                "date": t.get("date") or "-",
                "event": t.get("event") or "-",
                "location": t.get("location") or "-",
                "title": t.get("title", ""),
                "description": t.get("summary", ""),
                "data_source": "local",
                "status": "ok",
                "diff_detail": "",
                "record": t,
            })
        return display_rows, 0, 0

    remote_decks, remote_sessions, sess_speaker = fetch_remote_sources()

    # 1. Index local talks
    local_by_slide = {normalize_url(t.get("slides")): t for t in talks if t.get("slides")}
    local_by_slug = {slugify(t.get("title", "")): t for t in talks if t.get("title")}

    # 2. Combine remote sources: Precedence is Sessionize description over Speaker Deck description
    # Group remote entries by normalized title slug
    remote_candidates: Dict[str, dict] = {}

    # A. Index Sessionize sessions first (rich descriptions)
    for sess in remote_sessions:
        s_title = sess.get("title", "").strip()
        s_slug = slugify(s_title)
        s_desc = (sess.get("description") or "").strip()
        remote_candidates[s_slug] = {
            "title": s_title,
            "description": s_desc,
            "slides": None,
            "image": None,
            "sources": {"sessionize"},
            "raw_deck": None,
            "raw_sess": sess,
        }

    # B. Index Speaker Deck decks (slides & preview images)
    for deck in remote_decks:
        d_url = normalize_url(deck.get("link", ""))
        raw_title = deck.get("title", "").strip()
        raw_desc = (deck.get("description") or deck.get("summary") or "").strip()

        # Skip pre-2024 decks
        if "published_parsed" in deck and deck["published_parsed"] and deck["published_parsed"].tm_year < 2024:
            continue

        clean_d_title = re.sub(r"^\[.*?\]\s*", "", raw_title).strip()
        d_slug = slugify(clean_d_title)
        thumb = extract_thumbnail_url(deck)

        if d_slug in remote_candidates:
            # Match existing Sessionize session: attach slides & image, keep Sessionize description!
            cand = remote_candidates[d_slug]
            cand["slides"] = d_url
            cand["image"] = thumb
            cand["sources"].add("speakerdeck")
            cand["raw_deck"] = deck
            # Only use Speaker Deck description if Sessionize was empty
            if not cand["description"] and raw_desc:
                cand["description"] = raw_desc
        else:
            remote_candidates[d_slug] = {
                "title": clean_d_title,
                "description": raw_desc,
                "slides": d_url,
                "image": thumb,
                "sources": {"speakerdeck"},
                "raw_deck": deck,
                "raw_sess": None,
            }

    # 3. Match against local talks and process diffs/new discoveries
    processed_local_ids: Set[str] = set()
    display_rows: List[dict] = []
    new_added_count = 0
    updated_diff_count = 0

    for cand_slug, cand in remote_candidates.items():
        matched_local = None
        if cand.get("slides") and cand["slides"] in local_by_slide:
            matched_local = local_by_slide[cand["slides"]]
        elif cand_slug in local_by_slug:
            matched_local = local_by_slug[cand_slug]

        if matched_local:
            processed_local_ids.add(matched_local.get("id", matched_local.get("title")))
            sources = {"local"} | cand["sources"]

            # Diff detection
            diffs = []
            if cand.get("slides") and not matched_local.get("slides"):
                diffs.append("+slides")
                if not dry_run:
                    matched_local["slides"] = cand["slides"]
            if cand.get("image") and not matched_local.get("image"):
                diffs.append("+image")
                if not dry_run:
                    matched_local["image"] = cand["image"]

            status_str = f"diff: {', '.join(diffs)}" if diffs else "ok"
            if diffs:
                updated_diff_count += 1

            display_rows.append({
                "date": matched_local.get("date") or "-",
                "event": matched_local.get("event") or "-",
                "location": matched_local.get("location") or "-",
                "title": matched_local.get("title", cand["title"]),
                "description": matched_local.get("summary") or cand["description"],
                "data_source": ", ".join(sorted(sources)),
                "status": status_str,
                "diff_detail": ", ".join(diffs),
                "record": matched_local,
            })
        else:
            # Genuinely new talk discovered from remote sources!
            sources = cand["sources"]
            sources_label = ", ".join(sorted(sources))

            if client:
                console.print(f"[dim]Processing Stage 1 & Stage 2 for new presentation:[/dim] [bold]{cand['title']}[/bold]...")
                verdict = evaluate_and_heal_record(
                    client=client,
                    raw_title=cand["title"],
                    raw_description=cand["description"],
                    slide_url=cand.get("slides"),
                    thumbnail=cand.get("image"),
                    existing_talks=talks,
                )
                final_rec = verdict.final_record.model_dump()
            else:
                safe_id = generate_deterministic_id(None, None, cand["title"])
                final_rec = {
                    "id": safe_id,
                    "title": cand["title"],
                    "event": None,
                    "date": None,
                    "location": None,
                    "url": None,
                    "summary": cand["description"].split("\n")[0] if cand["description"] else cand["title"],
                    "image": cand.get("image"),
                    "slides": cand.get("slides"),
                    "source_code": None,
                    "recording": None,
                    "codelab": None,
                    "materials": None,
                    "language": "en",
                    "tags": ["tech"],
                    "categories": ["Software Engineering"],
                }

            if not dry_run:
                talks.insert(0, final_rec)
                new_added_count += 1

            display_rows.append({
                "date": final_rec.get("date") or "-",
                "event": final_rec.get("event") or "-",
                "location": final_rec.get("location") or "-",
                "title": final_rec.get("title", cand["title"]),
                "description": final_rec.get("summary", cand["description"]),
                "data_source": sources_label,
                "status": f"new ({sources_label})",
                "diff_detail": "new item",
                "record": final_rec,
            })

    # 4. Include remaining local talks that had no remote match
    for t in talks:
        t_id = t.get("id", t.get("title"))
        if t_id not in processed_local_ids:
            display_rows.append({
                "date": t.get("date") or "-",
                "event": t.get("event") or "-",
                "location": t.get("location") or "-",
                "title": t.get("title", ""),
                "description": t.get("summary", ""),
                "data_source": "local",
                "status": "ok",
                "diff_detail": "",
                "record": t,
            })

    # 5. If not dry_run and Sessionize speaker data is present, update speaker bio
    if not dry_run and sess_speaker:
        data["speaker"] = {
            "firstName": sess_speaker.get("firstName", ""),
            "lastName": sess_speaker.get("lastName", ""),
            "tagline": sess_speaker.get("tagline", ""),
            "bio": sess_speaker.get("bio", ""),
            "speakerProfileUrl": sess_speaker.get("speakerProfileUrl", "https://sessionize.com/daniela"),
            "photoUrl": sess_speaker.get("photoUrl", ""),
            "photoLargeUrl": sess_speaker.get("photoLargeUrl", ""),
        }

    return display_rows, new_added_count, updated_diff_count


def render_events_table(rows: List[dict], title_suffix: str = ""):
    """Render standardized table with dated talks first (descending) and undated proposals at bottom."""
    dated_rows = [r for r in rows if r["date"] != "-"]
    undated_rows = [r for r in rows if r["date"] == "-"]

    dated_rows.sort(key=lambda x: x["date"], reverse=True)
    undated_rows.sort(key=lambda x: x["title"].lower())

    sorted_rows = dated_rows + undated_rows

    table = Table(
        title=f"\n[bold cyan]Events & Presentations {title_suffix}[/bold cyan] [dim]({len(sorted_rows)} entries)[/dim]\n",
        header_style="bold magenta",
        box=box.ROUNDED,
    )
    table.add_column("Date", style="cyan", no_wrap=True, min_width=11)
    table.add_column("Event", style="magenta", max_width=20, overflow="ellipsis")
    table.add_column("Location", style="dim", max_width=13, overflow="ellipsis")
    table.add_column("Title", style="bold green", max_width=26, overflow="ellipsis")
    table.add_column("Description", style="white", max_width=32, overflow="ellipsis")
    table.add_column("Source", style="blue", no_wrap=True, min_width=16)
    table.add_column("Status / Diff", no_wrap=True, min_width=14)

    for r in sorted_rows:
        status_text = r["status"]
        if status_text == "ok":
            status_styled = "[bold green]ok[/bold green]"
        elif status_text.startswith("diff:"):
            status_styled = f"[bold cyan]{status_text}[/bold cyan]"
        elif status_text.startswith("new"):
            status_styled = f"[bold yellow]{status_text}[/bold yellow]"
        else:
            status_styled = f"[bold white]{status_text}[/bold white]"

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


def cmd_list(args):
    """
    Discovery mode: Runs the full Stage 1 & Stage 2 discovery and diff pipeline
    without modifying talks.json.
    """
    api_key = os.environ.get("GEMINI_API_KEY")
    data = load_talks()
    local_only = getattr(args, "local", False)

    console.print(f"\n[bold cyan]Running discovery pipeline...[/bold cyan]")
    rows, _, _ = process_discovery_pipeline(
        data=data,
        api_key=api_key,
        local_only=local_only,
        dry_run=True,
    )

    render_events_table(rows, title_suffix="[Discovery View]")


def cmd_import(args):
    """
    Sync mode: Runs the full Stage 1 & Stage 2 discovery pipeline and updates talks.json.
    """
    api_key = os.environ.get("GEMINI_API_KEY")
    if not api_key:
        console.print("[red]ERROR:[/red] GEMINI_API_KEY environment variable is not set.", file=sys.stderr)
        sys.exit(1)

    data = load_talks()
    console.print(f"\n[bold cyan]Running discovery & sync pipeline...[/bold cyan]")

    rows, added_count, diff_count = process_discovery_pipeline(
        data=data,
        api_key=api_key,
        local_only=False,
        dry_run=False,
    )

    # Backfill deterministic IDs if missing on existing items
    for t in data.get("talks", []):
        if "id" not in t or not t["id"]:
            t_date = t.get("date")
            t_name = t.get("event")
            t["id"] = generate_deterministic_id(t_date, t_name, t.get("title", "talk"))

    save_talks(data)
    render_events_table(rows, title_suffix="[Sync Result]")
    console.print(f"[bold green]✓ talks.json updated! Added {added_count} new item(s), enriched {diff_count} existing item(s).[/bold green]\n")


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

        d = talk.get("date")
        if d:
            try:
                datetime.date.fromisoformat(d)
            except Exception:
                console.print(f"[red]Error in '{title}':[/red] Invalid date format '{d}' (expected YYYY-MM-DD)")
                errors += 1

        t_id = talk.get("id")
        if not t_id:
            console.print(f"[red]Error in '{title}':[/red] Missing unique 'id' field")
            errors += 1

        tags = talk.get("tags", [])
        cat_slugs = {slugify(c) for c in categories}
        repeated_tags = [t for t in tags if slugify(t) in cat_slugs]
        if repeated_tags:
            console.print(f"[red]Error in '{title}':[/red] Tags repeat category name: {repeated_tags} for category {categories}")
            errors += 1

        sorted_tags = sorted(list(dict.fromkeys(t.strip().lower() for t in tags if t.strip() and slugify(t) not in cat_slugs)))
        if tags != sorted_tags:
            console.print(f"[red]Error in '{title}':[/red] Tags are not properly formatted/sorted: {tags} (expected {sorted_tags})")
            errors += 1

    if errors == 0 and warnings == 0:
        console.print("[green]✓ All presentations are strictly valid![/green]")
    else:
        console.print(f"\nCompleted validation: [red]{errors} error(s)[/red], [yellow]{warnings} warning(s)[/yellow]")


def cmd_add(args):
    """Interactive wizard to add a talk or workshop with Google Search Grounding and Structured Output Judge."""
    api_key = os.environ.get("GEMINI_API_KEY")
    data = load_talks()

    console.print("\n[bold cyan]=== Add New Event / Talk Wizard ===[/bold cyan]\n")
    input_text = Prompt.ask("Paste URL (Google Slides, YouTube, Sessionize) or brief description", default="")

    if input_text and api_key:
        console.print(f"\n[dim]Searching web for '{SPEAKER_CANONICAL_NAME}' and structuring with {MODEL_NAME}...[/dim]")
        client = genai.Client(api_key=api_key)
        verdict = evaluate_and_heal_record(
            client=client,
            raw_title=input_text,
            raw_description=input_text,
            slide_url=input_text if "speakerdeck" in input_text or "docs.google.com" in input_text else None,
            thumbnail=None,
            existing_talks=data.get("talks", []),
        )
        rec = verdict.final_record.model_dump()
    else:
        title = Prompt.ask("Talk Title")
        summary = Prompt.ask("Summary")
        event_name = Prompt.ask("Event Name", default="")
        event_date = Prompt.ask("Event Date (YYYY-MM-DD)", default="")
        event_loc = Prompt.ask("Location", default="")
        slides = Prompt.ask("Slides URL", default="")
        recording = Prompt.ask("Recording URL", default="")
        tags_str = Prompt.ask("Tags (comma-separated)", default="software-engineering")
        tags = [t.strip().lower() for t in tags_str.split(",") if t.strip()]

        rec = {
            "id": generate_deterministic_id(event_date if event_date else None, event_name if event_name else None, title),
            "title": title,
            "event": event_name or None,
            "date": event_date or None,
            "location": event_loc or None,
            "url": None,
            "summary": summary,
            "image": None,
            "slides": slides or None,
            "source_code": None,
            "recording": recording or None,
            "codelab": None,
            "materials": None,
            "language": "en",
            "tags": tags,
            "categories": ["Software Engineering"],
        }

    console.print("\n[bold]Validated Entry:[/bold]")
    console.print(json.dumps(rec, indent=2, ensure_ascii=False))

    if Confirm.ask("\nAdd this talk to talks.json?", default=True):
        data["talks"].insert(0, rec)
        save_talks(data)
        console.print(f"[green]✓ Successfully added '{rec['title']}' to talks.json![/green]")


def main():
    parser = argparse.ArgumentParser(description="Events Management CLI for danicat.dev")
    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    import_parser = subparsers.add_parser("import", help="Import/sync talks from Speaker Deck and Sessionize into talks.json")
    import_parser.set_defaults(func=cmd_import)

    sync_parser = subparsers.add_parser("sync", help="Alias for 'import'")
    sync_parser.set_defaults(func=cmd_import)

    add_parser = subparsers.add_parser("add", help="Add a new talk or workshop interactively")
    add_parser.set_defaults(func=cmd_add)

    list_parser = subparsers.add_parser("list", help="Discovery mode: list all events, highlight diffs, and discover new material without modifying talks.json")
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
