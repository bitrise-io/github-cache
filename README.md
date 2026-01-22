# Bitrise Cache GitHub Action

A GitHub Action that provides caching using Bitrise's cache backend, with an interface compatible with `actions/cache`.

## Usage

### Unified Action (Save and Restore)

Use the main action to handle both cache restoration and saving automatically:

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

### Separate Restore and Save Actions

For more control, use the separate restore and save actions:

#### Restore Only

```yaml
- uses: bitrise-io/github-cache/restore@v1
  id: cache
  with:
    path: |
      ~/.npm
      node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

#### Save Only

```yaml
- uses: bitrise-io/github-cache/save@v1
  with:
    path: |
      ~/.npm
      node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
```

#### Complete Example with Conditional Save

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # Restore cache
      - uses: bitrise-io/github-cache/restore@v1
        id: cache
        with:
          path: |
            ~/.npm
            node_modules
          key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
          restore-keys: |
            ${{ runner.os }}-node-

      # Install dependencies only if cache miss
      - name: Install dependencies
        if: steps.cache.outputs.cache-hit != 'true'
        run: npm ci

      # Build and test
      - name: Build
        run: npm run build

      - name: Test
        run: npm test

      # Save cache (only if dependencies changed)
      - uses: bitrise-io/github-cache/save@v1
        if: steps.cache.outputs.cache-hit != 'true'
        with:
          path: |
            ~/.npm
            node_modules
          key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
```

### When to Use Separate Actions

Use the separate `restore` and `save` actions when you need:

- **Conditional saving**: Save cache only under certain conditions (e.g., only on cache miss, only on specific branches)
- **Early exit workflows**: Restore cache early in the workflow but save it only if all steps succeed
- **Multiple cache operations**: Different restore/save logic in the same workflow
- **Explicit control**: Fine-grained control over when caching happens

Use the unified action when:

- **Simple workflows**: Default cache-on-miss, save-on-success behavior is sufficient
- **Less boilerplate**: You want minimal configuration
- **Standard pattern**: Your workflow follows the typical restore-build-save pattern

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `path` | A list of files, directories, and wildcard patterns to cache and restore | Yes | |
| `key` | An explicit key for restoring and saving the cache | Yes | |
| `restore-keys` | An ordered multiline string listing the prefix-matched keys for restoring stale cache | No | |
| `fail-on-cache-miss` | Fail the workflow if cache entry is not found | No | `false` |
| `lookup-only` | Check if a cache entry exists without downloading it | No | `false` |
| `verbose` | Enable verbose logging | No | `false` |

## Outputs

| Output | Description |
|--------|-------------|
| `cache-hit` | A boolean value indicating if an exact match was found for the primary key |

## How It Works

This action uses Bitrise's cache infrastructure instead of GitHub's cache. This can be beneficial when:

- You need more cache storage than GitHub provides
- You want to use Bitrise Runners for GitHub
- You want to use Bitrise's cache features

The action runs as a Node.js wrapper around a Go binary that handles the actual cache operations using Bitrise's cache SDK. The Go binary is automatically downloaded from GitHub releases at runtime.

### Cache Isolation

Cache artifacts are automatically scoped to each GitHub repository. The repository name is prepended to all cache keys to ensure that caches are not shared between different repositories. This happens transparently - you don't need to include the repository name in your cache keys.

For example, if you specify `key: node-modules-v1`, the actual cache key will be `myrepo-node-modules-v1`.

## Architecture

```
action.yml
    │
    ├── main (restore phase)
    │   └── dist/main/index.js → downloads bitrise-cache binary → restore
    │
    └── post (save phase)
        └── dist/post/index.js → uses cached binary → save
```

The Go binary is downloaded from GitHub releases on first run and cached in `~/.bitrise/bin/`.

## Development

### Prerequisites

- Node.js 20+
- Go 1.22+
- Make

### Building

```bash
# Install npm dependencies
npm install

# bump version in package.json (also used for releases)
npm version patch   # or minor, major

# Build JS bundles
make build
```

### Local Testing

```bash
# Build for current platform only
make build-local

# Test the binary directly
./bin/bitrise-cache restore
./bin/bitrise-cache save
```

### Releasing

Releases are managed via [GoReleaser](https://goreleaser.com/). When a new version is tagged, GoReleaser builds binaries for all platforms and publishes them to GitHub releases.

```bash
# Create a snapshot release (for testing)
make goreleaser-snapshot

# Create a release (requires a git tag)
make goreleaser
```

### Project Structure

```
github-cache/
├── action.yml          # GitHub Action definition
├── package.json        # Node.js dependencies (version defines release version)
├── Makefile            # Build automation
├── .goreleaser.yaml    # GoReleaser configuration
├── main.go             # Go source code
├── go.mod              # Go module definition
├── vendor/             # Go dependencies (vendored)
├── src/                # JavaScript source
│   ├── main.js         # Restore entry point
│   ├── post.js         # Save entry point
│   └── run.js          # Binary downloader and runner
└── dist/               # Built JS bundles (committed)
    ├── main/index.js
    └── post/index.js
```

## Requirements

This action requires the following environment variables to be set for Bitrise cache to work:

- `BITRISEIO_CACHE_SERVICE_URL` - Bitrise cache service URL
- `BITRISEIO_BUILD_API_TOKEN` - Bitrise build API token

These are automatically available in Bitrise builds. For GitHub Actions, you'll need to configure them as secrets.
