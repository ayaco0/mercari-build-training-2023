FROM node:16-alpine
WORKDIR /app

RUN addgroup -S mercari && adduser -S trainee -G mercari
COPY package*.json ./
COPY tsconfig.json tsconfig.json
COPY public public
COPY src src
RUN npm ci

RUN mkdir node_modules/.cache
RUN chown -R trainee:mercari node_modules/.cache
USER trainee

CMD ["npm", "start"]
