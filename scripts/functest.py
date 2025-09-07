#!/usr/bin/env python3

"""
runtests.py - Functional Test Runner for the Monogram Command-Line Tool

Overview:
-----------
This script provides a functional test runner for the Monogram tool, or
command-line tools in general, which converts text on the standard input into
structured outputs (like XML or JSON). Test cases are defined in a YAML file
that specifies the command to run, the input data (fed via standard input), and
the expected output. Additionally, the YAML file (or individual tests) may
specify a normalization key (e.g., "xml", "json") to preprocess outputs before
comparison, helping to eliminate discrepancies due to formatting differences.

Usage:
-----------
Run the test runner from the command line by providing the YAML test data file
using the --tests option. For example:

    python3 scripts/functest.py --tests tests.yaml

Normalization:
-----------
Normalization functions (such as for XML or JSON) can be specified either
globally in the YAML file (under "normalize") or overridden on a per-test basis.
This helps ensure that output differences in whitespace, attribute order, or key
order do not cause superficial test failures.

Output:
-----------
- Each test outputs a PASS or FAIL message.
- On failure, differences between the expected and actual output are displayed
  via a unified diff.
- A summary is printed at the end.
- Error messages and test failures are printed to stderr; normal pass messages
  are printed to stdout.
- The script exits with status code 0 if all tests pass, or 1 if any test fails.

"""

import argparse
import subprocess
import sys
import difflib
import yaml
import json
from pathlib import Path
from lxml import etree


def normalize_xml(xml_str):
    """
    Parse and pretty-print XML so that minor differences in whitespace,
    attribute order, etc., are eliminated. Non-ASCII characters are escaped
    using numeric character references (&# codes).
    """
    try:
        # Parse the XML string
        parser = etree.XMLParser(remove_blank_text=True)
        root = etree.XML(xml_str, parser)

        # Canonicalize the XML to standardize attribute order
        canonicalized = etree.tostring(
            root,
            method="c14n",  # Canonical XML ensures consistent attribute order
            exclusive=True,  # Use exclusive canonicalization
            with_comments=False  # Exclude comments from the output
        )

        # Re-parse the canonicalized XML for pretty-printing
        pretty_root = etree.XML(canonicalized)
        pretty_xml = etree.tostring(
            pretty_root,
            pretty_print=True,  # Pretty-print the XML with consistent line breaks
            encoding="ascii",  # Output as a string
            method="xml"
        ).decode("ascii")  # Decode the byte string into a Unicode string

        # Return the pretty-printed XML
        return pretty_xml.strip()
    except Exception:
        return xml_str

def normalize_json(json_str):
    """
    Load and re-dump JSON so that differences in spacing or key order are normalized.
    """
    try:
        obj = json.loads(json_str)
        return json.dumps(obj, sort_keys=True, indent=2)
    except Exception:
        return json_str

def normalize_yaml(yaml_str):
    """
    Load and re-dump YAML to normalize differences in spacing or key order.
    """
    try:
        obj = yaml.safe_load(yaml_str)
        return yaml.dump(obj, sort_keys=True, default_flow_style=False)
    except Exception:
        return yaml_str

# Mapping of normalization keys to functions.
normalization_functions = {
    "xml": normalize_xml,
    "json": normalize_json,
    "yaml": normalize_yaml,
    # If an unrecognized value (or "none") is provided, no normalization is done.
}

class Main:

    def __init__(self):
        parser = argparse.ArgumentParser(
            description="Functional test runner for the Monogram command-line tool."
        )
        # --tests to accept multiple files.
        parser.add_argument(
            "--tests", 
            required=True, 
            nargs="+",  # Accept multiple files
            help="One or more YAML files containing test data"
        )
        parser.add_argument(
            "--command",
            default="monogram",
            help="Path to the executable to test."
        )
        parser.add_argument(
            "--quiet",
            action="store_true",
            help="Suppress output for passing tests."
        )
        self.args = parser.parse_args()

    def run_test(self, tcount, test, default_normalize=None):
        """
        Execute a single test case.
        The normalization setting is determined first by a test-specific flag
        and then falls back to the default.
        """
        name = test.get("name", "<unnamed>")
        command = test["command"].format(command=Main().args.command).format(count=tcount)
        input_text = test.get("input", "")
        expected_output = test.get("expected_output", "")
        expected_exit_status = int(test.get("expected_exit_status", "0"))

        norm_key = test.get("normalize", default_normalize)
        normalization_func = normalization_functions.get(norm_key, None)

        result = subprocess.run(
            command,
            input=input_text,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            shell=True,
            text=True
        )
        actual_output = result.stdout

        if result.returncode != expected_exit_status:
            return (name, False, f"EXIT STATUS {result.returncode}", f"EXPECTED STATUS {expected_exit_status}", result.stderr)

        if normalization_func is not None:
            actual_output = normalization_func(actual_output)
            expected_output = normalization_func(expected_output)

        passed = (actual_output.strip() == expected_output.strip())
        return name, passed, actual_output, expected_output, result.stderr

    def run_single_test(self, tcount, test, default_normalize):
        """
        Run a single test and print its result.
        If the test fails or there is a command error, warning messages are sent to stderr.
        Returns True if the test passed, else False.
        """
        name, passed, actual, expected, stderr_text = self.run_test(
            tcount,
            test,
            default_normalize=default_normalize
        )
        if passed:
            if not self.args.quiet:
                print(f"PASS: {name}")
        else:
            print(f"FAIL: {name}")
            print("Expected:")
            print(expected)
            print("Actual:")
            print(actual)
            diff = difflib.unified_diff(
                expected.splitlines(),
                actual.splitlines(),
                fromfile="expected",
                tofile="actual",
                lineterm=""
            )
            print("Diff:")
            print("\n".join(diff))
            if stderr_text:
                print("Error output:")
                print(stderr_text)
            print("-" * 40)
        return passed
    
    def main(self):
        all_tests = []
        for test_file in self.args.tests:
            try:
                with open(test_file, "r", encoding="utf-8") as f:
                    data = yaml.safe_load(f)
                    default_normalize = data.get("normalize", None)
                    tests = data.get("tests", [])
                    if not tests:
                        print(f"No tests found in {test_file}!", file=sys.stderr)
                        continue
                    all_tests.extend((default_normalize, test) for test in tests)
            except Exception as e:
                print(f"Error reading {test_file}: {e}", file=sys.stderr)
                sys.exit(1)

        if not all_tests:
            print("No valid tests found in the provided YAML files!", file=sys.stderr)
            sys.exit(1)

        total = 0
        passed_count = 0

        for tcount, (default_normalize, test) in enumerate(all_tests):
            total += 1
            if self.run_single_test(tcount, test, default_normalize):
                passed_count += 1

        print(f"\nSummary: {passed_count} out of {total} tests passed.")
        sys.exit(0 if passed_count == total else 1)

if __name__ == "__main__":
    Main().main()
