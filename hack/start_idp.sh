#!/usr/bin/env bash

docker rm idp-mysql
docker run --name idp-mysql -p 3306:3306 -v $(pwd)/../ansible/roles/mysql/files:/docker-entrypoint-initdb.d \
-e MYSQL_ROOT_PASSWORD=foo  -e MYSQL_DATABASE=imovies -e MYSQL_USER=user -e MYSQL_PASSWORD=pass mysql

cd ./../IdP/
go run . -dsn="user:pass@(localhost)/imovies" -admin-url="https://hydra.fadalax.tech:9001"