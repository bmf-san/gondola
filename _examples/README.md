# _examples
This is a sample code for using gondola as a proxy server and a backend server to verify their operation.

# Get Started
## Edit a /etc/hosts
```sh
sudo vim /etc/hosts
```

```sh
# gondola
127.0.0.1 backend1.local
127.0.0.1 backend2.local
```

## Start a gondola
```sh
make create-cert
make up
```

## Access to a backend server
`https://backend1.local` and `https://backend2.local` are available.
