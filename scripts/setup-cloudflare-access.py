#!/usr/bin/env python3
"""Set up Cloudflare Access apps with an email allowlist for media hostnames."""
import json
import os
import sys
import urllib.error
import urllib.request

ACCOUNT_ID = os.environ.get("CF_ACCOUNT_ID", "5143b30f7d4d4f228f713f7b56f2fc96")
API_TOKEN = os.environ.get("CF_API_TOKEN", "")
ALLOWED_EMAILS = [
    e.strip()
    for e in os.environ.get(
        "CF_ACCESS_ALLOWED_EMAILS",
        "joshua.brock.hicks@gmail.com",
    ).split(",")
    if e.strip()
]
HOSTS = [
    ("Media Manager", "mm.jbhicks.dev"),
    ("Jellyfin", "media.jbhicks.dev"),
]


def api(method: str, path: str, body=None):
    if not API_TOKEN:
        print("ERROR: Set CF_API_TOKEN (Zero Trust Write permission).", file=sys.stderr)
        print("Create at: https://dash.cloudflare.com/profile/api-tokens", file=sys.stderr)
        print("Template: Account -> Zero Trust -> Edit", file=sys.stderr)
        sys.exit(1)
    data = None
    headers = {
        "Authorization": f"Bearer {API_TOKEN}",
        "Content-Type": "application/json",
    }
    if body is not None:
        data = json.dumps(body).encode()
    req = urllib.request.Request(
        f"https://api.cloudflare.com/client/v4{path}",
        data=data,
        headers=headers,
        method=method,
    )
    try:
        with urllib.request.urlopen(req) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        print(f"API error {e.code}: {e.read().decode()}", file=sys.stderr)
        sys.exit(1)


def list_apps():
    return api("GET", f"/accounts/{ACCOUNT_ID}/access/apps").get("result", []) or []


def ensure_app(name, domain, existing):
    for app in existing:
        if app.get("domain") == domain:
            print(f"App exists: {domain} ({app.get('id')})")
            return app["id"]
    payload = {
        "name": name,
        "domain": domain,
        "type": "self_hosted",
        "session_duration": "24h",
        "policies": [
            {
                "name": f"Allow listed emails - {domain}",
                "decision": "allow",
                "include": [{"email": {"email": email}} for email in ALLOWED_EMAILS],
                "precedence": 1,
            }
        ],
    }
    result = api("POST", f"/accounts/{ACCOUNT_ID}/access/apps", payload)
    app_id = result["result"]["id"]
    print(f"Created app: {domain} ({app_id})")
    return app_id


def main():
    print(f"Allowed emails: {', '.join(ALLOWED_EMAILS)}")
    existing = list_apps()
    for name, domain in HOSTS:
        ensure_app(name, domain, existing)
    print("Done. Only allowlisted emails can reach these hostnames.")


if __name__ == "__main__":
    main()