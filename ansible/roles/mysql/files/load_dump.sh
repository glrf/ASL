#!/bin/bash

mysql -uroot -p"$MYSQL_ROOT_PASSWORD" imovies < /docker-entrypoint-initdb.d/imovies_users.dump

