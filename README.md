# purge-dockerhub

## What is it?

This is an auto-purger for organisations on DockerHub.  Images that have not been pulled or pushed for longer than
a specified amount of days are deleted; a number of images can be retained in the repo regardless.

## How to compile it?

You need to have a recent version of Go (1.11 or newer) installed.  Then simply:

```bash
go build
./purge-dockerhub --help
```

If you're on e.g. Mac and want to compile the binary for Linux, for example, just set the environment variable GOOS before
compiling:

```bash
GOOS=linux go build
```

Consult the Go documentation for valid values of GOOS and GOARCH.

## How to run it?

```
$ ./purge-dockerhub  --help
2021/05/24 16:13:48 purge-dockerhub version 1.0, https://github.com/fredex42/purge-dockerhub
Usage of ./purge-dockerhub:
  -credentials string
    	yaml-format file containing the credentials to log in to Docker Hub (default "docker-credentials.yaml")
  -keepcount int
    	Always keep this many images regardless of age (default 150)
  -keepdays int
    	Any item which has been pulled or pushed within this number of days will be kept. 0 disables the check.
  -org string
    	organisation
  -really-delete
    	Unless this flag is set, the app will only print what would be done and not actually delete anything
```

### Create a credentials file

You should supply the user credentials in a file in yaml format.  By default, a file called "docker-credentials" is
loaded from the current directory, but you can change this with the `-credentials` option.

The contents should be like this:
```yaml
username: {your-docker-username}
password: {your-docker-password}
```

Passing credentials in this way keeps them safe from prying eyes who might list out commandlines.  Obviously you
should keep this file safe, e.g. with an 0600 file mode or better still deleting it as soon as you are done with it.
It is intended for use with a Kubernetes secrets map.

### Decide what to purge

If there is no value for the `-org` option then the logged in username is assumed as the organisation name.
If there is no value for the `-keepdays` option then _only_ the most recent number of images in `-keepcount` will be kept.
If there is no value specified for `-keepcount` then the default value of 150 is used.  This is intentionally high so
please reduce it to whatever number you need.

### Test it

Run the app, e.g. `./purge-dockerhub -credentials /etc/secret/docker.yaml -keepcount 5 -keepdays 30 -org my-organisation` 
will delete everything from "my-organisation" that has not been pulled or pushed for 30 days or more, but leave the first
5 tags regardless.
Check the output to verify it is doing what you expect.  **Nothing will be deleted** until you specify `-really-delete` on
the commandline.

### Run it

Finally, once you _know_ that only the stuff you really want to be deleted is going to be deleted, run the same commandline
appending `-really-delete` onto the end.  Watch as those old images disappear.