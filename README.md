# caddy-redirects

This plugins provides a middleware that redirects matching URL to another path.
The mapping between the matching URL and the dest path is stored in a PostgresSQL table.
The entire table is loaded in-memory during Caddy startup and can be reloaded using the endpoint
`http://caddy:2019/redirecter/reload`

The plugin expects a table named `redirects` with at least these fields:
```sql
CREATE TABLE redirects (src_url varchar(500), dst_path varchar(500));
```

This is an example redirect:
```sql
INSERT INTO redirects (src_url, dst_path)
VALUES ('https://www.vinissimus.com/blog/garnachas-de-culto/', '/es/garnacha')
```

We also need to add the `redirecter` directive in our Caddyfile to configure the pg:

```Caddyfile
{
    order redirecter first
}

www.vinissimus.com {
    redirecter {
        host "postgres-ip-addr"
        port 5432
        user "user"
        password "passw0rd"
        db_name "db-name"
    }

    # ...
}
```
