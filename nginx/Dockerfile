FROM nginx:alpine

COPY nginx.conf /etc/nginx/nginx.conf
COPY web.crt /etc/ssl/certs/my-cert.crt
COPY web.key /etc/ssl/private/my-cert.key

EXPOSE 8443
