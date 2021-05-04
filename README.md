# export-komoot

This is a proof-of-concept which allows you to export your planned tours from [Komoot](https://www.komoot.com).

Note that this is a unofficial tool which uses private API's from Komoot and can break at any time…

# Setup

Create a `.env` file which should include your username and password:

```env
KOMOOT_EMAIL=user@host.com
KOMOOT_PASSWD=password
```

# Running a full export

Run: `make run-full`

# Running an incremental export

Run: `make run-incremental`

# Usage

```
$ ./export-komoot -h
Usage: export-komoot [--email EMAIL] [--password PASSWORD] [--filter FILTER] [--format FORMAT] [--to TO] [--fulldownload] [--concurrency CONCURRENCY]

Options:
  --email EMAIL          Your Komoot email address
  --password PASSWORD    Your Komoot password
  --filter FILTER        Filter tours with name matching this pattern
  --format FORMAT        The format to export as: gpx or fit [default: gpx]
  --to TO                The path to export to
  --fulldownload         If specified, all data is redownloaded [default: false]
  --concurrency CONCURRENCY
                         The number of simultaneous downloads [default: 16]
  --help, -h             display this help and exit
```

# Caution

Use at your own risk!