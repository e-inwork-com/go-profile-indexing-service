# Get Golang 1.19
FROM golang:1.19-alpine

# Get Go Profile Service
RUN go install github.com/e-inwork-com/go-profile-service/cmd@latest

# Expose port
EXPOSE 4001

# Run Go Profile Service
CMD ["cmd"]
