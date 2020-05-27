FROM golang:1.14 as builder
ADD . /app
WORKDIR /app
RUN go build -o graffitiwallpainter

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y curl
COPY --from=builder /app/graffitiwallpainter /usr/local/bin/graffitiwallpainter
ENTRYPOINT ["/usr/local/bin/graffitiwallpainter"]
