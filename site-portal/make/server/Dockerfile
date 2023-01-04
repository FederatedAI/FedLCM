FROM golang:1.19 as builder

ARG BRANCH
ARG COMMIT

WORKDIR /workspace

COPY . .
RUN make server

FROM photon:4.0
WORKDIR /
COPY --from=builder /workspace/output .

RUN tdnf install -y tzdata shadow >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 portal-server \
    && useradd --no-log-init -r -m -g 10000 -u 10000 portal-server
USER portal-server

VOLUME /var/lib/site-portal/data/uploaded
ENTRYPOINT ["/site-portal"]
