# Release process

This repo's release process is git tag + README table update — nothing heavier. **There are no
GitHub Releases and no `CHANGELOG.md`; the "Versioning" table in `README.md` is the changelog.**
If you're looking for release notes, that table is the only place they live.

As of this writing the tag history is `v0.1.0` through `v0.6.0` (9 tags — `v0.2.1`, `v0.4.1`, and
`v0.4.2` are patch releases alongside the minor bumps). Run `git tag` to see the current list.

## Cutting a release

1. **Ship the change with its README table row in the same PR.** Add a new row to the top of the
   "Versioning" table in `README.md` (the table is newest-first), matching the exact format of the
   existing rows: `| vX.Y.Z | One-line description of what changed. |`. For a wire-format change,
   name the concrete fields that moved/changed size and say whether it's a flag-day break (see the
   `v0.5.0` and `v0.6.0` rows for the pattern — both call out "Flag-day, no backcompat"
   explicitly). See `CONTRIBUTING.md`'s "Semver rules" table for what bump (patch/minor) a given
   kind of change gets.

   In practice this table update has sometimes lagged a release by a follow-up commit (the
   `v0.4.0`–`v0.4.2` rows were backfilled together after the fact) — but the current convention,
   and the one to follow, is to include the table row in the same commit/PR as the change it
   describes, so the commit that ends up tagged already has its own changelog entry.

2. **Merge to `main`.**

3. **Tag the merge commit:**

   ```sh
   git tag vX.Y.Z
   ```

4. **Push the tag:**

   ```sh
   git push origin vX.Y.Z
   ```

   This is the documented convention (see `CONTRIBUTING.md`), and it's what every tag in this
   repo's history is consistent with — push the one tag you just created, not `git push --tags`.
   It also doesn't matter mechanically: none of the workflows in `.github/workflows/` (`ci.yml`,
   `codeql.yml`, `dependency-review.yml`) trigger on tag push, only on `push: branches: [main]` and
   `pull_request`. Pushing the tag doesn't kick off any CI — it's purely how consumers (`git fetch
   --tags`, Go module resolution) discover the release exists.

That's the whole process. No release-branch step, no version-bump commit separate from the
feature, no draft/publish flow — the tagged commit on `main` *is* the release.

## Coordination handoff

After pushing the tag, tell the `lattice-nodes` and `lattice-hub` maintainers a new version
exists — especially if it's a flag-day wire-format change (see
[`docs/ecosystem.md`](ecosystem.md) for why flag-day releases require coordinated rollout rather
than independent upgrades) — so they can bump their pin using the commands in
[`docs/consuming_a_release.md`](consuming_a_release.md).
