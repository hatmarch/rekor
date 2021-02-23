# Steps for running and debugging in OpenShift

## Cluster Setup

Apply all .yaml files in the [openshift](/workspaces/rekor/config/openshift) directory

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