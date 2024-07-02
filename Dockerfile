FROM golang:1.22.4 AS builder

WORKDIR /user/src/

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./proto/ ./proto/

RUN CGO_ENABLED=0 GOOS=linux go build -o /user/bin/ /user/src/cmd/...


FROM alpine:3.20.1 AS runner

ENV UID=1000 GID=1000 HOME=/user
ENV BIN=/user/bin PATH="${BIN}:${PATH}"
RUN addgroup -S -g "$GID" user \
 && adduser -S -u "$UID" -h "$HOME" -s /bin/sh user user \
 && mkdir -p "$BIN" \
 && chown -R user:user "$HOME"

RUN apk add --no-cache curl

WORKDIR /user

COPY --from=builder /user/bin/ /user/bin/

USER user
