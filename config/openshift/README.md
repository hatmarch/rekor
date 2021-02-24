# Steps for running and debugging in OpenShift

## Cluster Setup

Apply all .yaml files in the [openshift](/workspaces/rekor/config/openshift) directory

```
oc apply -R -f ${DEMO_HOME}/config/openshift
```

Once up and running, you can reach the `rekor-server` at:

```
curl http://$(oc get route rekor-server -o jsonpath='{.spec.host}')/api/v1/log/ 
```

## Build Cli

From the root `${DEMO_HOME}` of the git repo, run the following command

```
make cli
mv cli ${DEMO_HOME}/rekor
```

## Signing a release

### Create key

NOTE: You cannot run this in a container as there is not enough entropy to generate the key

This is based on [these instructions](https://www.gnupg.org/gph/en/manual/c14.html).  If you already have a key-pair associated with gpg, then you can skip this step.

From the command line, run the following which will interactively guide you in creating a keypair that you can use to sign releases.  Retain the EMAIL and key PASSWORD as you will need these later

```
gpg --gen-key
```

This will create a key in a keyring in your ~/.gnupg directory.  This is be mounted by the devcontainer (see [devcontainer.json](../../.devcontainer/devcontainer.json))

### Sign release with key

Use this command to sign an artifact at `ARTIFACT_PATH`

```
gpg --armor -u ${EMAIL} --output mysignature.asc --detach-sig ${ARTIFACT_PATH}
```

After you enter your password to unlock your key, this will put mysignature.asc in the current directory

You will also need to export your public key

```
gpg --export --armor ${EMAIL} > mypublickey.key
```

Upload the artifact to a rekor accessible URL. For now s3 but maybe later a nexus repo hosted in OpenShift

Set BUCKET equal to your bucket (e.g. `mwh-demo-assets`)

```
aws s3 cp rekor s3://$(echo $BUCKET)/rekor --acl public-read
```

The artifact URL will then be:

```
ARTIFACT_URL=http://$(echo $BUCKET).s3.amazonaws.com/rekor
```

Then finally upload your entry to rekor

```
${DEMO_HOME}/rekor upload --rekor_server http://$(oc get route rekor-server -o jsonpath='{.spec.host}') --signature mysignature.asc --public-key mypublickey.key --artifact ${ARTIFACT_URL} --sha $(sha256sum rekor | awk '{ print $1 }')

```

## Run in DevContainer

### Debugging in VSCode

In one terminal `port-forward` to redis
```
oc port-forward svc/redis 6379:6379
```

In another terminal, `port-forward` to `trillian-log`:
```
oc port-forward svc/trillian-log 8090:8090
```

Open [main.go](/workspaces/rekor/cmd/server/main.go)

Run the debugger