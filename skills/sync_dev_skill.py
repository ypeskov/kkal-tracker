#!/usr/bin/env python3
"""Generate skills/kkal-tracker-food-log-dev from skills/kkal-tracker-food-log.

The dev skill is the production skill pointed at a development server. The agent runs elsewhere,
so the server is reached by the public address of the development machine. That address is passed
as an argument and the generated skill is git-ignored: server addresses stay out of the repository.

Usage:  python3 skills/sync_dev_skill.py http://HOST:8080
Edit the production skill only, then regenerate. The api_key file of the dev skill is left untouched.
"""

import pathlib
import sys

SKILLS_DIR = pathlib.Path(__file__).resolve().parent
PROJECT_DIR = SKILLS_DIR.parent
PROD = SKILLS_DIR / "kkal-tracker-food-log"
DEV = SKILLS_DIR / "kkal-tracker-food-log-dev"

PROD_URL = "https://kcal.peskov.info"

DEV_DESCRIPTION = (
    "DEV/TEST copy of the Kkal Tracker food logging skill. It writes to the development server, "
    "not to the real diary. Use it ONLY when the user explicitly asks for the dev or test "
    'tracker — "запиши на дев", "в тестовый трекер", "на тестовый сервер", "log it to dev". For every ordinary '
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
    if len(sys.argv) != 2 or not sys.argv[1].startswith(("http://", "https://")):
        sys.exit(f"Usage: {sys.argv[0]} http://HOST:PORT")
    dev_url = sys.argv[1].rstrip("/")

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
