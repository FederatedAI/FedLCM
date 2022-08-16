FROM node:15.4.0 as builder

WORKDIR /build_dir

COPY . .
RUN cd frontend && npm install && cd .. && make frontend

FROM photon:4.0

RUN tdnf install -y nginx shadow gettext >> /dev/null \
    && tdnf clean all \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log

COPY --from=builder /build_dir/output/frontend /usr/share/nginx/html
COPY --from=builder /build_dir/make/frontend/nginx.conf.template /etc/nginx/conf.d/nginx.conf.template

RUN groupadd -r -g 10000 nginx && useradd --no-log-init -r -g 10000 -u 10000 nginx \
    && chown -R nginx:nginx /etc/nginx \
    && chown -R nginx:nginx /usr/share/nginx/html

VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

USER nginx
CMD ["/bin/bash", "-c", "SITEPORTAL_SERVER_HOST=${SITEPORTAL_SERVER_HOST:-server} envsubst '${SITEPORTAL_SERVER_HOST}' < /etc/nginx/conf.d/nginx.conf.template > /etc/nginx/nginx.conf && exec nginx -g 'daemon off;'"]