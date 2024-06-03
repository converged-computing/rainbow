import argparse
import os
import sys
import json

import sqlite3

# Config file from a few directories up
here = os.path.abspath(os.path.dirname(__file__))
root = here

# rainbow root directory
for iter in range(4):
    root = os.path.dirname(root)


def get_parser():
    parser = argparse.ArgumentParser(description="🌈️ Rainbow database reader")
    parser.add_argument(
        "--path",
        help="path to sqlite database file",
        default=os.path.join(root, "rainbow.db"),
    )
    return parser


def main():
    parser = get_parser()
    args, _ = parser.parse_known_args()

    if not os.path.exists(args.path):
        sys.exit(f'{args.path} does not exist')

    print(f'Connecting to database {args.path}...')
    conn = sqlite3.connect(args.path)
    cursor = conn.cursor()
    print(cursor)

    print("\n🥣️ Inspecting clusters table:")
    res = cursor.execute("SELECT * FROM clusters")
    for cluster in res.fetchall():
        print(f'  name: {cluster[0]}, token: {cluster[1]}, secret: {cluster[2]}')

    print("\n📐️ Inspecting metrics table:")
    res = cursor.execute("SELECT * FROM metrics")
    for metric in res.fetchall():
        result = f'  id: {metric[0]}, name: {metric[1]}, value: {metric[2]}'
        metadata = metric[3]
        if metadata is not None:
            result += f', metadata: {metadata}'
        print(result)

    print("\n💼️ Inspecting jobs table:")
    res = cursor.execute("SELECT * FROM jobs")
    for job in res.fetchall():
        print(f'  id: {job[0]}, cluster: {job[1]}, name: {job[2]}, jobspec: {json.dumps(job[3])}')
    conn.close()


if __name__ == "__main__":
    main()
