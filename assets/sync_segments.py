from __future__ import annotations

import argparse
import json
import os
import sys
from typing import Any
from urllib.error import HTTPError
from urllib.request import Request, urlopen

from b1k5_capstone_model import predict_customer_segment


DEFAULT_API_BASE_URL = "http://localhost:8080/api"
DEFAULT_ADMIN_CUSTOMER_ID = "ADM-000001"
DEFAULT_ADMIN_PASSWORD = "123456"


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Predict customer segments from B1K5 API data and sync them back to the API."
    )
    parser.add_argument("--api-base-url", default=os.getenv("API_BASE_URL", DEFAULT_API_BASE_URL))
    parser.add_argument("--admin-customer-id", default=os.getenv("ADMIN_CUSTOMER_ID", DEFAULT_ADMIN_CUSTOMER_ID))
    parser.add_argument("--admin-password", default=os.getenv("ADMIN_PASSWORD", DEFAULT_ADMIN_PASSWORD))
    parser.add_argument("--customer-id", help="Limit sync to one customer_id, for example CUS-000001.")
    parser.add_argument("--dry-run", action="store_true", help="Print predictions without POSTing updates.")
    args = parser.parse_args()

    api = B1K5API(args.api_base_url)
    token = api.login(args.admin_customer_id, args.admin_password)

    users = api.get_users(token)
    if args.customer_id:
        users = [user for user in users if user.get("customer_id") == args.customer_id]

    predictions = []
    for user in users:
        if user.get("role") != "customer":
            continue

        customer_id = user.get("customer_id")
        if not customer_id:
            continue

        activity = api.get_user_activity(token, customer_id)
        predictions.append(predict_customer_segment(user, activity))

    payload_segments = [prediction.to_api_payload() for prediction in predictions]

    if args.dry_run:
        print(json.dumps({"segments": payload_segments}, indent=2))
        return 0

    if not payload_segments:
        print("No customer segments to sync.")
        return 0

    result = api.update_segments(token, payload_segments)
    print(json.dumps(result, indent=2))
    return 0


class B1K5API:
    def __init__(self, base_url: str) -> None:
        self.base_url = base_url.rstrip("/")

    def login(self, customer_id: str, password: str) -> str:
        response = self._request(
            "POST",
            "/auth/login",
            body={
                "customer_id": customer_id,
                "password": password,
            },
        )
        return str(response["token"])

    def get_users(self, token: str) -> list[dict[str, Any]]:
        response = self._request("GET", "/users/", token=token)
        return list(response.get("users") or [])

    def get_user_activity(self, token: str, customer_id: str) -> dict[str, Any]:
        return self._request("GET", f"/users/{customer_id}/activity", token=token)

    def update_segments(self, token: str, segments: list[dict[str, Any]]) -> dict[str, Any]:
        return self._request("POST", "/segments/update", token=token, body={"segments": segments})

    def _request(
        self,
        method: str,
        path: str,
        token: str | None = None,
        body: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        data = None
        headers = {
            "Accept": "application/json",
        }

        if body is not None:
            data = json.dumps(body).encode("utf-8")
            headers["Content-Type"] = "application/json"

        if token:
            headers["Authorization"] = f"Bearer {token}"

        request = Request(
            f"{self.base_url}{path}",
            data=data,
            headers=headers,
            method=method,
        )

        try:
            with urlopen(request, timeout=20) as response:
                return json.loads(response.read().decode("utf-8"))
        except HTTPError as error:
            detail = error.read().decode("utf-8")
            raise RuntimeError(f"{method} {path} failed with HTTP {error.code}: {detail}") from error


if __name__ == "__main__":
    sys.exit(main())
