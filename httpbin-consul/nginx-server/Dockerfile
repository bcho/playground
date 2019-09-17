FROM hashicorp/consul-template:latest as consul-template

fROM nginx:1.17.3

COPY --from=consul-template /consul-template /bin/consul-template

RUN rm -f /etc/nginx/conf.d/*.conf

COPY entrypoint.sh /bin/entrypoint.sh
ENTRYPOINT /bin/entrypoint.sh
