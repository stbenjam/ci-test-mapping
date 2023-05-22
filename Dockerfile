FROM registry.access.redhat.com/ubi9/ubi:latest AS builder
WORKDIR /go/src/openshift-eng/ci-test-mapping
COPY . .
ENV PATH="/go/bin:${PATH}"
ENV GOPATH="/go"
RUN dnf install -y \
        git \
        go \
        make && make build

FROM registry.access.redhat.com/ubi9/ubi:latest AS base
COPY --from=builder /go/src/openshift-eng/ci-test-mapping/ci-test-mapping /ci-test-mapping
ENTRYPOINT ["/ci-test-mapping"]
