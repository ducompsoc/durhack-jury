FROM node:lts-hydrogen AS client-builder
WORKDIR /client

ENV VITE_JURY_NAME=$VITE_JURY_NAME
ENV VITE_JURY_URL=$VITE_JURY_URL

CMD [ "yarn", "run", "docker" ]
