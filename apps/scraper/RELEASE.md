# Release Guide

This document explains how to build and release the QuakeWatch Scraper using GitHub Actions.

## Automated Releases

The project uses GitHub Actions to automatically build and release the application when you push a tag.

### Creating a Release

1. **Prepare your changes**: Make sure all your changes are committed and pushed to the master branch.

2. **Create and push a tag**: 
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Monitor the workflow**: The GitHub Action will automatically:
   - Build the application for multiple platforms (Linux, macOS, Windows)
   - Create a GitHub release with the built binaries
   - Generate release notes

### Supported Platforms

The release workflow builds binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### Manual Release

You can also trigger a release manually:
1. Go to the "Actions" tab in your GitHub repository
2. Select the "Build and Release" workflow
3. Click "Run workflow"
4. Choose the branch and click "Run workflow"

## Development Workflow

### Testing

The project includes a test workflow that runs on:
- Push to `master` or `dev` branches
- Pull requests to `master` branch

This workflow:
- Runs tests
- Builds the application for the current platform

### Local Development

For local development, you can use the Makefile:

```bash
# Install dependencies
make install

# Run tests
make test

# Build for current platform
make build

# Build for all platforms
make build-all

# Run the application
make run
```

## Release Notes

GitHub Actions automatically generates release notes based on:
- Commit messages since the last release
- Pull requests merged since the last release
- Issues closed since the last release

You can customize the release notes by editing the release after it's created.

## Versioning

We follow semantic versioning (SemVer):
- `MAJOR.MINOR.PATCH`
- Example: `v1.2.3`

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## Troubleshooting

### Common Issues

1. **Build fails**: Check that all dependencies are properly specified in `go.mod`
2. **Release not created**: Ensure you're pushing a tag that starts with `v` (e.g., `v1.0.0`)
3. **Permission denied**: Make sure the repository has the necessary permissions for creating releases

### Manual Build

If you need to build manually for a specific platform:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o quakewatch-scraper-linux-amd64 cmd/scraper/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o quakewatch-scraper-darwin-amd64 cmd/scraper/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o quakewatch-scraper-windows-amd64.exe cmd/scraper/main.go
``` 