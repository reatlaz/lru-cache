FROM golang:alpine

COPY . /project

WORKDIR /project

# RUN file="$(ls -1 /project/)" && echo $file
# RUN echo $(ls -1 /)

RUN go mod vendor && go build -mod=vendor -o bin/lru ./cmd/lru

ENTRYPOINT ["/project/bin/lru"]
