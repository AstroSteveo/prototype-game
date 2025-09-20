#!/usr/bin/env python3
import os
import re
import json
import tempfile
import subprocess
from pathlib import Path

REPO = os.environ.get("REPO", "AstroSteveo/prototype-game")
PROJECT_OWNER = os.environ.get("PROJECT_OWNER", "AstroSteveo")
PROJECT_NUMBER = os.environ.get("PROJECT_NUMBER", "2")
ASSIGNEE = os.environ.get("ASSIGNEE", "")
LABELS_DEFAULT = os.environ.get("LABELS", "task").split(",")
DRY_RUN = os.environ.get("DRY_RUN", "false").lower() in ("1","true","yes","on")

ROOT = Path(__file__).resolve().parents[1]
TASKS_FILE = ROOT / "tasks.md"
OUT_DIR = ROOT / ".agent_work" / "issues"


def sh(cmd:list, check=True, capture=True, env=None):
    if capture:
        res = subprocess.run(cmd, check=check, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, env=env)
        return res.stdout.strip()
    else:
        subprocess.run(cmd, check=check, env=env)
        return ""


def gh_available():
    try:
        sh(["gh", "--version"])  # just to see if it runs
        auth = sh(["gh", "auth", "status", "-h", "github.com"]) 
        return True
    except Exception:
        return False


task_re = re.compile(r"^\s*\d+\.\s*(T-\d{3})\s*[—-]\s*(.+?)\s*$")
phase_re = re.compile(r"^##\s*(Phase\s+\d+\s*[—-]\s*.+)")


def parse_tasks(md_text:str):
    lines = md_text.splitlines()
    tasks = []
    cur_phase = None
    i = 0
    while i < len(lines):
        line = lines[i]
        m_phase = phase_re.match(line)
        if m_phase:
            cur_phase = m_phase.group(1)
            i += 1
            continue
        m_task = task_re.match(line)
        if m_task:
            tid = m_task.group(1)
            title = m_task.group(2)
            # collect subsequent bullet block until blank line that precedes next numbered item or phase
            desc = []
            i += 1
            while i < len(lines):
                ln = lines[i]
                if phase_re.match(ln) or task_re.match(ln):
                    break
                desc.append(ln)
                i += 1
            block = "\n".join(desc)
            data = extract_fields(block)
            tasks.append({
                "id": tid,
                "title": title.strip(),
                "phase": cur_phase or "",
                **data,
            })
            continue
        i += 1
    return tasks


def extract_fields(block:str):
    # lines like: "   - Description: ..."
    fields = {
        "description": "",
        "requirements": "",
        "dependencies": "",
        "estimate": "",
        "priority": "",
        "acceptance": [],
    }
    acc = []
    for ln in block.splitlines():
        s = ln.strip()
        if s.startswith("- Description:"):
            fields["description"] = s.split(":",1)[1].strip()
        elif s.startswith("- Related Requirements:"):
            fields["requirements"] = s.split(":",1)[1].strip()
        elif s.startswith("- Dependencies:"):
            fields["dependencies"] = s.split(":",1)[1].strip()
        elif s.startswith("- Estimate:"):
            fields["estimate"] = s.split(":",1)[1].strip()
        elif s.startswith("- Priority:"):
            fields["priority"] = s.split(":",1)[1].strip()
        elif s.startswith("- Acceptance Criteria:"):
            crit = s.split(":",1)[1].strip()
            if crit:
                acc.append(crit)
        elif s.startswith("- ") and "Acceptance Criteria" not in s:
            # treat as potential acceptance criteria continuation if we are inside acceptance block
            pass
    # Also scan for additional indented acceptance bullets following the Acceptance Criteria line
    # Using a simple heuristic: lines that start with "- " after the Acceptance Criteria declaration in block
    after_acc = False
    for ln in block.splitlines():
        s = ln.strip()
        if s.startswith("- Acceptance Criteria:"):
            after_acc = True
            continue
        if after_acc:
            if s.startswith("-"):
                item = s[1:].strip()
                if item:
                    acc.append(item)
            elif s == "":
                break
    fields["acceptance"] = [a for a in [x.strip("- ") for x in acc] if a]
    return fields


def ensure_out_dir():
    OUT_DIR.mkdir(parents=True, exist_ok=True)


def issue_exists(tid:str, title:str):
    try:
        data = sh(["gh", "issue", "list", "--state", "all", "--search", tid, "--json", "number,title,htmlURL"]) or "[]"
        items = json.loads(data)
        for it in items:
            if tid in it.get("title", ""):
                return it
        return None
    except Exception:
        return None


def create_issue(repo:str, title:str, body:str, labels:list, assignee:str=""):
    with tempfile.NamedTemporaryFile("w", delete=False) as tf:
        tf.write(body)
        tmp = tf.name
    cmd = ["gh", "issue", "create", "--repo", repo, "--title", title, "--body-file", tmp]
    if labels:
        for lb in labels:
            lb = lb.strip()
            if lb:
                cmd += ["--label", lb]
    if assignee:
        cmd += ["--assignee", assignee]
    # capture URL
    out = sh(cmd)
    os.unlink(tmp)
    # gh prints URL on success
    url = out.strip().splitlines()[-1].strip() if out else ""
    return url


def add_to_project(owner:str, number:str, url:str):
    try:
        sh(["gh", "project", "item-add", "--owner", owner, "--number", str(number), "--url", url])
        return True
    except Exception as e:
        return False


def ensure_label(name:str, color:str="8b949e", description:str=""):
    try:
        # Try to view; if not present, create
        sh(["gh", "label", "view", name, "--repo", REPO])
        return True
    except Exception:
        try:
            sh(["gh", "label", "create", name, "--color", color, "--description", description, "--repo", REPO])
            return True
        except Exception:
            return False


def build_issue_body(t:dict):
    lines = []
    lines.append(f"# [{t['id']}] {t['title']} ({t.get('phase','')})")
    lines.append("")
    lines.append(f"- Summary: {t['description']}")
    if t.get("requirements"):
        lines.append(f"- Related Requirements: {t['requirements']}")
    if t.get("dependencies"):
        lines.append(f"- Dependencies: {t['dependencies']}")
    if t.get("estimate"):
        lines.append(f"- Estimate: {t['estimate']}")
    if t.get("priority"):
        lines.append(f"- Priority: {t['priority']}")
    lines.append("")
    if t.get("acceptance"):
        lines.append("## Acceptance Criteria")
        for a in t["acceptance"]:
            # render as task list
            lines.append(f"- [ ] {a}")
        lines.append("")
    lines.append("References: `requirements.md`, `design.md`, `tasks.md`")
    return "\n".join(lines)


def priority_label(p:str):
    p = (p or "").lower()
    if p.startswith("high"):
        return "priority:high"
    if p.startswith("medium"):
        return "priority:medium"
    if p.startswith("low"):
        return "priority:low"
    return None


def phase_label(phase:str):
    m = re.match(r"Phase\s*(\d+)", phase or "")
    if m:
        return f"phase:{m.group(1)}"
    return None


def slugify(text:str):
    s = re.sub(r"[^a-zA-Z0-9]+", "-", text).strip("-")
    return s.lower()


def main():
    if not TASKS_FILE.exists():
        raise SystemExit(f"tasks.md not found at {TASKS_FILE}")
    if not DRY_RUN and not gh_available():
        raise SystemExit("GitHub CLI (gh) not available or not authenticated. Install gh and run `gh auth login`.")

    ensure_out_dir()
    tasks = parse_tasks(TASKS_FILE.read_text())
    # Ensure labels exist
    base_labels = set(LABELS_DEFAULT)
    prios = {"priority:high": "High priority", "priority:medium": "Medium priority", "priority:low": "Low priority"}
    phases = {f"phase:{i}": f"Phase {i} tasks" for i in range(1, 10)}
    needed = set(base_labels) | set(prios.keys()) | set(phases.keys())
    if not DRY_RUN:
        for lb in sorted(needed):
            desc = prios.get(lb, phases.get(lb, ""))
            ensure_label(lb, description=desc)
    results = []
    for t in tasks:
        title = f"[{t['id']}] {t['title']}"
        body = build_issue_body(t)
        # Always write body to file for traceability
        fname = f"{t['id']}_{slugify(t['title'])}.md"
        (OUT_DIR / fname).write_text(body)
        labels = LABELS_DEFAULT.copy()
        pl = priority_label(t.get("priority",""))
        if pl:
            labels.append(pl)
        ph = phase_label(t.get("phase",""))
        if ph:
            labels.append(ph)

        existing = issue_exists(t['id'], title) if not DRY_RUN else None
        if DRY_RUN:
            url = existing.get("htmlURL") if existing else "(dry-run)"
            created = False if existing else True
            added = False
        else:
            if existing:
                url = existing.get("htmlURL")
                created = False
            else:
                url = create_issue(REPO, title, body, labels, ASSIGNEE)
                created = True
            added = False
            if url and not url.startswith("(dry-run)"):
                added = add_to_project(PROJECT_OWNER, PROJECT_NUMBER, url)
        results.append({
            "id": t['id'],
            "title": title,
            "url": url,
            "created": created,
            "project_added": added,
            "labels": labels,
        })

    index_path = OUT_DIR / "index.json"
    index_path.write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
