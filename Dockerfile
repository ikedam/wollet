FROM golang:1.23.7-bookworm AS build

WORKDIR /workspace
ADD . /workspace/
RUN go build -o wolbolt.cgi ./cmd/wolbolt-cgi

FROM httpd:2.4.63-bookworm

RUN echo "Include conf.d/*.conf" >> /usr/local/apache2/conf/httpd.conf
COPY docker/apache/wollet.conf /usr/local/apache2/conf.d/wollet.conf
COPY --from=build --chmod=755 /workspace/wolbolt.cgi /usr/local/apache2/htdocs/wollet-cgi/wolbolt.cgi
COPY docker/apache/wolbolt.yaml /usr/local/apache2/htdocs/wollet-cgi/wolbolt.yaml
