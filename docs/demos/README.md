# GCM Feature Demo

This folder contains the VHS source for the combined GCM demo. The walkthrough covers profile management, SSH/GPG key management, provider integrations, credential isolation, diagnostics, templates, and backup/restore.

## Demo Source

| Source | Output | Notes |
| ------ | ------ | ----- |
| [demo.tape](demo.tape) | [../demo.gif](../demo.gif) | Full CLI walkthrough with GitHub and GitLab provider commands |

## Re-recording

From the repository root:

```bash
vhs < docs/demos/demo.tape
```

The tape writes the generated animation to `docs/demo.gif`.
