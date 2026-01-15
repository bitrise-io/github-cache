# Bitrise Cache GitHub Action

A GitHub Action that provides caching functionality using Bitrise's cache backend while maintaining the same interface as the official `actions/cache` action.

## Features

- **Drop-in replacement**: Same input/output interface as `actions/cache`
- **Bitrise cache backend**: Uses Bitrise's high-performance cache storage
- **Automatic restore/save**: Restores cache at the start of the job and saves it at the end
- **Restore key fallback**: Supports fallback restore keys for partial cache hits
- **zstd compression**: Uses zstd compression for fast and efficient archiving

## Usage

```yaml
- uses: bitrise-io/github-cache@v1
  with:
    path: |
      ~/.npm
      node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `path` | A list of files, directories, and wildcard patterns to cache and restore | Yes | |
| `key` | An explicit key for restoring and saving the cache | Yes | |
| `restore-keys` | An ordered multiline string listing prefix-matched keys for fallback restore | No | |
| `fail-on-cache-miss` | Fail the workflow if cache entry is not found | No | `false` |
| `lookup-only` | Check if cache exists without downloading | No | `false` |
| `verbose` | Enable verbose logging | No | `false` |

## Outputs

| Output | Description |
|--------|-------------|
| `cache-hit` | A boolean value indicating if an exact match was found for the primary key |

## Environment Variables

The following environment variables must be set for the action to work:

| Variable | Description |
|----------|-------------|
| `BITRISEIO_ABCS_API_URL` | Bitrise cache API base URL |
| `BITRISEIO_BITRISE_SERVICES_ACCESS_TOKEN` | Bitrise services access token |

## Example Workflows

### Basic Usage

```yaml
name: Build
on: push

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Cache node modules
        uses: bitrise-io/github-cache@v1
        with:
          path: node_modules
          key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
          restore-keys: |
            ${{ runner.os }}-node-
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build
        run: npm run build
```

### Multiple Cache Paths

```yaml
- uses: bitrise-io/github-cache@v1
  with:
    path: |
      ~/.gradle/caches
      ~/.gradle/wrapper
    key: ${{ runner.os }}-gradle-${{ hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') }}
    restore-keys: |
      ${{ runner.os }}-gradle-
```

### Fail on Cache Miss

```yaml
- uses: bitrise-io/github-cache@v1
  with:
    path: node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    fail-on-cache-miss: true
```

## How It Works

1. **Restore Phase (main step)**: 
   - Attempts to find a cache entry matching the provided key
   - Falls back to restore-keys if no exact match is found
   - Downloads and extracts the cache archive
   - Sets `cache-hit` output to `true` for exact matches, `false` otherwise

2. **Save Phase (post step)**:
   - Skips saving if an exact cache hit occurred during restore
   - Creates a compressed archive of the specified paths
   - Uploads the archive to Bitrise cache storage

## Differences from actions/cache

- Uses Bitrise cache backend instead of GitHub's cache
- Requires Bitrise credentials via environment variables
- Some advanced features (like `upload-chunk-size`, `enableCrossOsArchive`, `save-always`) are not fully supported

## Building

```bash
go build -o bitrise-cache .
```

## Docker

The action runs in a Docker container. To build locally:

```bash
docker build -t bitrise-cache .
```
