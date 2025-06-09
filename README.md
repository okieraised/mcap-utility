# mcap-utility

A command-line utility to edit, trim, rename, shift, and compress `.mcap` files or directories containing `.mcap` files using the `edit` subcommand.

## Features
- Rename topics in MCAP files
- Trim messages by timestamp range
- Shift message log or publish timestamps
- Delete specific topics
- Switch ROS timestamps to use publish time
- Apply compression (lz4 or zstd) with selectable compression levels
- Process single files or entire directories
- Concurrent processing for faster batch operations

## Requirements
- Go 1.18+
- MCAP Go library

## Installation
- For current machine
    ```bash
    go build -o mcap-cli .
    ```
- For all OSes:
    ```shell
    make build-all
    ```

## Usage

```shell
mcap-cli edit -i <input> -o <output> [flags]
```

### Required Flags

- `-i`, `--input`: Path to input `.mcap` file or directory
- `-o`, `--output`: Path to output directory to save processed `.mcap` files

### Optional Flags

- `-r`, `--rename`: Rename topic mappings. Example:
  ```bash
  --rename /old_topic_1=/new_topic_1,/old_topic_2=/new_topic_2
  ```

- `-s`, `--trim-start`: Start timestamp to trim messages from. Formats supported:
    - RFC3339: `2006-01-02T15:04:05+07:00`
    - Unix Nano: `1672531199000000000`

- `-e`, `--trim-end`: End timestamp to trim messages to (same format as trim-start)

- `-l`, `--shift-log`: Duration to shift message log time. Example: `100ms`, `10s`, `-1h`

- `-p`, `--shift-pub`: Duration to shift message publish time

- `-t`, `--topics`: List of topics to apply time shift. If unspecified, all topics are affected

- `-d`, `--delete`: Topics to delete from MCAP files

- `-b`, `--pub-time`: Use publish time as the ROS timestamp

- `-c`, `--compression`: Compression algorithm: `lz4` or `zstd`

- `-n`, `--compression-level`: Compression level:
    - `0`: default
    - `1`: fastest
    - `2`: better
    - `3`: best

## Examples

### Rename a topic and apply zstd compression

```bash
mcap-cli edit -i input.mcap -o output_dir --rename /old_topic=/new_topic --compression zstd
```

### Trim messages between two timestamps

```bash
mcap-cli edit -i logs/ -o trimmed_logs/ --trim-start "2024-01-01T00:00:00Z" --trim-end "2024-01-01T01:00:00Z"
```

### Shift all publish times by +10 seconds

```bash
mcap-cli edit -i session.mcap -o shifted/ --shift-pub 10s
```

### Delete specific topics

```bash
mcap-cli edit -i record.mcap -o clean/ --delete /debug /logs
```

## Notes

- The output directory must not be the same as the input directory.
- If no edit flags are provided, the command exits with "Nothing to do".
- The tool will automatically process all `.mcap` files in a given directory if a folder is passed to `--input`.

## License

MIT
