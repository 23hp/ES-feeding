FROM --platform=$BUILDPLATFORM golang:1.22.3-alpine AS builder

WORKDIR /app

#COPY go.mod go.sum ./
#RUN go mod download
COPY . .
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -o bin

# Final stage
FROM scratch

COPY --from=builder /app/bin /bin
CMD ["/bin"]