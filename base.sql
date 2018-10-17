create extension if not exists citext;
create table users (
  nickname    citext PRIMARY KEY,
  fullname    TEXT,
  about       TEXT,
  email       TEXT NOT NULL,

  CONSTRAINT  email_unique UNIQUE(email)
);

create table forums (
  slug        TEXT PRIMARY KEY,
  title       TEXT,
  owner       citext REFERENCES users(nickname),
  posts       BIGINT DEFAULT 0,
  threads     BIGINT DEFAULT 0
);

create table threads (
  id          BIGSERIAL PRIMARY KEY,
  title       TEXT,
  message     TEXT,
  votes       BIGINT,
  slug        TEXT,
  created     TIME,
  forum       TEXT REFERENCES forums(slug),
  author      citext REFERENCES users(nickname),

  CONSTRAINT  slug_unique UNIQUE(slug)
);

create table posts (
  id          BIGSERIAL PRIMARY KEY,
  parent      BIGINT,
  message     TEXT,
  isEdit      BOOLEAN,
  forum       TEXT REFERENCES forums(slug),
  created     TIME,
  thread      BIGINT REFERENCES threads(id),
  author      citext REFERENCES users(nickname)
);

