#!/usr/bin/env python3
"""Gate govulncheck results on actionable findings only.

Reads the JSON stream written by ``govulncheck -format json`` and exits
non-zero only when a vulnerability is both reachable from this module's
code (a call-level finding) AND the advisory lists a fixed version we can
upgrade to.

Reachable vulnerabilities with no fix available (for example
golang.org/x/crypto/openpgp, which is flagged on every released version)
are printed for visibility but do not fail the job: nothing in this
repository can act on them until upstream ships a fix, and letting them
fail the workflow permanently hides regressions that we can fix.

Usage: govulncheck -format json ./... > out.json
       python3 .github/scripts/govulncheck-gate.py out.json
"""
import json
import sys


def read_stream(text):
    """Yield each top-level JSON object from a concatenated JSON stream."""
    decoder = json.JSONDecoder()
    pos, end = 0, len(text)
    while pos < end:
        while pos < end and text[pos].isspace():
            pos += 1
        if pos >= end:
            break
        obj, pos = decoder.raw_decode(text, pos)
        yield obj


def entry_point(finding):
    """Return 'file:line' of the frame in our code that reaches the vuln."""
    frame = finding["trace"][-1]
    position = frame.get("position") or {}
    where = position.get("filename", "?")
    if "line" in position:
        where = f"{where}:{position['line']}"
    return where, f"{frame.get('package', '')}.{frame.get('function', '')}"


def describe(osv_id, findings, osvs, traces):
    osv = osvs.get(osv_id, {})
    top = findings[0]["trace"][0]
    fixed = next((f["fixed_version"] for f in findings if f.get("fixed_version")), "N/A")
    print(f"  {osv_id}: {osv.get('summary', '')}")
    print(f"    module:   {top.get('module')}@{top.get('version')}")
    print(f"    fixed in: {fixed}")
    print(f"    info:     https://pkg.go.dev/vuln/{osv_id}")
    if traces:
        for finding in findings[:5]:
            where, caller = entry_point(finding)
            print(f"    - {where}: {caller} -> {top.get('package')}.{top.get('function')}")
        if len(findings) > 5:
            print(f"    ... {len(findings) - 5} more call paths")


def main():
    source = open(sys.argv[1], encoding="utf-8") if len(sys.argv) > 1 else sys.stdin
    osvs = {}
    findings = []
    with source:
        for obj in read_stream(source.read()):
            if "osv" in obj:
                osvs[obj["osv"]["id"]] = obj["osv"]
            elif "finding" in obj:
                findings.append(obj["finding"])

    reachable = {}
    not_reachable = set()
    for finding in findings:
        trace = finding.get("trace") or []
        if trace and trace[0].get("function"):
            reachable.setdefault(finding["osv"], []).append(finding)
        else:
            not_reachable.add(finding["osv"])
    not_reachable -= set(reachable)

    fixable = {k: v for k, v in reachable.items() if any(f.get("fixed_version") for f in v)}
    unfixable = {k: v for k, v in reachable.items() if k not in fixable}

    if fixable:
        print(f"FAIL: {len(fixable)} reachable vulnerabilities have a fix available. "
              "Upgrade the listed modules to clear them:")
        for osv_id in sorted(fixable):
            describe(osv_id, fixable[osv_id], osvs, traces=True)
        print()
    if unfixable:
        print(f"INFO: {len(unfixable)} reachable vulnerabilities have no fixed version yet "
              "(not gating; re-check when upstream publishes a fix):")
        for osv_id in sorted(unfixable):
            describe(osv_id, unfixable[osv_id], osvs, traces=False)
        print()
    if not_reachable:
        print(f"INFO: {len(not_reachable)} vulnerabilities in imported or required modules "
              "are not reachable from this module's code: "
              + ", ".join(sorted(not_reachable)))
    if not findings:
        print("No vulnerabilities found.")

    sys.exit(1 if fixable else 0)


if __name__ == "__main__":
    main()
