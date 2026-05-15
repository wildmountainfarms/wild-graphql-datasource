---
name: Grafana GitHub Workflows
description: Documentation and instructions for keeping GitHub workflow YAML up to date to stay in sync with Grafana changes. Helps understand differences between Grafana's workflow templates and this plugin's workflow files.
---

# Grafana GitHub Workflows

At the root of the project is `.github` directory.
Within that is `workflows`.

```
.github
└── workflows
    ├── ci.yml
    ├── is-compatible.yml
    └── release.yml
```

## Workflow Files

- `ci.yml`

## Grafana Plugin Template Reference

The repository [grafana-plugin-examples](https://github.com/grafana/grafana-plugin-examples)
contains a [datasource-with-backend](https://github.com/grafana/grafana-plugin-examples/tree/main/examples/datasource-with-backend) example,
which contains its own [.github/workflows](https://github.com/grafana/grafana-plugin-examples/tree/main/examples/datasource-with-backend/.github/workflows) directory.

To access the files within that directory, you may fetch the URL
https://api.github.com/repos/grafana/grafana-plugin-examples/contents/examples/datasource-with-backend/.github/workflows

## Intentional modifications made to workflow files

Some workflows have intentional modifications made to them.

`ci.yml`
* "Lint backend" step is commented out until https://github.com/wildmountainfarms/wild-graphql-datasource/issues/27 is fixed

`is-compatible.yml`
* Should be triggers on `pull_request` (default) and also on pushes to the main branch (custom)

`release.yml`
* Must use the `staging` environment
* Permissions of the `release` job should include `id-token: write` and `attestations: write`
* `build-plugin` version was intentionally bumped to a higher than shown in the template workflow
* `build-plugin` action must have `policy_token`, `attestation` and `go-version` set
  * Note that the default `go-version` changes depending on the version of `build-plugin` used, so we explicitly set it. 
  * This version should be kept up to date with the version in [go.mod](../../../go.mod)

## Understanding differences ("diffing")

When asked to "diff" our CI/CD workflows with Grafana's template ones, you should do a couple of things:
- Answer "what files are present in Grafana's template but are not present here"
- For the files that are present in both, answer "what are the differences between them"
  - Include intentional differences, but make sure to highlight differences that may actually be updates that should be applied.


## Keeping up to date

**When asked** to update Grafana workflow files or to sync with Grafana's template ones, you should:

Report:
- Files present in Grafana's template but are not present here
  - These should not be added unless explicitly asked to add them
- Report "custom" workflows there are present here, but not in Grafana's template
  - These should not be removed or modified

For each workflow YAML present here and in Grafana's template, update it when you deem a difference something that needs to be updated
by confirming what you are "keeping in sync" is not an intentional modification.

