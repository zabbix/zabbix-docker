"""Tests for .github/scripts/rhel_description.py

The script has been refactored into three testable functions:
  - validate_env()           – validates required environment variables
  - load_description()       – loads description from .html or .md file
  - update_registry_description() – PATCHes the Red Hat Pyxis API
"""

import json
import os
import sys
import pytest
from pathlib import Path

# ---------------------------------------------------------------------------
# Make the script importable without executing main()
# ---------------------------------------------------------------------------
SCRIPT_DIR = Path(__file__).parent.parent.parent / ".github" / "scripts"
sys.path.insert(0, str(SCRIPT_DIR))

import rhel_description  # noqa: E402


# ===========================================================================
# validate_env()
# ===========================================================================

REQUIRED_ENV_VARS = {
    "DESCRIPTION_FILE": "/some/path/description",
    "PYXIS_API_TOKEN": "tok-abc123",
    "API_URL": "https://catalog.redhat.com/api/containers/v1/repositories/id/",
    "PROJECT_ID": "abc123project",
}


def _set_env(monkeypatch, overrides=None):
    """Set all required env vars, applying optional overrides / removals."""
    env = dict(REQUIRED_ENV_VARS)
    if overrides:
        env.update(overrides)
    for key, value in env.items():
        if value is None:
            monkeypatch.delenv(key, raising=False)
        else:
            monkeypatch.setenv(key, value)


class TestValidateEnv:
    def test_passes_when_all_vars_are_set(self, monkeypatch):
        _set_env(monkeypatch)
        # Should not raise
        rhel_description.validate_env()

    @pytest.mark.parametrize("missing_var", list(REQUIRED_ENV_VARS.keys()))
    def test_exits_when_required_var_is_missing(self, monkeypatch, missing_var):
        _set_env(monkeypatch, {missing_var: None})
        with pytest.raises(SystemExit) as exc_info:
            rhel_description.validate_env()
        assert exc_info.value.code == 1

    @pytest.mark.parametrize("empty_var", list(REQUIRED_ENV_VARS.keys()))
    def test_exits_when_required_var_is_empty_string(self, monkeypatch, empty_var):
        _set_env(monkeypatch, {empty_var: ""})
        with pytest.raises(SystemExit) as exc_info:
            rhel_description.validate_env()
        assert exc_info.value.code == 1

    def test_error_message_for_missing_description_file(self, monkeypatch, capsys):
        _set_env(monkeypatch, {"DESCRIPTION_FILE": None})
        with pytest.raises(SystemExit):
            rhel_description.validate_env()
        captured = capsys.readouterr()
        assert "Description file" in captured.out

    def test_error_message_for_missing_api_token(self, monkeypatch, capsys):
        _set_env(monkeypatch, {"PYXIS_API_TOKEN": None})
        with pytest.raises(SystemExit):
            rhel_description.validate_env()
        captured = capsys.readouterr()
        assert "API token" in captured.out

    def test_error_message_for_missing_api_url(self, monkeypatch, capsys):
        _set_env(monkeypatch, {"API_URL": None})
        with pytest.raises(SystemExit):
            rhel_description.validate_env()
        captured = capsys.readouterr()
        assert "API URL" in captured.out

    def test_error_message_for_missing_project_id(self, monkeypatch, capsys):
        _set_env(monkeypatch, {"PROJECT_ID": None})
        with pytest.raises(SystemExit):
            rhel_description.validate_env()
        captured = capsys.readouterr()
        assert "project ID" in captured.out


# ===========================================================================
# load_description()
# ===========================================================================

class TestLoadDescription:
    def test_loads_html_file(self, tmp_path):
        html_file = tmp_path / "desc.html"
        html_file.write_text("<p>Hello World</p>")
        result = rhel_description.load_description(str(tmp_path / "desc"))
        assert result == "<p>Hello World</p>"

    def test_loads_markdown_file_and_converts_to_html(self, tmp_path):
        md_file = tmp_path / "desc.md"
        md_file.write_text("# Hello\n\nWorld")
        result = rhel_description.load_description(str(tmp_path / "desc"))
        assert "<h1>" in result
        assert "Hello" in result
        assert "World" in result

    def test_prefers_html_over_markdown_when_both_exist(self, tmp_path):
        html_file = tmp_path / "desc.html"
        md_file = tmp_path / "desc.md"
        html_file.write_text("<p>From HTML</p>")
        md_file.write_text("From Markdown")
        result = rhel_description.load_description(str(tmp_path / "desc"))
        assert result == "<p>From HTML</p>"

    def test_returns_none_when_no_file_exists(self, tmp_path):
        result = rhel_description.load_description(str(tmp_path / "missing"))
        assert result is None

    def test_html_file_content_is_returned_verbatim(self, tmp_path):
        html_content = "<div><h1>Test</h1><p>Content with &amp; entities</p></div>"
        html_file = tmp_path / "desc.html"
        html_file.write_text(html_content)
        result = rhel_description.load_description(str(tmp_path / "desc"))
        assert result == html_content

    def test_load_description_returns_full_content_without_truncation(self, tmp_path):
        # load_description itself does NOT truncate; truncation happens in
        # update_registry_description.  Verify the full content is returned.
        long_content = "A" * 40000
        (tmp_path / "desc.html").write_text(long_content)
        result = rhel_description.load_description(str(tmp_path / "desc"))
        assert len(result) == 40000


# ===========================================================================
# update_registry_description()
# ===========================================================================

class TestUpdateRegistryDescription:
    def test_calls_patch_with_correct_url(self, mocker):
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-01-01T00:00:00Z"}
        ).encode()

        rhel_description.update_registry_description(
            "https://api.example.com/v1/repos/id/",
            "project123",
            "tok-abc",
            "Some description",
        )

        mock_patch.assert_called_once()
        call_args = mock_patch.call_args
        assert call_args[0][0] == "https://api.example.com/v1/repos/id/project123"

    def test_sends_correct_authorization_header(self, mocker):
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-01-01T00:00:00Z"}
        ).encode()

        rhel_description.update_registry_description(
            "https://api.example.com/v1/repos/id/",
            "project123",
            "supersecrettoken",
            "Some description",
        )

        headers = mock_patch.call_args[1]["headers"]
        assert headers["X-API-KEY"] == "supersecrettoken"

    def test_sends_content_type_json_header(self, mocker):
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-01-01T00:00:00Z"}
        ).encode()

        rhel_description.update_registry_description(
            "https://api.example.com/v1/repos/id/",
            "project123",
            "tok-abc",
            "Some description",
        )

        headers = mock_patch.call_args[1]["headers"]
        assert headers["Content-Type"] == "application/json"

    def test_description_is_truncated_to_32768_chars(self, mocker):
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-01-01T00:00:00Z"}
        ).encode()

        long_desc = "X" * 40000
        rhel_description.update_registry_description(
            "https://api.example.com/v1/repos/id/",
            "project123",
            "tok-abc",
            long_desc,
        )

        payload = json.loads(mock_patch.call_args[1]["data"])
        assert len(payload["container"]["repository_description"]) == 32768

    def test_returns_response_object(self, mocker):
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-01-01T00:00:00Z"}
        ).encode()

        response = rhel_description.update_registry_description(
            "https://api.example.com/v1/repos/id/",
            "project123",
            "tok-abc",
            "Description",
        )

        assert response is mock_patch.return_value

    def test_description_under_32768_is_not_truncated(self, mocker):
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-01-01T00:00:00Z"}
        ).encode()

        short_desc = "Short description"
        rhel_description.update_registry_description(
            "https://api.example.com/v1/repos/id/",
            "project123",
            "tok-abc",
            short_desc,
        )

        payload = json.loads(mock_patch.call_args[1]["data"])
        assert payload["container"]["repository_description"] == short_desc


# ===========================================================================
# main() integration tests
# ===========================================================================

class TestMain:
    def test_exits_when_description_file_missing(self, monkeypatch, tmp_path):
        _set_env(monkeypatch, {"DESCRIPTION_FILE": str(tmp_path / "nonexistent")})
        with pytest.raises(SystemExit) as exc_info:
            rhel_description.main()
        assert exc_info.value.code == 1

    def test_prints_response_code_on_success(self, monkeypatch, mocker, tmp_path):
        desc_file = tmp_path / "desc.html"
        desc_file.write_text("<p>Hello</p>")
        _set_env(
            monkeypatch,
            {
                "DESCRIPTION_FILE": str(tmp_path / "desc"),
                "API_URL": "https://api.example.com/v1/repos/id/",
                "PROJECT_ID": "proj123",
                "PYXIS_API_TOKEN": "tok",
            },
        )
        mock_patch = mocker.patch("rhel_description.requests.patch")
        mock_patch.return_value.status_code = 200
        mock_patch.return_value.content = json.dumps(
            {"last_update_date": "2024-06-01T12:00:00Z"}
        ).encode()

        import io
        from contextlib import redirect_stdout

        out = io.StringIO()
        with redirect_stdout(out):
            rhel_description.main()

        output = out.getvalue()
        assert "200" in output
        assert "2024-06-01T12:00:00Z" in output

    def test_exits_when_all_env_vars_missing(self, monkeypatch):
        for var in REQUIRED_ENV_VARS:
            monkeypatch.delenv(var, raising=False)
        with pytest.raises(SystemExit) as exc_info:
            rhel_description.main()
        assert exc_info.value.code == 1
