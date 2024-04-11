# UDDNS
UDDNS - Universal (or Ultimate) Dynamic DNS updater

## How to use
### Obtaining the program
You can either:

`go install github.com/we11adam/uddns@latest`

or download the binary for you platform directly from the [releases page](https://github.com/we11adam.com/uddns/releases/)


### Configuration file
Ceate a `uddns.yaml` file in the same directory as the binary. The file should look like this:

```yaml
providers:
  routeros:
    endpoint: https://192.168.88.1 # RouterOS API endpoint
    username: admin # RouterOS user with API access
    password: "" # RouterOS user password

updaters:
  cloudflare:
    email: "user@exmaple.com" # Cloudflare account email
    apikey: fd25bdc03a4c17450aa4327aa37de4573270f # Cloudflare API key
    domain: ddns.yourdomain.com # Domain to update
```

Where:
- `providers` is a list of providers that UDDNS can use to obtain the current public IP address. Currently, `routeros` is supported.
- `updaters` is a list of updaters that UDDNS can use to update the DNS records. Currently, `cloudflare` is supported.


### Running
Run the binary as the following. It will update the DNS record with the current public IP address with a default interval of 30 seconds, which can be overriden with the `UDDNS_INTERVAL` environment variable. The format for specifying the interval is flexible, allowing values such as `60s`, `5m`, `1h`, etc.
```shell
nohup ./uddns &
```


## Roadmap
- [ ] Add more providers
- [ ] Add more updaters
- [ ] Add granular configuration options
- [ ] Add sensible logging
- [ ] Add tests
- [ ] Add CI/CD
- [ ] Add Dockerfile
- [ ] Add Daemon mode
- [ ] Add systemd service file



## Contributing
Pull requests are very welcome! For major changes, please open an issue first to discuss what you would like to change.

## License
MIT

