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
   config            Show the owh configuration file location
   deploy            Deploy websites from a directory
   domains           List domains attached to a hosting
   domains:attach    Attach a domain
   domains:detach    Detach a domain
   hostings          List all your hostings
   info              Show info about the linked website
   link              Link current directory to an existing website on OVHcloud
   login             Login to your OVHcloud account
   open              Open browser to current deployed website
   remove, rm        Remove websites (files & attached domains)
   tasks             List tasks
   users             List ssh/ftp users
   users:changepass  Change ssh/ftp users password
   users:delete      Delete ssh/ftp users
   whoami            Show info about the user currently logged in

GLOBAL OPTIONS:
   --debug, -d    enable verbose output (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## TODO

MVP:
- Add matching flags to interactive inputs
- Uniformize terminal output across all commands (color, style, stderr/stdout)
- Add tests
- Better handling of expired OVH API token

Nice to have:
- Stats command ?
- Logs command if possible ?
