# Development



https://grafana.com/developers/plugin-tools/create-a-plugin/develop-a-plugin/best-practices

https://grafana.com/developers/plugin-tools/tutorials/build-a-data-source-backend-plugin


## Installing via grafana-cli

If you want to test a released, but unsigned plugin, follow this.

https://grafana.com/docs/grafana/latest/cli/#override-default-plugin-zip-url

```shell
grafana cli --pluginUrl https://github.com/wildmountainfarms/wild-graphql-datasource/releases/download/v0.0.2/retrodaredevil-wildgraphql-datasource-0.0.2.zip plugins install retrodaredevil-wildgraphql-datasource
```

Then update `grafana.ini` with

```ini
[plugins]
allow_loading_unsigned_plugins = retrodaredevil-wildgraphql-datasource
```

## Building and Development

### Setup your system

* Web requirements
  * Install nvm https://github.com/nvm-sh/nvm#installing-and-updating
* Backend requirements
  * Install go https://go.dev/doc/install
  * Install Mage https://magefile.org/
* Grafana requirements
  * Install Docker: https://docs.docker.com/engine/install/
    * `npm run server` command uses `docker compose` to bring up Grafana
  * Make sure your user is part of the `docker` group

An example of commands you *could* run.
Customize this setup to your liking.

```shell
# install nvm https://github.com/nvm-sh/nvm#installing-and-updating
nvm install 20

rm -rf /usr/local/go
wget -c https://dl.google.com/go/go1.21.5.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
# /usr/local/go is GOROOT $HOME/go is GOPATH, so add both bins to path
echo 'export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"' >> ~/.bashrc

cd ~/Documents/Clones
git clone https://github.com/magefile/mage
cd mage
# This will install to GOPATH, which is $HOME/go by default
go run bootstrap.go

# *docker installation not shown*
# *adding $USER to docker group not shown*
```

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


### Test inside Grafana instance

Note that `npm run server` uses `docker compose` under the hood.

```shell
npm run dev
mage -v build:linux
npm run server
```

### Recommended development environment

You may choose to use VS Code, which has free tools for developing TypeScript and Go code.
IntelliJ IDEA Ultimate is a non-free choice that has support for both TypeScript and Go code.
Alternatively, WebStorm (also not free) covers TypeScript development and GoLand covers Go development.

If you are using IntelliJ IDEA Ultimate, make sure go to "Language & Frameworks > Go Modules" and click "Enable go modules integration".

If you are using VS Code, this is a good read: https://github.com/golang/vscode-go/blob/master/docs/gopath.md


### Common Errors During Development

* `Watchpack Error (watcher): Error: ENOSPC: System limit for number of file watchers reached, watch`
  * https://stackoverflow.com/a/55543310/5434860

### Example repos

Some random examples of data source plugin source code

* https://github.com/grafana/grafana-infinity-datasource/tree/main/pkg
* https://github.com/cnosdb/grafana-datasource-plugin/blob/main/cnosdb/pkg/plugin/query_model.go
* https://github.com/grafana/grafana-plugin-examples/tree/main/examples/datasource-http-backend

## Dependency Notes

This section contains notes about dependencies.

* `graphql-ws` is not actually required by us, but this issue is unresolved so that's why we include it
  * https://github.com/graphql/graphiql/issues/2405#issuecomment-1469851608 (yes as of writing it says it's closed, but it's not)
  * It's not a bad thing that we include this dependency because it gives us a couple of types that we end up using


## To-Do

* Make data returned in the same order as the JSON specified it was
  * GoDS seems like a good solution, but there is a problem with its linked hash map deserializer: https://github.com/emirpasic/gods/issues/171
* Add ability to move parsing options up and down
* Add support for secure variable data defined in the data source configuration
  * The variables defined here cannot be overridden for any request - this is for security
  * Also add support for secure HTTP headers
* See what minimum Grafana version we can support
* Add support for variables: https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-support-for-variables
* Add metrics to backend component: https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-logs-metrics-traces-for-backend-plugins#implement-metrics-in-your-plugin
* Support returning logs data: https://grafana.com/developers/plugin-tools/tutorials/build-a-logs-data-source-plugin
  * We could just add `"logs": true` to `plugin.json`, however we need to support the renaming of fields because sometimes the `body` or `timestamp` fields will be nested
* Publish as a plugin
  * https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin
  * https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin#generate-an-access-policy-token
  * https://grafana.com/legal/plugins/
  * https://grafana.com/developers/plugin-tools/publish-a-plugin/provide-test-environment
* Publish on a private plugin repository
  * https://volkovlabs.io/blog/installing-grafana-plugins-from-a-private-repository-805b54a1add3/
* Create a GraphQL button panel (or a SolarThing app) that has a button panel that can be used to
  * If we create an app, we can follow https://github.com/RedisGrafana/grafana-redis-app
    * https://github.com/RedisGrafana/grafana-redis-app/blob/e093d18a021bb28ba7df3a54d7ad17c2d8e38f88/src/redis-gears-panel/components/RedisGearsPanel/RedisGearsPanel.tsx#L314
* Auto-populate the data path field by using `documentAST` to recognize the first path to an array
