# Step 1: Build the app in image 'builder'
FROM node:13.1-alpine AS builder

WORKDIR /usr/src/app
COPY . .
RUN yarn && yarn prod

# Step 2: Grab a certificate from letsencrypt
FROM certbot/certbot:v0.40.1 AS certbot
COPY gandi.ini .
RUN chmod 600 gandi.ini
RUN pip install certbot-plugin-gandi
RUN certbot certonly -m letsencrypt.org@d4ve.email --agree-tos -n -a certbot-plugin-gandi:dns --certbot-plugin-gandi:dns-credentials gandi.ini -d fadalax.tech

# Step 3: Use build output from 'builder'
FROM nginx:stable-alpine
LABEL version="0.0.1"

COPY nginx.conf /etc/nginx/nginx.conf

WORKDIR /usr/share/nginx/html
COPY --from=certbot /etc/letsencrypt/live/fadalax.tech/fullchain.pem /root/certs/fadalax.tech/chain.pem
COPY --from=certbot /etc/letsencrypt/live/fadalax.tech/privkey.pem /root/certs/fadalax.tech/privkey.key 
COPY --from=builder /usr/src/app/dist/frontend/ .
