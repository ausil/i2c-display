# Release Process

This document describes the steps to cut a new release of i2c-display.

## Files to update for every release

| File | What to change |
|------|----------------|
| `VERSION` | New version number (e.g. `0.5.2`) |
| `rpm/i2c-display.spec` | `Version:` field |
| `debian/changelog` | New entry at the top via `dch` or by hand |
| `CHANGELOG.md` | Move `[Unreleased]` items to a new versioned section; add compare link |
| `BUILDING.md` | Update the "Current version" line and any example paths |
| `man/i2c-displayd.1` | Version string in `.TH` header |

## Step-by-step

### 1. Decide the version number

Follow [Semantic Versioning](https://semver.org/):

- **Patch** (`0.5.x`): bug fixes, CI/tooling changes, doc-only changes
- **Minor** (`0.x.0`): new user-visible features, backward-compatible
- **Major** (`x.0.0`): breaking changes to config schema, CLI, or behaviour

### 2. Update `VERSION`

```bash
echo "0.5.2" > VERSION
```

### 3. Update `rpm/i2c-display.spec`

Change the `Version:` field near the top of the file:

```spec
Version:                0.5.2
```

The release number is handled automatically by `%autorelease`.
The changelog is handled automatically by `%autochangelog` — no manual entry needed in the spec.

### 4. Update `debian/changelog`

Either use `dch`:

```bash
dch -v 0.5.2-1 "Brief description of changes"
```

Or add the entry manually at the top of `debian/changelog`:

```
i2c-display (0.5.2-1) unstable; urgency=medium

  * Brief summary of what changed

 -- Dennis Gilmore <dennis@ausil.us>  Sat, 22 Feb 2026 00:00:00 -0500
```

Use `date -R` to get the correct RFC 2822 timestamp.

### 5. Update `CHANGELOG.md`

Move items from `[Unreleased]` into a new versioned section and leave
`[Unreleased]` empty:

```markdown
## [Unreleased]

## [0.5.2] - 2026-02-22

### Added
- ...

### Changed
- ...

### Fixed
- ...
```

Add a compare link at the bottom of the file:

```markdown
[0.5.2]: https://github.com/ausil/i2c-display/compare/v0.5.1...v0.5.2
```

### 6. Update `BUILDING.md` and `man/i2c-displayd.1`

Replace the old version number with the new one.
For the man page, also update the date in the `.TH` header.

```bash
# Quick check — find all remaining old version strings
grep -r "0\.5\.1" . --include="*.md" --include="*.1" --include="*.spec"
```

### 7. Verify everything builds and tests pass

```bash
go build ./...
go test ./...
```

### 8. Commit the version bump

```bash
git add VERSION rpm/i2c-display.spec debian/changelog CHANGELOG.md \
        BUILDING.md man/i2c-displayd.1
git commit -m "Bump version to 0.5.2"
```

### 9. Tag the release

```bash
git tag -a v0.5.2 -m "Release v0.5.2"
```

### 10. Push branch and tag

```bash
git push origin main
git push origin v0.5.2
```

### 11. Build release artifacts (optional, on target system)

```bash
make dist       # Debian source tarball (with vendor)
make dist-rpm   # RPM source + vendor tarballs
make rpm        # Binary and source RPMs
make deb        # Debian binary package
make build-all  # Cross-compiled binaries for all architectures
```

### 12. Verify CI passes

Check that all jobs (Lint, Test, Build, Security Scan) are green before
announcing the release.

## Checklist

```
[ ] VERSION updated
[ ] rpm/i2c-display.spec Version: updated
[ ] debian/changelog entry added
[ ] CHANGELOG.md [Unreleased] items moved to versioned section
[ ] CHANGELOG.md compare link added
[ ] BUILDING.md version references updated
[ ] man/i2c-displayd.1 .TH version updated
[ ] go build ./... passes
[ ] go test ./... passes
[ ] git commit with all changed files
[ ] git tag -a v0.5.2 pushed
[ ] CI green
```
