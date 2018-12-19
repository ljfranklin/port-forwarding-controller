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
[x] Make helm chart
[x] Add to ansible
[x] set `KUBEBUILDER_ASSETS` to some tmp dir so you don't need to install kubebuilder to run tests
[ ] Add docs
    - When to use NodePort vs LoadBalancer
    - Potential `annotations`:
      - `port-forwarding.lylefranklin.com/enable: true`, required
      - `port-forwarding.lylefranklin.com/unifi-site: SOME_SITE`, defaults to `default` site
    - install kubebuilder
      - https://book.kubebuilder.io/getting_started/installation_and_setup.html
    - generate skeleton
      - `kubebuilder create api --group core --version v1 --kind Service --controller=true --resource=false`

## Docker build

- Enable experimental docker CLI commands
  - add `"experimental": "enabled"` to `$HOME/.docker/config.json`
- Register QEMU:
  - `docker run --rm --privileged multiarch/qemu-user-static:register`
  - if you skip this step you'll get `standard_init_linux.go:190: exec user process caused "exec format error"`
