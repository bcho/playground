#!/bin/bash

echo "$CONSUL_URL"
service nginx start &
consul-template \
    -consul-addr="$CONSUL_URL" \
    -template="$CONSUL_TEMPLATE:$CONSUL_ACTION"
