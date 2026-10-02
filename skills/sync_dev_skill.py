#!/usr/bin/env python3
"""Generate skills/kkal-tracker-food-log-dev from skills/kkal-tracker-food-log.

The dev skill is the production skill pointed at the development server. The agent runs elsewhere,
so the server is reached by the public address of the development machine.

Usage:  python3 skills/sync_dev_skill.py [http://HOST:PORT]
Edit the production skill only, then regenerate. The api_key file of the dev skill (git-ignored)
is left untouched; the API key in the repository is always the XXXXXXXXX placeholder.
"""

import pathlib
import sys

SKILLS_DIR = pathlib.Path(__file__).resolve().parent
PROJECT_DIR = SKILLS_DIR.parent
PROD = SKILLS_DIR / "kkal-tracker-food-log"
DEV = SKILLS_DIR / "kkal-tracker-food-log-dev"

PROD_URL = "https://kcal.peskov.info"
DEV_URL = "https://dev-kcal.peskov.info"

DEV_DESCRIPTION = (
    "DEV/TEST copy of the Kkal Tracker food logging skill. It writes to the development server, "
    "not to the real diary. Use it ONLY when the current message itself names the dev or test "
    'tracker — "запиши на дев", "в тестовый трекер", "на тестовый сервер", "log it to dev". The mention is '
    "required in every request — an earlier dev request in the same conversation does not carry over. "
    "The one exception is editing or deleting an entry that this dev copy created, which is always done here. "
    'For every other '
    '"I ate X" / "запиши еду" request use the regular kkal-tracker-food-log skill instead.'
)
DEV_BANNER = (
    "> **DEV skill.** Everything below goes to the development server, not to the production diary. "
    "Start every reply with \"[DEV]\" so the user can see where the meal was written.\n"
)


def replace_once(text, old, new, where):
    if text.count(old) != 1:
        sys.exit(f"{where}: expected exactly one occurrence of {old!r}, found {text.count(old)}")
    return text.replace(old, new)


def main():
    dev_url = sys.argv[1] if len(sys.argv) > 1 else DEV_URL
    if len(sys.argv) > 2 or not dev_url.startswith(("http://", "https://")):
        sys.exit(f"Usage: {sys.argv[0]} [http://HOST:PORT]")
    dev_url = dev_url.rstrip("/")

    (DEV / "scripts").mkdir(parents=True, exist_ok=True)

    skill_path = DEV / "SKILL.md"
    lines = (PROD / "SKILL.md").read_text(encoding="utf-8").split("\n")
    if lines[0] != "---" or not lines[1].startswith("name: ") or not lines[2].startswith("description: "):
        sys.exit("SKILL.md: unexpected frontmatter layout")
    lines[1] = "name: kkal-tracker-food-log-dev"
    lines[2] = "description: " + DEV_DESCRIPTION
    skill = "\n".join(lines)

    title = "# Kkal Tracker: food diary logging\n"
    skill = replace_once(skill, title, title.rstrip("\n") + " (DEV)\n\n" + DEV_BANNER, "SKILL.md")
    skill = replace_once(skill, f"- Base URL: `{PROD_URL}`", f"- Base URL: `{dev_url}`", "SKILL.md")
    skill_path.write_text(skill, encoding="utf-8")

    client_path = DEV / "scripts" / "kkal_client.py"
    client = (PROD / "scripts" / "kkal_client.py").read_text(encoding="utf-8")
    client = replace_once(client, f'DEFAULT_BASE_URL = "{PROD_URL}"', f'DEFAULT_BASE_URL = "{dev_url}"', "kkal_client.py")
    client_path.write_text(client, encoding="utf-8")

    print(f"Updated {DEV.relative_to(PROJECT_DIR)}")


if __name__ == "__main__":
    main()
