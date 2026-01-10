#!/usr/bin/env python3
import argparse
import base64
import json
import os
import sys
import time


def b64url_decode(data: str) -> bytes:
    padding = "=" * ((4 - len(data) % 4) % 4)
    return base64.urlsafe_b64decode(data + padding)


def load_token(token: str) -> tuple[dict, dict]:
    parts = token.strip().split(".")
    if len(parts) < 2:
        raise ValueError("token must have at least two JWT segments")
    header = json.loads(b64url_decode(parts[0]))
    payload = json.loads(b64url_decode(parts[1]))
    return header, payload


def roles_from_payload(payload: dict, client_id: str) -> list[str]:
    roles: list[str] = []
    realm = payload.get("realm_access", {})
    if isinstance(realm, dict):
        r = realm.get("roles", [])
        if isinstance(r, list):
            roles.extend([str(x) for x in r])
    resource = payload.get("resource_access", {})
    if isinstance(resource, dict):
        client = resource.get(client_id, {})
        if isinstance(client, dict):
            cr = client.get("roles", [])
            if isinstance(cr, list):
                roles.extend([str(x) for x in cr])
    return sorted(set(roles))


def fmt_ts(value) -> str:
    try:
        ts = int(value)
    except (TypeError, ValueError):
        return ""
    return time.strftime("%Y-%m-%d %H:%M:%S %Z", time.localtime(ts))


def main() -> int:
    default_client_id = os.getenv("AUTH_CLIENT_ID", "hackathon-service-api-client")
    parser = argparse.ArgumentParser(description="Decode a JWT and print roles/claims (no signature verification).")
    parser.add_argument("token", nargs="?", help="JWT access token")
    parser.add_argument("--client-id", default=default_client_id, help="client_id for resource_access roles")
    args = parser.parse_args()

    token = args.token
    if not token:
        if sys.stdin.isatty():
            parser.print_usage()
            return 1
        token = sys.stdin.read().strip()
    if not token:
        print("empty token", file=sys.stderr)
        return 1

    try:
        header, payload = load_token(token)
    except Exception as exc:
        print(f"failed to decode token: {exc}", file=sys.stderr)
        return 1

    roles = roles_from_payload(payload, args.client_id)
    admin_roles = {"hackathon_admin", "hackathon_organizer"}
    has_admin = any(r in admin_roles for r in roles)

    print("client_id:", args.client_id)
    print("issuer:", payload.get("iss", ""))
    print("subject:", payload.get("sub", ""))
    print("email:", payload.get("email", ""))
    print("preferred_username:", payload.get("preferred_username", ""))
    print("audience:", payload.get("aud", ""))
    print("authorized_party:", payload.get("azp", ""))
    print("expires_at:", fmt_ts(payload.get("exp")))
    print("admin_or_organizer:", "yes" if has_admin else "no")
    print("roles:", ", ".join(roles) if roles else "(none)")

    if os.getenv("TOKEN_CHECK_DEBUG", ""):
        print("\nheader:")
        print(json.dumps(header, indent=2, sort_keys=True))
        print("\npayload:")
        print(json.dumps(payload, indent=2, sort_keys=True))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
