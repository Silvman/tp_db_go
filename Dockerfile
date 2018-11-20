FROM ubuntu:18.04

MAINTAINER Silvman

USER root

ENV PGVER 11
RUN apt update -y &&\
    apt install -y wget gnupg &&\
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - &&\
    echo "deb http://apt.postgresql.org/pub/repos/apt/ bionic-pgdg main 11" > /etc/apt/sources.list.d/pgdg.list &&\
    apt update -y  &&\
    apt install -y postgresql-$PGVER

RUN wget https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz &&\
    tar -xvf go1.11.2.linux-amd64.tar.gz &&\
    mv go /usr/local

ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR $GOPATH/src/github.com/Silvman/tech-db-forum/
ADD . $GOPATH/src/github.com/Silvman/tech-db-forum/

RUN go build -ldflags "-s -w" ./cmd/forum-server

EXPOSE 5000
EXPOSE 5432

# postgres settings
# postgres settings
RUN echo "include_dir = 'conf.d'" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    mv pg_hba.conf /etc/postgresql/$PGVER/main/ &&\
    mv forum.conf /etc/postgresql/$PGVER/main/conf.d/

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql -q docker -f base.sql &&\
    /etc/init.d/postgresql stop

CMD service postgresql start && ./forum-server