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
  created     TIME DEFAULT CURRENT_TIME,
  forum       TEXT REFERENCES forums(slug),
  author      citext REFERENCES users(nickname),

  CONSTRAINT  slug_unique UNIQUE(slug)
);

create table posts (
  id          BIGSERIAL PRIMARY KEY,
  parent      BIGINT,
  parents     BIGINT[],
  message     TEXT,
  isEdit      BOOLEAN DEFAULT false ,
  forum       TEXT REFERENCES forums(slug),
  created     TIME DEFAULT CURRENT_TIME,
  thread      BIGINT REFERENCES threads(id),
  author      citext REFERENCES users(nickname)
);

create table votes (
  author      citext references users(nickname),
  thread      bigint references threads(id),
  vote        int
);

drop trigger if exists recount_votes_trigger on votes;
drop trigger if exists new_post_trigger on posts;
drop trigger if exists new_thread_trigger on threads;

create trigger recount_votes_trigger
  after insert or update on votes
  for each row execute procedure recount_votes();

create trigger new_post_trigger
  after insert on posts
  for each row execute procedure inc_counters();

create trigger new_thread_trigger
  after insert on threads
  for each row execute procedure inc_counters();

create or replace function recount_votes() returns trigger as $$
  begin
    update threads set votes = a.v from (select sum(vote) as v from votes where thread = NEW.thread group by thread) as a
    where threads.id = NEW.thread;
    return NEW;
  end;
$$ language plpgsql;

create or replace function inc_counters() returns trigger as $$
  begin
    if tg_name = 'new_post_trigger' then
      update forums set posts = posts + 1 where slug = NEW.forum;
    elseif tg_name = 'new_thread_trigger' then
      update forums set threads = threads + 1 where slug = NEW.forum;
    end if;
    return NEW;
  end;
$$ language plpgsql;


