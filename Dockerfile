FROM golang:1.15 as builder
ADD . /app
WORKDIR /app
RUN make

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/bin/graffitiwallpainter /usr/local/bin/graffitiwallpainter
ENTRYPOINT ["/usr/local/bin/graffitiwallpainter"]
