# ACE Data Telemetry

This telemetry server implements a longitudinal tracking system for ACE Data Bottles.  When `ace-dt bottle push` or `ace-dt bottle pull` is run it pushed the bottle (metadata and data) to the OCI registry that is provided.  In addition those commands also push the metadata to 0 or more telemetry servers configured in ace-dt's configuration file.  In this way the telemetry server can be used to discover and track data bottles.

The server has a REST API defined [here](./internal/api/api.go) and a web interface defined [here](./internal/webapp/webapp.go).  The web interface has a data catalog, leaderboard, detail bottle view, and detail artifact view.

The server currently supports SQLite or Postgres as the database but other RDMS can be added easily.  The design and architecture are discussed in detail in the [writeup](./docs/writeup.md).

## Setup

The below is the prerequisite for installing ACE Telemetry application to your local machine:

### Install Dependencies

Install `jq` and `sqlite`:

```shell
sudo apt install build-essential jq sqlite3
```

Install the GO toolchain by any option:

- [Official GO toolchain](https://go.dev/doc/install)
- `snap install go --classic`
- `brew install go`

### SQLite Extension for VSCode (Optional)

Install sqlite extension in VSCode with `code --install-extension alexcvzz.vscode-sqlite`.  If this does not work, set the Sqlite extension settings to DEBUG

### Build and Run

Build and run with `make run` and direct your browser to <http://localhost:8100/> to view the web UI.

## Testing

In a shell call `make reload` to load data into the sqlite database.  Then run `make run` to run the server.  This will stand up a simple sqlite DB using the file `test.db`.

Alternatively you can upload test data into a running server with `make upload`.

Using the [VSCode REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) you can open the [test.http](testdata/test.http) to hit the REST API.

You can select your database by creating a `config.yaml` file at the top level of the project, with the contents

```yaml
apiVersion: config.telemetry.act3-ace.io/v1alpha1
kind: ServerConfiguration

db:
  dsn: file:test.db
  # dsn: "postgres://tester:myPassword@localhost/test"
  prometheus: no # disable annoying prometheus logging
```

You might also want to link telemetry to ACE Hub by adding the following to your configuration file.

```yaml
webapp:
  acehubs:
  - name: Lynx
    url:  https://hub.lion.act3-ace.ai
  - name: GCP
    url:  https://hub.ace.afresearchlab.com

  viewers:
  - name: "VS Code"
    accept: "image/*,application/json,text/plain;q=0.5, application/vnd.act3-ace.bottle;q=0.9"
    acehub:
      image: reg.git.act3-ace.com/ace/hub/vscode-server:v0
      resources:
        cpu: "2"
        memory: "2Gi"
      proxyType: normal
      jupyter: false
```

The client configuration file, which is often named `client-config.yaml`, looks like:

```yaml
apiVersion: config.telemetry.act3-ace.io/v1alpha1
kind: ClientConfiguration

locations:
  - name: Telemetry Server
    url: https://telemetry.example.com
    cookies:
      jwt: redacted
  - name: localhost
    url: http://localhost:8100
    cookies:
      foo: bar
```

The configuration API docs are [here](./docs/apis/config.telemetry.act3-ace.io/v1alpha1.md).

## Development on ACE Hub

We cannot use "sqlite browser" on ACE Hub since it is a desktop application.  As an alternative you can use the SQLite VSCode extension.

Run `make hub` to build and push the image.

Then launch in ACE Hub using [zot.lion](https://hub.lion.act3-ace.ai/environments/0?replicas=1&image=zot.lion.act3-ace.ai/ace/data/telemetry/hub:latest&hubName=telemetry2&proxyType=normal&resources[cpu]=4&resources[memory]=8Gi&shm=64Mi&env[GIT_CLONE_URL]=https://gitlab.com/act3-ai/asce/data/telemetry.git) or [reg.git](https://hub.lion.act3-ace.ai/environments/0?replicas=1&image=reg.git.act3-ace.com/ace/data/telemetry/hub:latest&hubName=telemetry2&proxyType=normal&resources[cpu]=4&resources[memory]=8Gi&shm=64Mi&env[GIT_CLONE_URL]=https://gitlab.com/act3-ai/asce/data/telemetry.git). <!-- markdownlint-disable-line MD034 -->

Make sure you set your NETRC envfile to get access to GIT.  Also set the default repo for skaffold with `skaffold config set default-repo zot.lion.act3-ace.ai`.  Make sure your KUBECONFIG is in the environment and set to the correct namespace.  Ensure helm has the bitnami repo with `helm repo add bitnami https://charts.bitnami.com/bitnami`. Then you can run `skaffold dev -p hub` and it will build  in the a pod running kaniko and then deploy with helm.

Then clone this repo (e.g., `https://gitlab.com/act3-ai/asce/data/telemetry.git`) from the left sidebar.  It will prompt for a username and a password (e.g., personal access token) unless you have NETRC set.

## PostgreSQL Setup Via Podman

You can use SQLite or PostgreSQL. If using PostgreSQL, you can develop locally with a container:

```bash
podman run \
  --name pg\
  -p 5432:5432 \
  -d \
  -e POSTGRES_PASSWORD=password \
  docker.io/library/postgres
podman exec \
  -it \
  pg \
  /usr/bin/psql -U postgres -c 'CREATE DATABASE test'
```

If you want a console to the database:

```bash
podman exec -it pg /usr/bin/psql -U postgres -d test
```

Drop/re-create:

```bash
podman restart pg && \
podman exec -it pg /usr/bin/psql -U postgres -c 'DROP DATABASE test' && \
podman exec -it pg /usr/bin/psql -U postgres -c 'CREATE DATABASE test'
```

Here is a [cheat sheet](https://tomcam.github.io/postgres/) for `psql`.

To run pgAdmin:

```bash
podman run \
  -d \
  --name pgadmin \
  -e 'PGADMIN_DEFAULT_EMAIL=user@domain.com' \
  -e 'PGADMIN_DEFAULT_PASSWORD=password' \
  -e 'PGADMIN_LISTEN_PORT=8080' \
  --net=host \
  dpage/pgadmin4
```

Then go to `localhost:8080` in your browser, log in with `user@domain.com` and `password`, and add data source `localhost` with username `postgres` and password `password` (on the `connection` tab). Check `save password`.

## Postgres Setup for Ubuntu (local or ACE Hub)

```shell
curl https://www.pgadmin.org/static/packages_pgadmin_org.pub | sudo apt-key add
sudo echo "deb https://ftp.postgresql.org/pub/pgadmin/pgadmin4/apt/$(lsb_release -cs) pgadmin4 main" > /etc/apt/sources.list.d/pgadmin4.list
sudo apt install postgresql pgadmin4-web
```

Add a user/role for yourself with `sudo -u postgres createuser --interactive` so you can use `psql` without `sudo`.
Run `createdb test` to create a database.  Then you can use `psql` without specifying the database name ($USER is the default).

Run `psql -c "CREATE USER tester WITH PASSWORD 'myPassword'"` to create a new user called test.

Then run the telemetry server with `go run ./cmd/telemetry serve --dsn 'postgres://tester:myPassword@localhost/test'`.  To re-create the database use `dropdb test; createdb test`.

## Build With Kaniko

```shell
docker run --rm -it -v $PWD:/workspace gcr.io/kaniko-project/executor:v1.6.0 --no-push --dockerfile /workspace/Dockerfile --context dir:///workspace/

docker run --rm -it -v $PWD:/workspace gcr.io/kaniko-project/executor:v1.6.0 --no-push --dockerfile /workspace/.acehub/Dockerfile --context dir:///workspace/.acehub

# does not work
docker run --rm -it -v $PWD:/workspace gcr.io/kaniko-project/executor:v1.6.0 --no-push --dockerfile /workspace/.acehub/Dockerfile --context dir:///workspace/ --context-sub-path .acehub/
```

## Mirror Telemetry Data

Mirroring the telemetry data is supported by the `telemtry` executable itself.  For example to pull down data from the telemetry server on Lynx and upload it to a local instance of telemetry use the following commands.

```shell
make download DOWNLOAD_DIR=testdata-download URL=https://telemetry.lion.act3-ace.ai
make upload UPLOAD_DIR=testdata-download
```

The downloads come in batches.
The download is incremental as it downloads in cronological order (saves the latest timestamp to file) and you can specify the oldest time you are interested in when downloading on the command line.

## Install Telemetry as a Tool

Simply run `go install gitlab.com/act3-ai/asce/data/telemetry/cmd/telemetry@latest`, then test it with `telemetry -h`.

To uninstall run `go clean -i gitlab.com/act3-ai/asce/data/telemetry`.

## Pretty Logs

Telemetry writes all logs as JSONL.  They are human readable but not easy to quickly scan.  To improve this we post process the logs with ad `jq` filter called [logs.jq](./log.jq).  `make run` does this to aid developers running it locally.

To make it easier for an operations team member to view logs we have a kubectl plugin called `kubectl-plogs` implemented as a simple [script](./kubectl-plogs) that needs to go on your PATH.  Then you can run `kubectl plogs my-pod-name` to view pretty logs.  The `jq` filter is encoded in the pod's and deployment's annotations to allow pretty logs to work without the source code.  There is also a `k9s` plugin for [plogs](./k9s-plugin.yaml).

## Support

<!-- - **[Troubleshooting FAQ](docs/troubleshooting-faq.md)**: consult list of frequently asked questions and their answers. -->
- **Mattermost [channel](https://chat.git.act3-ace.com/act3/channels/devops)**: create a post in the Telemetry channel for assistance.
- **[Web Browser](https://gitlab.com/act3-ai/asce/data/telemetry/-/issues)**: create an issue in GitLab.
<!-- TODO reinstate once operational - **Create a GitLab issue by [email](mailto:incoming+ace-data-telemetry-518-issue-@mail.act3-ace.com)** -->
