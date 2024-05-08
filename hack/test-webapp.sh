#!/usr/bin/env sh

URL=$1

curl -fsS -o /dev/null "${URL}/www/about.html"
curl -fsS -o /dev/null "${URL}/www/catalog.html"
curl -fsS -o /dev/null "${URL}/www/leaderboard.html?metric=training+loss"
curl -fsS -o /dev/null "${URL}/www/bottle.html?digest=sha256:$(sha256sum testdata/bottle/bottle2.json | cut -d' ' -f1)"
curl -fsS -o /dev/null "${URL}/www/artifact/sha256:$(sha256sum testdata/bottle/bottle1.json | cut -d' ' -f1)/foo/bar/sample.txt"
