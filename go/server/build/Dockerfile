FROM golang:1.19-buster AS build

WORKDIR ../

COPY ./go.mod go.sum ./

COPY . .

RUN go mod download
RUN go build -o main .

EXPOSE 5432 8080

ENTRYPOINT [ "./main", "-config=build/config.yaml" ]