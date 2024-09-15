# CDNS

An open-source lightweight DNS server that switches to ACME DNS challenge.

# Usage

## Run Container

If you have already run the Smallstep CA container locally, just run a CDNS container using the following command:

```shell
docker run -d --name cdns \
  --restart always \
  --env CDNS_ENV=production \
  --env CDNS_LOGGING_LEVEL=INFO \
  --env CDNS_DNS_LISTEN=0.0.0.0:53 \
  --env CDNS_DNS_PROTOCOL=both \
  --env CDNS_HTTP_TLS_MODE=acme \
  --env CDNS_HTTP_DOMAIN=dns.svc.dev \
  --env CDNS_HTTP_LISTEN=0.0.0.0:443 \
  --env CDNS_PROVIDERS_ACME_EMAIL=george@betterde.com \
  --env CDNS_PROVIDERS_ACME_SERVER=https://ca.svc.dev/acme/acme/directory \
  --env CDNS_PROVIDERS_ACME_STORAGE=/etc/cdns/certs \
  betterde/cdns:latest serve
```

## Docker Compose

In this example, I use Smallstep CA as the ACME Challenge Provider, To learn more about Smallstep CA, visit their [official website](https://smallstep.com/).

```shell
wget -O docker-compose.yaml https://raw.githubusercontent.com/betterde/cdns/master/docker-compose.yaml
docker compose up -d step
docker exec -it -u root step-ca bash -c "step certificate install -all /home/step/certs/root_ca.crt"
docker compose up -d
```

# License

This library is licensed under MIT Full license text is available in [LICENSE](LICENSE).