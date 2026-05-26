# Split GCM Demo

This folder contains one VHS tape per demo point from the combined walkthrough.
Each tape writes one GIF in the same folder.

## Recording

From the repository root:

```bash
for tape in docs/demo/*.tape; do
  vhs < "$tape"
done
```

The combined walkthrough source remains in `docs/demos/demo.tape` and writes `docs/demo.gif`.

## Demo Points

| Point | Tape | Output |
| ----- | ---- | ------ |
| 1. Version & Overview | `01-version-overview.tape` | `01-version-overview.gif` |
| 2. Create Git Profiles | `02-create-git-profiles.tape` | `02-create-git-profiles.gif` |
| 3. List & Inspect Profiles | `03-list-inspect-profiles.tape` | `03-list-inspect-profiles.gif` |
| 4. Switch Between Profiles | `04-switch-between-profiles.tape` | `04-switch-between-profiles.gif` |
| 5. SSH Key Management | `05-ssh-key-management.tape` | `05-ssh-key-management.gif` |
| 6. GPG Commit Signing | `06-gpg-commit-signing.tape` | `06-gpg-commit-signing.gif` |
| 7. Provider Integrations | `07-provider-integrations.tape` | `07-provider-integrations.gif` |
| 8. Auth Ownership & Sources | `08-auth-ownership-sources.tape` | `08-auth-ownership-sources.gif` |
| 9. Validate Profiles | `09-validate-profiles.tape` | `09-validate-profiles.gif` |
| 10. System Health Check | `10-system-health-check.tape` | `10-system-health-check.gif` |
| 11. Status Dashboard | `11-status-dashboard.tape` | `11-status-dashboard.gif` |
| 12. Configuration Templates | `12-configuration-templates.tape` | `12-configuration-templates.gif` |
| 13. Backup & Restore | `13-backup-restore.tape` | `13-backup-restore.gif` |
| 14. Shell Integration | `14-shell-integration.tape` | `14-shell-integration.gif` |
| 15. Profile Export & Diff | `15-profile-export-diff.tape` | `15-profile-export-diff.gif` |
