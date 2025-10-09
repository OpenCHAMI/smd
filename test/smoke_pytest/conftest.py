# MIT License
#
# SPDX-FileCopyrightText: Copyright © 2025 OpenCHAMI a Series of LF Projects, LLC
#
# SPDX-License-Identifier: MIT

import json

from urllib.parse import urljoin

def pytest_addoption(parser):
    # Add extra arguments to pytest to support passing in data to configure smoke tests.
    parser.addoption(
        "--smoke-json",
        action="store",
        default="",
        required=True,
        help="Path to smoke.json",
    )
    parser.addoption(
        "--smoke-url",
        action="store",
        default=None,
        help="Base service url",
    )

def pytest_generate_tests(metafunc):
    print(metafunc.function, metafunc.fixturenames)
    if "smoke_test_data" in metafunc.fixturenames:
        # Generate tests cases based on the contents from the provided smoke.json file.

        # Read in smoke data
        print("Reading in smoke json file")
        ids = []
        testdata = []
        with open(metafunc.config.getoption("smoke_json"), 'r') as f:
            smoke_test = json.load(f)
            
            # Determine the base URL for the service
            base_url = smoke_test["default_base_url"]
            override_url = metafunc.config.getoption("smoke_url")
            if override_url is not None:
                base_url = override_url

            # Add a trailing slash if not present, needed by urljoin to work properly.
            if not base_url.endswith("/"):
                base_url += "/"

            for test_case in smoke_test["test_paths"]:
                test_case["url"] = urljoin(base_url, test_case["path"])
                testdata.append(test_case)

                # Override the string that is shown by pytest to be more informational in log output, instead of 'smoke_test_data0'.
                ids.append(f'Verify {test_case["method"]} {test_case["path"]}')
                # ids.append(json.dumps(test_case))

        metafunc.parametrize('smoke_test_data', testdata, ids=ids, indirect=False,)
