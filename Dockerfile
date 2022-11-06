FROM golang:1.19.2 AS build-stage

RUN mkdir /app
WORKDIR /app
COPY . .
RUN ./scripts/release.sh

FROM scratch AS export-stage
COPY --from=build-stage /app/dist/ /