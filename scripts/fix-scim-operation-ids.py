#!/usr/bin/env python3
"""
Fix SCIM OpenAPI operationId values so oapi-codegen produces sensible method names.

The upstream spec at dev.proof.com ships with copy-paste operationIds
(e.g. "retrieve-resource-types-copy" on GET /Users) that generate
confusing SDK methods like RetrieveResourceTypesCopyWithResponse.

This script rewrites the operationIds on the affected paths+methods to
accurate values keyed on the HTTP path and method. It is idempotent:
rerunning on a fixed spec is a no-op.
"""

import json
import sys


_PREFIX = "/scim/v1/organizations"

OPERATION_ID_OVERRIDES = {
    (f"{_PREFIX}/{{organization_id}}/Users/", "post"): "create-user",
    (f"{_PREFIX}/{{organization_id}}/Users", "get"): "list-users",
    (f"{_PREFIX}/{{organization_id}}/Users/{{user_id}}", "get"): "get-user",
    (f"{_PREFIX}/{{organization_id}}/Users/{{user_id}}", "put"): "replace-user",
    (f"{_PREFIX}/{{organization_id}}/Users/{{user_id}}", "patch"): "patch-user",
    (f"{_PREFIX}/{{organization_id}}/Users/{{user_id}}", "delete"): "delete-user",
    (f"{_PREFIX}/{{organization_id}}/Schemas/Users", "get"): "retrieve-users-schema",
    (f"{_PREFIX}/{{organization_id}}/ServiceProviderConfig", "get"): "get-service-provider-config",
    (f"{_PREFIX}/{{organization_id}}/ResourceTypes", "get"): "get-resource-types",
}


def fix_operation_ids(spec):
    changed = []
    paths = spec.get("paths", {})
    for path, methods in paths.items():
        if not isinstance(methods, dict):
            continue
        for method, op in methods.items():
            if not isinstance(op, dict):
                continue
            new_id = OPERATION_ID_OVERRIDES.get((path, method.lower()))
            if new_id and op.get("operationId") != new_id:
                changed.append((method.upper(), path, op.get("operationId"), new_id))
                op["operationId"] = new_id
    return changed


def main():
    if len(sys.argv) < 2:
        print("Usage: fix-scim-operation-ids.py <scim.json> [output.json]")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else input_file

    with open(input_file, "r") as f:
        spec = json.load(f)

    changed = fix_operation_ids(spec)

    with open(output_file, "w") as f:
        json.dump(spec, f, indent=2)

    if changed:
        print(f"Fixed operationIds in {input_file} -> {output_file}")
        for method, path, old, new in changed:
            print(f"  {method} {path}: {old!r} -> {new!r}")
    else:
        print(f"No operationId changes needed in {input_file}")


if __name__ == "__main__":
    main()
