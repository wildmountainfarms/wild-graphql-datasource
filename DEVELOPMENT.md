# Development

This document contains information for developers and those wanting to contribute to the project.

https://grafana.com/developers/plugin-tools/create-a-plugin/develop-a-plugin/best-practices

https://grafana.com/developers/plugin-tools/tutorials/build-a-data-source-backend-plugin

## Testing

In one shell, run this:

```shell
# keep this in the background
npm run dev
```
And in another shell:
```shell
# Note that build:debug is used here to allow debugging
mage -v build:debug && npm run server
# Kill and restart this as necessary
```

---

## Updating dependencies

For security reasons, please do not make PRs with massive changes to `package-lock.json`.
If you would like dependencies to be up to date please make a PR for me ([@retrodaredevil](https://github.com/retrodaredevil)) to do so.

https://grafana.com/developers/plugin-tools/reference/cli-commands#update

```shell
npx @grafana/create-plugin@latest update
```

```shell
npm update
go get -u
go mod tidy
```

## Releasing a new version

Bump version in `package.json`.
Run `npm install`.
Update `CHANGELOG.md`.
`git commit ...`.
If the version in `package.json` is `1.5.1`, make a git tag that is `v1.5.1`,
with `git tag v1.5.1`.
h

## Updating provisioned dashboards

After updating a provisioned dashboard, make sure it's data source is set correctly.

## Installing a version before its official release

If you want to test a released, but unsigned plugin, follow this.

https://grafana.com/docs/grafana/latest/cli/#override-default-plugin-zip-url

```shell
grafana cli --pluginUrl https://github.com/wildmountainfarms/wild-graphql-datasource/releases/download/v0.0.6/retrodaredevil-wildgraphql-datasource-0.0.6.zip plugins install retrodaredevil-wildgraphql-datasource
```

If necessary, you can `grafana.ini` with

```ini
# NOTE: Only do this if absolutely necessary. Even unreleased versions of Wild GraphQL Datasource should not require this
[plugins]
allow_loading_unsigned_plugins = retrodaredevil-wildgraphql-datasource
```

## Building and Development

### Set up your system

* Web requirements
  * Install node using your preferred way:
    * https://github.com/tj/n
    * https://github.com/nvm-sh/nvm#installing-and-updating
* Backend requirements
  * Install go https://go.dev/doc/install
  * Install Mage https://magefile.org/
* Grafana requirements
  * Install Docker: https://docs.docker.com/engine/install/
    * `npm run server` command uses `docker compose` to bring up Grafana
  * Make sure your user is part of the `docker` group


### Setup build environment

Go setup
```shell
# This will install to your GOPATH, which is $HOME/go by default.
go get -u github.com/grafana/grafana-plugin-sdk-go
go mod tidy
mage -v
mage -l
```

Node setup:

```shell
npm install
```

### Recommended development environment

You may choose to use VS Code, which has free tools for developing TypeScript and Go code.
IntelliJ IDEA Ultimate is a non-free choice that has support for both TypeScript and Go code.
Alternatively, WebStorm (also not free) covers TypeScript development and GoLand covers Go development.

If you are using IntelliJ IDEA Ultimate, make sure go to "Language & Frameworks > Go Modules" and click "Enable go modules integration".

If you are using VS Code, this is a good read: https://github.com/golang/vscode-go/blob/master/docs/gopath.md

## To-Do

* Make annotation queries more intuitive
* Add support for secure variable data defined in the data source configuration
  * The variables defined here cannot be overridden for any request - this is for security
  
Lower priority To-Dos

* Customize default query (https://github.com/wildmountainfarms/wild-graphql-datasource/issues/1)
* Add support for variables: https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-support-for-variables#add-support-for-query-variables-to-your-data-source
* Add metrics to backend component: https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-logs-metrics-traces-for-backend-plugins#implement-metrics-in-your-plugin
* Support returning logs data: https://grafana.com/developers/plugin-tools/tutorials/build-a-logs-data-source-plugin
  * We could just add `"logs": true` to `plugin.json`, however we need to support the renaming of fields because sometimes the `body` or `timestamp` fields will be nested
* Create a GraphQL button panel (or a SolarThing app) that has a button panel that can be used to
  * If we create an app, we can follow https://github.com/RedisGrafana/grafana-redis-app
    * https://github.com/RedisGrafana/grafana-redis-app/blob/e093d18a021bb28ba7df3a54d7ad17c2d8e38f88/src/redis-gears-panel/components/RedisGearsPanel/RedisGearsPanel.tsx#L314
* Auto-populate the data path field by using `documentAST` to recognize the first path to an array
* Look into Apollo GraphQL
  * https://studio.apollographql.com/sandbox/explorer
  * https://www.apollographql.com/docs/graphos/explorer/
  * https://www.npmjs.com/package/@apollo/explorer
