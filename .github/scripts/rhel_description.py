import sys
import requests
import json
import markdown
import os


def load_description(description_file):
    """Load repository description from .html or .md file.

    Returns the description string, or None if no suitable file is found.
    """
    if os.path.isfile(description_file + '.html'):
        with open(description_file + '.html', mode='r') as f:
            return f.read()
    if os.path.isfile(description_file + '.md'):
        with open(description_file + '.md', mode='r') as f:
            markdown_data = f.read()
        return markdown.markdown(markdown_data)
    return None


def validate_env():
    """Validate required environment variables.  Exits with code 1 on failure."""
    if not os.environ.get("DESCRIPTION_FILE"):
        print("::error::Description file environment variable is not specified")
        sys.exit(1)
    if not os.environ.get("PYXIS_API_TOKEN"):
        print("::error::API token environment variable is not specified")
        sys.exit(1)
    if not os.environ.get("API_URL"):
        print("::error::API URL environment variable is not specified")
        sys.exit(1)
    if not os.environ.get("PROJECT_ID"):
        print("RedHat project ID environment variable is not specified")
        sys.exit(1)


def update_registry_description(api_url, project_id, api_token, description):
    """PATCH the Red Hat container registry with the given description.

    Returns the requests.Response object.
    """
    data = {'container': {'repository_description': description[:32768]}}
    headers = {
        'accept': 'application/json',
        'X-API-KEY': api_token,
        'Content-Type': 'application/json',
    }
    return requests.patch(api_url + project_id, headers=headers, data=json.dumps(data))


def main():
    validate_env()

    repository_description = load_description(os.environ["DESCRIPTION_FILE"])

    if not repository_description:
        print("::error::No description file found")
        sys.exit(1)

    result = update_registry_description(
        os.environ["API_URL"],
        os.environ["PROJECT_ID"],
        os.environ["PYXIS_API_TOKEN"],
        repository_description,
    )

    print("::group::Result")
    print("Response code: " + str(result.status_code))
    print("Last update date: " + json.loads(result.content)['last_update_date'])
    print("::endgroup::")


if __name__ == '__main__':
    main()
