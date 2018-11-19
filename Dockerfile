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
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf &&\
    echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "fsync = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "full_page_writes = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "autovacuum = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "shared_buffers = 256MB" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "wal_level = minimal" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "effective_cache_size = 1024MB" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "max_wal_senders = 0" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "work_mem = 16MB" >> /etc/postgresql/$PGVER/main/postgresql.conf

USER postgres
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f base.sql &&\
    /etc/init.d/postgresql stop

CMD service postgresql start && ./forum-server