import json, sys

version = sys.argv[1]

new_version = {
    "version": version,
    "title": version,
    "aliases": [
        "latest"
    ]
}

with open('../docs/versions.json', 'r+') as f:
    data = json.load(f)
    f.seek(0)
    for version in data:
        version['aliases'] = []
    data.append(new_version)
    json.dump(data, f, indent=4)
    f.truncate()