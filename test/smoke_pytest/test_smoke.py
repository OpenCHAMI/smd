#! /usr/bin/env python3
#
# SPDX-FileCopyrightText: Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import requests

def get_access_token():
    req = requests.request(url="http://opaal:3333/token", method="GET")
    return req.json().get("access_token")

def test_smoke(smoke_test_data):

    headers=smoke_test_data["headers"]
    if smoke_test_data.get("use_auth"):
        if "Authorization" not in headers:
            access_token = get_access_token()
            headers["Authorization"] = f"Bearer {access_token}"


    req = requests.request(url=smoke_test_data["url"], method=str(smoke_test_data["method"]).upper(), data=smoke_test_data["body"],
                                   headers=smoke_test_data["headers"])

    assert req.status_code == smoke_test_data["expected_status_code"], "unexpected status code."
