FROM golang:1-alpine AS build

WORKDIR /build
COPY . /build/
ENV CGO_ENABLED=1

RUN apk add gcc musl-dev && go build -o e621 -ldflags="-s -w" .

FROM alpine:3.19
ENV E621_OUTPUT_DIRECTORY=/data/
COPY --from=build /build/e621 /bin/
CMD [ "/bin/e621" ]