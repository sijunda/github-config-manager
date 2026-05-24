# GCM Feature Demos

Individual animated demos for each GCM feature. Each GIF is self-contained and shows a single capability.

## Demos

| # | Feature | GIF | Tape Source |
|---|---------|-----|-------------|
| 01 | Version & Help | ![](01-version.gif) | [01-version.tape](../tapes/01-version.tape) |
| 02 | Create Profiles | ![](02-create-profiles.gif) | [02-create-profiles.tape](../tapes/02-create-profiles.tape) |
| 03 | List & Inspect Profiles | ![](03-list-profiles.gif) | [03-list-profiles.tape](../tapes/03-list-profiles.tape) |
| 04 | Switch Profiles | ![](04-switch-profiles.gif) | [04-switch-profiles.tape](../tapes/04-switch-profiles.tape) |
| 05 | SSH Key Management | ![](05-ssh.gif) | [05-ssh.tape](../tapes/05-ssh.tape) |
| 06 | GPG Commit Signing | ![](06-gpg.gif) | [06-gpg.tape](../tapes/06-gpg.tape) |
| 07 | GitHub Integration | ![](07-github.gif) | [07-github.tape](../tapes/07-github.tape) |
| 08 | Profile Validation | ![](08-validate.gif) | [08-validate.tape](../tapes/08-validate.tape) |
| 09 | System Diagnostics | ![](09-doctor.gif) | [09-doctor.tape](../tapes/09-doctor.tape) |
| 10 | Status Dashboard | ![](10-status.gif) | [10-status.tape](../tapes/10-status.tape) |
| 11 | Configuration Templates | ![](11-templates.gif) | [11-templates.tape](../tapes/11-templates.tape) |
| 12 | Backup & Restore | ![](12-backup.gif) | [12-backup.tape](../tapes/12-backup.tape) |
| 13 | Shell Integration | ![](13-shell.gif) | [13-shell.tape](../tapes/13-shell.tape) |
| 14 | Export & Diff | ![](14-export-diff.gif) | [14-export-diff.tape](../tapes/14-export-diff.tape) |

## Re-recording

To re-record all demos:

```bash
./docs/tapes/record-all.sh
```

To re-record a single demo:

```bash
vhs < docs/tapes/05-ssh.tape
```

## Full Combined Demo

The full combined demo (all features in one GIF) is at [`docs/demo.gif`](../demo.gif), generated from [`docs/demo.tape`](../demo.tape).
