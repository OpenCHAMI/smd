# MIT License
#
# (C) Copyright [2023] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
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
