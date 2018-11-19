FROM ubuntu:18.10

MAINTAINER Silvman

USER root

RUN apt-get -y update

ENV PGVER 10
RUN apt-get install -y postgresql-$PGVER
RUN apt install -y wget

RUN wget https://dl.google.com/go/go1.9.2.linux-amd64.tar.gz
RUN tar -xvf go1.9.2.linux-amd64.tar.gz
RUN mv go /usr/local
ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR $GOPATH/src/github.com/Silvman/tech-db-forum/
ADD . $GOPATH/src/github.com/Silvman/tech-db-forum/

RUN go build -ldflags "-s -w" ./cmd/forum-server

EXPOSE 5000

USER postgres

RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "fsync = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "full_page_writes = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "autovacuum = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
#RUN echo "shared_buffers = 300MB" >> /etc/postgresql/$PGVER/main/postgresql.conf

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker docker &&\
    psql docker -f base.sql &&\
    /etc/init.d/postgresql stop

#VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root
EXPOSE 5432

CMD service postgresql start && ./forum-server