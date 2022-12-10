# owh

An opinionated command line interface to OVHcloud Web Hosting. The goal is to
provide a similar experience to Netlify or Vercel cli to deploy static (or
PHP) websites.

## Things you should note

- It requires at least a Pro plan (for SSH access).
- The underlying file system is made invisible: deploying a website with a domain www.example.com will upload the content to a www.example.com folder. This is by design and it can't be overridden.
- A few operations are asynchronous and create a task on OVHcloud infrastructure that will be picked up by robots and executed. Therefore, some operations may take several seconds or more.
- When attaching a domain to a hosting, you'll have to wait ~1h for the Let's Encrypt SSL certificates.

# Usage

```
Usage: owh [--version] [--help] <command> [<args>]

Deploy websites to OVHcloud Web Hosting.

Available commands are:
    deploy      Deploy websites from a directory
    domains     Handle various domain operations
    hostings    List all your hostings
    info        Show info about the linked website
    link        Link current directory to an existing website on OVHcloud
    login       Login to your OVHcloud account
    logs        View access logs
    open        Open browser to current deployed website
    remove      Remove websites (files & attached domains)
    tasks       Lists tasks
    tool        Group useful extra-commands
    users       Manage users
    whoami      Show info about the user currently logged in
```

## Development

Requirements:
- go version > 1.19+
- docker

Run the app

```sh
go run main.go
```

Run the tests

```sh
go test ./...
```

Force `go test` to run all the tests (by disabling caching)
```sh
./scrits/test.sh
```

## Tools

- check domains
- ci

## License

[MIT](https://choosealicense.com/licenses/mit/)
