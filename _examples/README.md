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

# Demonstration
## Access to a backend server
`https://backend1.local` and `https://backend2.local` are available.

## Static Files and Fallback
You can test static file hosting and fallback functionality:

1. Normal static file access
```
https://backend1.local/public/index.html   # Successfully displays index.html
https://backend1.local/public/example.html # Successfully displays example.html
```

2. Fallback demonstration
```
https://backend1.local/public/not-exist.html  # Access to a non-existent file
```
When accessing a non-existent file, it will redirect to the 404.html page specified in config.yaml (`fallback_path: /public/404.html`).
