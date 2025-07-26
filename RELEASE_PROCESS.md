# Release Process

This document explains how releases are automatically created in the QuakeWatch project.

## Automated Release Process

### Overview

Releases are automatically triggered when pull requests are merged into specific branches:

- **Master Branch**: Creates a proper release (e.g., `v1.2.2`)
- **Dev Branch**: Creates a development release (e.g., `dev-release-v1.2.2`)

### How It Works

1. **Pull Request Creation**: Create a PR targeting either `master` or `dev` branch
2. **Version Update**: Ensure the version in `apps/scraper/pkg/cli/commands.go` is updated
3. **Merge**: When the PR is merged, the automation will:
   - Extract the version from the code
   - Run tests to ensure quality
   - Build binaries for all supported platforms
   - Create a GitHub release with the appropriate tag
   - Mark dev releases as pre-releases

### Supported Platforms

The automation builds binaries for:
- **Linux**: AMD64, ARM64
- **macOS**: AMD64, ARM64  
- **Windows**: AMD64, ARM64

### Release Types

#### Production Releases (Master Branch)
- **Tag Format**: `v1.2.2`
- **Release Type**: Full release
- **Pre-release**: No
- **Use Case**: Stable, production-ready releases

#### Development Releases (Dev Branch)
- **Tag Format**: `dev-release-v1.2.2`
- **Release Type**: Pre-release
- **Pre-release**: Yes
- **Use Case**: Testing and development builds

### Manual Release Process

If you need to create a release manually:

1. Go to the GitHub Actions tab
2. Select "Build and Release (Manual)"
3. Click "Run workflow"
4. Choose the branch and version
5. Click "Run workflow"

### Version Management

#### Updating Version

To update the version:

1. Edit `apps/scraper/pkg/cli/commands.go`
2. Update both version references:
   ```go
   // In runVersion function
   fmt.Println("QuakeWatch Scraper v1.2.2")
   
   // In showBanner function  
   fmt.Println("║  Version: 1.2.2                                              ║")
   ```
3. Commit and push the changes
4. Create a PR to the target branch

#### Version Format

Use semantic versioning: `MAJOR.MINOR.PATCH`
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Workflow Files

- **`.github/workflows/pr-release.yml`**: Main automation for PR-based releases
- **`.github/workflows/release.yml`**: Manual release workflow (disabled automatic triggers)
- **`.github/workflows/test.yml`**: Testing workflow for PRs

### Best Practices

1. **Always update version** before creating a PR
2. **Test locally** before pushing changes
3. **Use descriptive PR titles** and descriptions
4. **Review the automation** logs if issues occur
5. **Tag important releases** for easy reference

### Troubleshooting

#### Common Issues

1. **Version not found**: Ensure version is properly formatted in `commands.go`
2. **Build failures**: Check Go version compatibility and dependencies
3. **Permission errors**: Ensure GitHub Actions have proper permissions
4. **Tag conflicts**: Delete conflicting tags before re-running

#### Getting Help

- Check GitHub Actions logs for detailed error messages
- Review the workflow files for configuration issues
- Ensure all required secrets are configured in repository settings 