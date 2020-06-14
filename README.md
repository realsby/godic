godic [![Build Status](https://travis-ci.org/ottotech/godic.svg?branch=master)](https://travis-ci.org/ottotech/godic)
=========

## Overview

**godic** is a web application written in Go that helps you create and maintain a [data dictionary](https://en.wikipedia.org/wiki/Data_dictionary) of your relational database. <br> Currently it supports both mysql and postgres databases (latest versions). 

## How to use?

## With Docker
Pull [docker image](https://hub.docker.com/repository/docker/ottosg/godic):
```
docker pull ottosg/godic
```

### Environment variables:

```GODIC_SERVER_PORT```

This environment variable is not required. 
If not defined godic will use a default port 8080. This environment variable sets 
the port that godic will use for the http server to serve the UI.

```GODIC_DB_USER```

This environment variable **is required**. It represents the user of your database
that godic will scan. **Remember** that this user should have already the required
privileges to access the information schema of your database. Providing the correct
permissions to the given user is key for the godic app to work.  

```GODIC_DB_PASSWORD```

This environment variable is **required**. It represents the password of the given user to login into your database. 

```GODIC_DB_HOST```

This environment variable is **required**. It represents the host where your database is being served.

```GODIC_DB_PORT```

This environment variable is **required**. It represents the port on the host where your database is
being served.

```GODIC_DB_NAME```

This environment variable is **required**. It represents name of your database, the one godic will scan.

```GODIC_DB_DRIVER```

This environment variable is **required**. It represents the driver that your database supports.
godic currently support two drivers: **mysql** and **postgres**.

```GODIC_DB_SCHEMA```

This environment variable is not required, but it is desirable to pass it when initializing a container
if the default value is not the one you use. This environment variable represents the specific schema that
you want to allow godic to check. If not given godic will use **public** as the default schema. 

```GODIC_FORCE_DELETE```

This environment variable is no required. This variable is needed only in cases where you want to delete
the data dictionary you have stored in order to start fresh. I recommend you to use this flag wisely.
It can be useful in cases where there is some data corruption of some sort or when you just want to switch 
to a new database and create a new data dictionary. 

### VOLUME mount point:

Use this mount point if you want to create a VOLUME to preserve your data dictionary information.

Volume point location:
```
VOLUME /go/src/github.com/ottotech/godic/data
```

See example below to check how you can use this mount point.


### Start a godic instance:

```
$ docker run -d \
    --name some-godic \
    -e GODIC_DB_USER=master \
    -e GODIC_DB_PASSWORD=secret \
    -e GODIC_DB_HOST=NAME_OF_CONTAINER_SERVING_DB \
    -e GODIC_DB_PORT=5432 \
    -e GODIC_DB_NAME=mydb \ 
    -e GODIC_DB_DRIVER=postgres \
    -v godic_mount:/go/src/github.com/ottotech/godic/data \
    --network godic_net \
    godic
```

**NOTE**

Usually you will use a godic container with other containers where you are hosting your database, in that case
you will need to add all containers under the same network to be able to use DNS resolution and talk to your 
container holding your database. Creating a docker network and attaching existing containers to the network
in docker is pretty straightforward.

For example:

```
docker network create godic_net
docker network connect godic_net db_container
``` 

If it happens that your are running your database directly in your local machine there are ways to allow a docker 
container access your local port, but this talk is out of scope here :) 

## With this repo
```
go get -d github.com/ottotech/godic
cd $GOPATH/src/github.com/ottotech/godic
go build . 
```

Now in order to run godic you have to use the same environment variables described above
for the docker use case, but as flags.

For example:

```
$ ./godic \
    GODIC_DB_USER=master \
    GODIC_DB_PASSWORD=secret \
    GODIC_DB_HOST=NAME_OF_CONTAINER_SERVING_DB \
    GODIC_DB_PORT=5432 \
    GODIC_DB_NAME=mydb \ 
    GODIC_DB_DRIVER=postgres
``` 

## TODO
- more tests
- UI can be improved.
- Would be nice to be able to support sqlite as well. 
- Would be nice to support multiple versions of mysql and postgres and not just the latest ones.

## Contributing
Check repository [godic](https://github.com/ottotech/godic)

Clone the repo and run:
```
go get -u github.com/jteeuwen/go-bindata/...
```

Then on the repo path run:
```
go-bindata -debug /assets
go run .
``` 

There you are, you can start contributing on the go code or the UI.

This project uses react for websites, see [link](https://reactjs.org/docs/add-react-to-a-website.html)

If you make any changes, run ```go fmt ./...``` before submitting a pull request.

## License

Copyright ©‎ 2020, [ottotech](https://ottotech.site/)

Released under MIT license, see [LICENSE](https://github.com/ottotech/godic/blob/master/LICENSE.md) for details.