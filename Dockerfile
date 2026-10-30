FROM --platform=$BUILDPLATFORM golang:1.27.3@sha256:957861185dbf29d93668ace41e011d091c6e779953d47f0316ef2cc58e28fe2b

ARG MAIN_GO_FILE
ENV MAIN_GO_FILE=$MAIN_GO_FILE
ARG BIN_NAME
ENV BIN_NAME=$BIN_NAME

WORKDIR /src

# The modules first, so that a change to the source alone keeps the layer they are downloaded in.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["sh", "-c", "go tool air -c .air.${BIN_NAME}.toml"]
