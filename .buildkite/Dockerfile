FROM golang:1.16-alpine


RUN apk add --update --no-cache \
    make \
    git
    
WORKDIR /samples-go

RUN wget -O- https://raw.githubusercontent.com/fossas/fossa-cli/master/install.sh | sh

