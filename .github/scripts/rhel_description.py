import os
import sys
from pathlib import Path

import markdown
import requests

MAX_DESCRIPTION_LEN = 32768
REQUEST_TIMEOUT = 30

##################

def fail(msg: str) -> None:
    print(f"::error::{msg}")
    sys.exit(1)


def get_env(name: str) -> str:
    value = os.environ.get(name, "").strip()
    if not value:
        fail(f"{name} environment variable is not specified")
    return value

##################

def load_description(file_name: str) -> str:
    html_path = Path(f"{file_name}.html")
    md_path = Path(f"{file_name}.md")

    if html_path.is_file():
        with html_path.open("r", encoding="utf-8") as f:
            return f.read()

    if md_path.is_file():
        with md_path.open("r", encoding="utf-8") as f:
            markdown_data = f.read()
        return markdown.markdown(markdown_data)

    fail(f"No description file found: expected '{html_path}' or '{md_path}'")

def main() -> int:
    description_file = get_env("DESCRIPTION_FILE")
    api_token = get_env("PYXIS_API_TOKEN")
    api_url = get_env("API_URL")
    project_id = get_env("PROJECT_ID")

    repository_description = load_description(description_file)

    if not repository_description.strip():
        fail("Description file is empty")

    was_truncated = len(repository_description) > MAX_DESCRIPTION_LEN
    repository_description = repository_description[:MAX_DESCRIPTION_LEN]

    payload = {
        "container": {
            "repository_description": repository_description,
        }
    }

    headers = {
        "accept": "application/json",
        "X-API-KEY": api_token,
        "Content-Type": "application/json",
    }

    url = f"{api_url}{project_id}"

    try:
        with requests.Session() as session:
            response = session.patch(
                url,
                headers = headers,
                json = payload,
                timeout = REQUEST_TIMEOUT,
            )
            response.raise_for_status()
    except requests.RequestException as exc:
        fail(f"Request to Pyxis API failed: {exc}")

    try:
        response_data = response.json()
    except ValueError:
        fail(f"API returned non-JSON response: {response.text[:500]}")

    print("::group::Result")
    print(f"Response code: {response.status_code}")
    if was_truncated:
        print(f"Warning: repository_description was truncated to {MAX_DESCRIPTION_LEN} characters")
    print(f"Last update date: {response_data.get('last_update_date', '<missing>')}")
    print("::endgroup::")

    return 0

if __name__ == "__main__":
    raise SystemExit(main())
