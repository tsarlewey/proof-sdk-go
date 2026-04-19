#!/usr/bin/env python3
"""
Fix OpenAPI specs that have deep nested $ref pointing to properties within other schemas.
oapi-codegen doesn't support refs like #/components/schemas/document/properties/bundle_position
This script resolves those references by inlining the actual type definitions.
"""

import json
import sys
import os
import copy


def resolve_deep_ref(spec, ref):
    """
    Resolve a $ref path that goes deep into schema properties.
    Returns the resolved schema or None if not found.
    """
    if not ref.startswith('#/'):
        return None

    parts = ref[2:].split('/')
    current = spec

    try:
        for part in parts:
            current = current[part]
        return copy.deepcopy(current)
    except (KeyError, TypeError):
        return None


def is_deep_property_ref(ref):
    """Check if a $ref points to a property within a schema (problematic for oapi-codegen)."""
    if not ref:
        return False
    # Refs like #/components/schemas/document/properties/bundle_position
    return '/properties/' in ref


def fix_type_enum_mismatch(obj):
    """
    Fix schemas where type doesn't match enum values.
    e.g., type: boolean with enum: ["v1", "v2"] should be type: string.
    Only fix when 'type' is a string (not an object schema).
    """
    if isinstance(obj, dict):
        # Only fix if 'type' is a simple string value (not a nested schema object)
        if 'type' in obj and isinstance(obj['type'], str) and 'enum' in obj:
            enum_values = obj['enum']
            if enum_values and all(isinstance(v, str) for v in enum_values):
                # If all enum values are strings, type should be string
                if obj['type'] != 'string':
                    obj['type'] = 'string'

        # Recurse into nested objects
        for key, value in obj.items():
            fix_type_enum_mismatch(value)
    elif isinstance(obj, list):
        for item in obj:
            fix_type_enum_mismatch(item)


def fix_refs_in_object(spec, obj):
    """
    Recursively fix $ref in an object.
    Replace deep property refs with the actual schema definition.
    """
    if isinstance(obj, dict):
        if '$ref' in obj and is_deep_property_ref(obj['$ref']):
            # Resolve the deep reference
            resolved = resolve_deep_ref(spec, obj['$ref'])
            if resolved:
                # Replace the $ref with the resolved schema
                obj.clear()
                obj.update(resolved)
                # Recursively fix any refs in the resolved schema
                fix_refs_in_object(spec, obj)
            return

        for key, value in list(obj.items()):
            fix_refs_in_object(spec, value)

    elif isinstance(obj, list):
        for item in obj:
            fix_refs_in_object(spec, item)


def fix_openapi_spec(input_file, output_file=None):
    """Fix an OpenAPI spec file."""
    with open(input_file, 'r') as f:
        spec = json.load(f)

    # Fix all deep refs in components/schemas
    if 'components' in spec and 'schemas' in spec['components']:
        fix_refs_in_object(spec, spec['components']['schemas'])

    # Fix all deep refs in paths
    if 'paths' in spec:
        fix_refs_in_object(spec, spec['paths'])

    # Fix type/enum mismatches (e.g., type: boolean with enum: ["v1", "v2"])
    fix_type_enum_mismatch(spec)

    output = output_file or input_file
    with open(output, 'w') as f:
        json.dump(spec, f, indent=2)

    print(f"Fixed: {input_file} -> {output}")


def main():
    if len(sys.argv) < 2:
        print("Usage: fix-openapi-refs.py <openapi.json> [output.json]")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None

    fix_openapi_spec(input_file, output_file)


if __name__ == '__main__':
    main()
