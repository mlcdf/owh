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
NAME:
   owh - Deploy to OVHcloud Web Hosting

USAGE:
   owh [global options] command [command options] [arguments...]

VERSION:
   (devel)

COMMANDS:
   login              Login to your OVHcloud account
   hosting:list, hl   List all the hostings
   domains:list, dl   List attached domains
   domains:attach     Attach a domain
   domains:detach     Detach a domain
   deploy             Deploy the content of a folder to a site
   users:list         List ssh/ftp users
   users:changepw     Change ssh/ftp user password
   users:delete       Delete ssh/ftp users
   remove, rm         Remove a deployment (files & attached domains)
   tasks:list, tasks  List attached tasks
   whoami             Shows info about the user currently logged in
   help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d    enable verbose output (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## TODO

MVP:
- Add spinner on long task
- Add matching flags to interactive inputs
- Uniformize terminal output across all commands (color, style, stderr/stdout)
- Add tests
- Better handling of expired OVH API token

Nice to have:
- Stats command ?
- Logs command if possible ?
