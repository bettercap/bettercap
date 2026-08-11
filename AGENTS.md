# Cutting a release

## ⚠️ STOP — read this before doing anything

Before bumping the version, tagging, or creating any release, **warn the user that "Immutable releases" must be OFF** under repo Settings → General → "Immutable releases" on `bettercap/bettercap`. Do not proceed until the user confirms it is disabled.

Why it matters:
- If immutability is on, any release published under a tag burns that tag **forever**. GitHub permanently reserves it — even deleting the release, disabling the setting, and contacting Support will not free it (per GitHub Support: "Once a tag has been associated with an immutable release, it becomes permanently reserved and can't be recreated or reused"). The version number is gone for good.
- This already happened on this repo: `v2.41.6` is permanently unusable. We shipped the same fixes as `v2.41.7` instead.
- The CI workflow uses `softprops/action-gh-release@v2`, whose draft → upload → publish flow has *historically* failed against immutable releases (PATCH on existing release rejected). Until that's verified safe with immutability on, keep it off.

Also: never run `gh release create` manually before CI gets to publish the tag. CI creates the release end-to-end on tag push; doing it manually first is what originally burned `v2.41.6`.

After the user confirms immutability is off, proceed with the steps below.

## 1. Generate the changelog

```
LAST_TAG=$(git describe --tags --abbrev=0)
git log "$LAST_TAG"..HEAD --oneline --pretty=format:"%h %s"
```

Write the changelog yourself from those commits. Format:

```markdown
### ✨ New Features
- ...

### 🐛 Fixes
- ...

### 🔧 Improvements
- ...

### Miscellaneous
- ...
```

Rules:
- Focus on new features and major fixes. Group dependabot bumps and trivial items under **Miscellaneous**.
- Emojis only on *important* items (e.g. 🛡️ for security/DoS fixes). Don't sprinkle them everywhere.
- Skip sections that are empty (e.g. omit "New Features" on a fix-only release).
- Show the draft to the user before tagging.

## 2. Cut the release

Current version lives in `core/banner.go` as `Version = "X.Y.Z"`. Ask the user for the next version, then:

```
# bump
sed -i 's/Version = "OLD"/Version = "NEW"/' core/banner.go   # or use the Edit tool
git add core/banner.go
git commit -m "releasing version NEW"
git push origin master

# tag
git tag -a vNEW -m "releasing vNEW"
git push origin vNEW
```

Pushing the tag triggers `.github/workflows/build-and-deploy.yml`, which builds binaries for macOS arm64 / Linux amd64 / Windows amd64 and creates the GitHub release with all 6 assets (zip + sha256 per platform) via `softprops/action-gh-release@v2`. Wait for the workflow with `gh run watch <id>`.

## 3. Attach the changelog

CI creates the release with an empty body. After CI finishes:

```
gh release edit vNEW --notes-file <changelog.md>
```

Then verify: `gh release view vNEW --json tagName,isDraft,assets`.
