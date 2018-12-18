## TODO

[x] Handle updates properly
    - e.g. keep name the same but change port
[x] add port number to name
    - avoid name collisions
[x] Docker stuff
[x] Deploy
[x] `port-forwarding.lylefranklin.com/unifi_site: SOME_SITE`
[x] Reduce number of API calls
    - e.g. try to make call, login if necessary
[x] Refactor
[ ] Make helm chart
[ ] Add to ansible
[ ] Add docs

## Docker build

- Enable experimental docker CLI commands
  - add `"experimental": "enabled"` to `$HOME/.docker/config.json`
- Register QEMU:
  - `docker run --rm --privileged multiarch/qemu-user-static:register`
  - if you skip this step you'll get `standard_init_linux.go:190: exec user process caused "exec format error"`

## Workflow

- install kubebuilder
  - https://book.kubebuilder.io/getting_started/installation_and_setup.html
- generate skeleton
  - `kubebuilder create api --group core --version v1 --kind Service --controller=true --resource=false`
- TODO: set `KUBEBUILDER_ASSETS` to some tmp dir so you don't need to install kubebuilder to run tests
- trigger on new svcs of type LoadBalancer, svc update, svc delete and
  on periodic timer
  - timer ensures rules removed out-of-band are re-added
- To login:
  - attempt some API call
  - if response is `meta.msg: api.err.LoginRequired` then client hits /api/login
  - store cookie in memory, attach to each request
- For each svc:
  - if it is of type LoadBalancer or NodePort
    - List existing rules in unifi, add rule if missing
      - Rule name should be `$DEPLOYMENT-$PORT`
      - Use `LoadBalancer Ingress` field for forward IP or `External IP` for NodePort
- For each forwarding rule in list:
  - if rule is in form `$DEPLOYMENT-\d+`
    - delete rule if no longer needed
- Potential `annotations`:
  - `port-forwarding.lylefranklin.com/enable: true`, required
  - `port-forwarding.lylefranklin.com/unifi-site: SOME_SITE`, defaults to `default` site
- Docs
  - When to use NodePort vs LoadBalancer

## API calls

- login:
  - `curl --cookie /tmp/cookie --cookie-jar /tmp/cookie https://unifi.batcomputer.io/api/login -d '{"username": "$USERNAME", "password": "$PASSWORD"}'`

```
{ "data" : [ ] , "meta" : { "rc" : "ok"}}
```

- get sites:
  - `/api/self/sites`

```
{ "data" : [ { "_id" : "5bd85ec40889ae0019308fbe" , "attr_hidden_id" : "default" , "attr_no_delete" : true , "desc" : "Default" , "name" : "default" , "role" : "admin"}] , "meta" : { "rc" : "ok"}}
```

- get forwarding rules:
  - `/api/s/default/rest/portforward`

```
{
  "data": [
    {
      "_id": "5bd919f20889ae0019309113",
      "dst_port": "443",
      "enabled": true,
      "fwd": "192.168.1.51",
      "fwd_port": "443",
      "name": "https",
      "proto": "tcp_udp",
      "site_id": "5bd85ec40889ae0019308fbe",
      "src": "any"
    },
    {
      "_id": "5bd91a040889ae0019309114",
      "dst_port": "80",
      "enabled": true,
      "fwd": "192.168.1.51",
      "fwd_port": "80",
      "name": "http",
      "proto": "tcp_udp",
      "site_id": "5bd85ec40889ae0019308fbe",
      "src": "any"
    },
    {
      "_id": "5bd91a2c0889ae0019309115",
      "dst_port": "1194",
      "enabled": true,
      "fwd": "192.168.1.50",
      "fwd_port": "1194",
      "name": "vpn",
      "proto": "tcp_udp",
      "site_id": "5bd85ec40889ae0019308fbe",
      "src": "any"
    }
  ],
  "meta": {
    "rc": "ok"
  }
}
```

- Add rule:
  - `-X POST -d '{ "dst_port":	1010, "enabled":	true, "fwd": "192.168.1.100", "fwd_port": 1010, "name": "test", "proto": "tcp_udp", "src": "any" }' /api/s/default/rest/portforward`
  - Note: there are no enforced unique constraints so duplicate rules may get created on race condition

- Delete rule:
  - `-X DELETE /api/s/default/rest/portforward/RULE_ID`
