# Desktop/MSIX Consumption and CDN Mirror

[Back to release artifacts](../release-artifacts.md)

Riido Desktop should treat GitHub Release assets like an open-source helper
binary source:

1. select the asset by platform and architecture;
2. download it over HTTPS into the app user data area;
3. verify `SHA256SUMS`;
4. extract the binary under user-writable app data;
5. launch the daemon with `RIIDO_DEVICE_ID`, `RIIDO_DEVICE_SECRET`, and
   `RIIDO_SAAS_URL`.

The Desktop/MSIX launcher must not install the daemon under Program Files or
another administrator-owned path. Microsoft Store/MSIX packaging should keep the
downloaded daemon under package local data or another user-writable app data
root, then execute it as the current user.

This release channel does not replace Store package update rules. Store app
updates remain owned by the Desktop packaging target. The daemon release asset
is only the helper binary that the launcher chooses and runs.

## CDN latest mirror

GitHub Release assets are the immutable daemon binary source. The CDN path
`https://cdn.riido.io/releases/latest/ai-agent/` is a mutable Desktop
development/test mirror of those release assets, not a separate build source.

When Desktop consumes the CDN mirror, the operator must update it from a tagged
GitHub Release asset with the same archive name, then invalidate the CloudFront
paths before asking client developers to retest:

- `releases/latest/ai-agent/riido-daemon_darwin_arm64.tar.gz`
- `releases/latest/ai-agent/riido-daemon_darwin_amd64.tar.gz`

The archive `VERSION` file must identify the release tag that was mirrored. A
stale CDN mirror is a release defect even when GitHub Releases already contain a
newer daemon tag.
