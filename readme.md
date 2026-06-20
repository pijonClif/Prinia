# Prinia

Prinia is a command-line tool for downloading files concurrently.

## Build

```
go build -o prinia ./cmd/main.go
```

## Usage

```
./prinia -u <url> -f <filename> -s <sections>
```

*   `-u`: Download URL
*   `-f`: Output filename
*   `-s`: Number of sections to split the download into
