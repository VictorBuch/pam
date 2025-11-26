# Release Guide for pam

This guide explains how to create a new release of pam.

## Understanding Releases

### What is a Release?

A release is a packaged version of your software that users can download and use. For pam, each release includes:

1. **Git tag** - Marks the exact commit (e.g., `v0.1.0`)
2. **Binaries** - Compiled programs for different platforms
3. **Release notes** - Changelog and installation instructions
4. **Checksums** - SHA256 hashes to verify downloads

### Semantic Versioning

We use [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR** (1.0.0) - Breaking changes, incompatible API changes
- **MINOR** (0.1.0) - New features, backwards compatible
- **PATCH** (0.0.1) - Bug fixes, backwards compatible

Examples:
- `v0.1.0` → `v0.2.0` - Added new features
- `v0.2.0` → `v0.2.1` - Fixed bugs
- `v0.9.0` → `v1.0.0` - First stable release

## Release Process

### Step 1: Update CHANGELOG.md

Before releasing, document what changed:

```bash
# Edit CHANGELOG.md
vim CHANGELOG.md
```

Add a new section for your version:

```markdown
## [0.2.0] - 2025-02-15

### Added
- New feature X
- New feature Y

### Fixed
- Bug in feature Z

### Changed
- Improved performance of search
```

### Step 2: Commit and Push Changes

```bash
# Commit the changelog
git add CHANGELOG.md
git commit -m "docs: update changelog for v0.2.0"

# Push to main branch
git push origin main
```

**Why?** The changelog should be included in the release, so commit it before tagging.

### Step 3: Create and Push a Git Tag

```bash
# Create an annotated tag (recommended)
git tag -a v0.2.0 -m "Release v0.2.0"

# Push the tag to GitHub
git push origin v0.2.0
```

**What happens next?**
1. GitHub detects the new tag
2. The release workflow (`.github/workflows/release.yml`) starts automatically
3. GoReleaser builds binaries for all platforms
4. A GitHub Release is created with all the binaries

### Step 4: Monitor the Release

1. Go to GitHub Actions: `https://github.com/yourusername/pam/actions`
2. You'll see the "Release" workflow running
3. Wait for it to complete (usually 2-5 minutes)

If it succeeds ✅:
- Go to `https://github.com/yourusername/pam/releases`
- Your new release will be there with downloadable binaries

If it fails ❌:
- Click on the failed workflow to see logs
- Fix the issue
- Delete the tag and try again:
  ```bash
  # Delete local tag
  git tag -d v0.2.0
  # Delete remote tag
  git push --delete origin v0.2.0
  # Fix the issue, then create tag again
  ```

### Step 5: Verify the Release

After the workflow completes:

1. **Check GitHub Releases:**
   - Go to your repository's Releases page
   - Verify the version number is correct
   - Check that all 4 binaries are present:
     - `pam_0.2.0_linux_amd64.tar.gz`
     - `pam_0.2.0_linux_arm64.tar.gz`
     - `pam_0.2.0_darwin_amd64.tar.gz`
     - `pam_0.2.0_darwin_arm64.tar.gz`
   - Verify `checksums.txt` is present

2. **Test a Binary:**
   ```bash
   # Download the binary for your platform
   wget https://github.com/yourusername/pam/releases/download/v0.2.0/pam_0.2.0_linux_amd64.tar.gz

   # Verify checksum (optional but recommended)
   sha256sum pam_0.2.0_linux_amd64.tar.gz
   # Compare with value in checksums.txt

   # Extract and test
   tar -xzf pam_0.2.0_linux_amd64.tar.gz
   ./pam --version
   ```

3. **Update Release Notes (Optional):**
   - You can edit the release on GitHub to add:
     - Screenshots
     - Additional context
     - Breaking change warnings
     - Migration guides

## Making Your First Release (v0.1.0)

Let's create your first release right now!

### Prerequisites

1. **Ensure all tests pass:**
   ```bash
   go test ./...
   ```

2. **Ensure the code builds:**
   ```bash
   go build
   ```

3. **Commit all changes:**
   ```bash
   git status  # Should show clean working tree
   ```

### Execute the Release

```bash
# 1. Make sure you're on main branch
git checkout main
git pull origin main

# 2. CHANGELOG.md is already updated (we created it)
git add CHANGELOG.md .goreleaser.yml .github/workflows/release.yml RELEASING.md
git commit -m "chore: prepare for v0.1.0 release"
git push origin main

# 3. Create the tag
git tag -a v0.1.0 -m "Release v0.1.0 - Initial release"

# 4. Push the tag (this triggers the release)
git push origin v0.1.0
```

### What Happens Next

1. **Immediate** - Tag appears on GitHub
2. **~10 seconds** - GitHub Actions workflow starts
3. **~2-3 minutes** - Workflow builds all binaries
4. **~3-4 minutes** - Release is published with all assets

Watch it happen:
- Actions: `https://github.com/yourusername/pam/actions`
- Releases: `https://github.com/yourusername/pam/releases`

## Testing Releases Locally

Before pushing a tag, you can test the release process locally:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Run a local "snapshot" build (doesn't publish)
goreleaser release --snapshot --clean

# Check the dist/ folder for binaries
ls -lh dist/
```

This builds all the binaries without creating a release, perfect for testing.

## Rollback / Delete a Release

If something goes wrong:

```bash
# Delete the GitHub release (manually on GitHub UI)
# Then delete the tag:

# Delete local tag
git tag -d v0.2.0

# Delete remote tag
git push --delete origin v0.2.0
```

Then fix the issue and re-release.

## Common Issues

### Issue: "Tag already exists"

**Cause:** You tried to create a tag that already exists.

**Fix:**
```bash
git tag -d v0.2.0                    # Delete local
git push --delete origin v0.2.0      # Delete remote
# Now create it again
```

### Issue: "GoReleaser fails to build"

**Cause:** Usually a Go build error or missing dependencies.

**Fix:**
```bash
# Test locally first
go build
go test ./...
goreleaser release --snapshot --clean
```

### Issue: "Release workflow doesn't trigger"

**Cause:** Tag format doesn't match the pattern in `release.yml`.

**Fix:** Ensure tag follows the format: `v0.1.0` (starts with 'v', three numbers)

## Best Practices

1. **Always update CHANGELOG.md first** - Document your changes
2. **Test locally before tagging** - Run tests and build
3. **Use annotated tags** - `git tag -a` not `git tag`
4. **Follow semantic versioning** - Be consistent
5. **Keep releases small and frequent** - Don't accumulate too many changes
6. **Write clear release notes** - Help users understand what changed

## Release Checklist

Before creating a release:

- [ ] All tests pass (`go test ./...`)
- [ ] Code builds successfully (`go build`)
- [ ] CHANGELOG.md is updated with new version
- [ ] All changes are committed and pushed
- [ ] Working tree is clean (`git status`)
- [ ] You're on the main branch
- [ ] Version number follows semantic versioning

After creating a release:

- [ ] GitHub Actions workflow completed successfully
- [ ] All binaries are present in the GitHub Release
- [ ] Checksums are included
- [ ] You can download and run a binary
- [ ] Release notes look good

## Further Reading

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Git Tagging](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
