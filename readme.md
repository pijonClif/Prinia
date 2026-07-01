# Prinia

command-line tool for downloading files concurrently. Splits the target file into byte-range sections, downloads them in parallel, then merges them back into one file.

## build

```
go build -o prinia ./cmd/prinia
```

## usage

```
./prinia -u <url> -f <filename> -s <sections>
```

- `-u`: download URL
- `-f`: output filename
- `-s`: number of sections to split the download into
- `-dir`: download directory (default: `./downloads`)

pass `-dir` on its own (with no `-u`/`-f`/`-s`) to just set up the download directory without starting a download:

```
./prinia -dir ./my-downloads
```

## tests

```
go test ./...
```

covers section byte-range math (even/uneven splits, edge cases) and section merging/cleanup (ordering, missing-file errors, stale-file overwrite)
