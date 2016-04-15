For ducati-dns to work you will need to add:
```
properties:
  uaa:
    clients:
      ducati_dns:
        secret: some_ducati_dns_secret
```
to your `property_overides.yml` in `cf-release`
