FROM node:0.12

ENV APP_DIR=/opt/app
RUN mkdir -p ${APP_DIR}
WORKDIR ${APP_DIR}

COPY package.json ${APP_DIR}/
RUN npm install --production
COPY . ${APP_DIR}
RUN mkdir -p ${APP_DIR}/assets
RUN ./node_modules/.bin/browserify src/app.js -o assets/app.js -t [ babelify ]
ENV PORT=3457
EXPOSE ${PORT}
CMD [ "npm", "start" ]
