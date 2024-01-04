# Wild GraphQL Datasource

This is a Grafana datasource that aims to make requesting timeseries data via a GraphQL endpoint easy.

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

# NOTICE: This `rm` command will remove your existing go installation
rm -rf ~/go 
wget -c https://dl.google.com/go/go1.21.5.linux-amd64.tar.gz -O - | tar -xz -C ~/
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc

cd ~/Documents/Clones
git clone https://github.com/magefile/mage
cd mage
go run bootstrap.go

# *docker installation not shown*
# *adding $USER to docker group not shown*
```

### Setup build environment

```shell
go get -u github.com/grafana/grafana-plugin-sdk-go
go mod tidy
mage -v
mage -l

npm install
npm run dev
npm run build
```

### Test inside Grafana instance

Note that `npm run server` uses `docker compose` under the hood.

```shell
npm run server
```

### Recommended development environment

You may choose to use VS Code, which has free tools for developing TypeScript and Go code.
IntelliJ IDEA Ultimate is a non-free choice that has support for both TypeScript and Go code.
Alternatively, WebStorm (also not free) covers TypeScript development and GoLand covers Go development.

