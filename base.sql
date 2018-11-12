create extension if not exists citext;
create table users (
  nickname    citext collate "ucs_basic" primary key,
  fullname    text default '',
  about       text default '',
  email       citext not null,

  constraint  email_unique unique(email)
);

create table forums (
  slug        citext primary key,
  title       text not null,
  owner       citext references users(nickname),
  posts       bigint default 0,
  threads     bigint default 0
);

create table threads (
  id          bigserial primary key,
  title       text not null,
  message     text not null,
  votes       bigint default 0,
  slug        citext,
  created     timestamp with time zone default current_timestamp,
  forum       citext references forums(slug),
  author      citext collate "ucs_basic" references users(nickname),

  constraint  slug_unique unique(slug)
);

create table posts (
  id          bigserial primary key,
  parent      bigint default 0,
--   rootParent  bigint default 0,
  mPath       bigint[],
  message     text not null,
  isEdit      boolean default false,
  forum       citext references forums(slug),
  created     timestamp with time zone default current_timestamp,
  thread      bigint references threads(id),
  author      citext collate "ucs_basic" references users(nickname)
);

create table votes (
  author      citext collate "ucs_basic" references users(nickname),
  thread      bigint references threads(id),
  vote        int default 1,

  constraint  author_thread_unique unique(author, thread)
);

create index on threads (created);
create index on votes (thread);
create index on posts (parent);
create index on posts (thread);
-- create index on posts (mPath);
-- create index on posts(rootParent);


create or replace function recount_votes() returns trigger as $$
begin
  update threads set votes = a.v from (select sum(vote) as v from votes where thread = new.thread group by thread) as a
  where threads.id = new.thread;
  return new;
end;
$$ language plpgsql;

create or replace function inc_counters() returns trigger as $$
begin
  if tg_name = 'new_post_trigger' then
    update forums set posts = posts + 1 where slug = new.forum;
  elseif tg_name = 'new_thread_trigger' then
    update forums set threads = threads + 1 where slug = new.forum;
  end if;
  return new;
end;
$$ language plpgsql;

create or replace function posts_build_path() returns trigger as $$
begin
  if new.parent = 0 then
    update posts set mPath = array_append('{}'::bigint[], id) where id = new.id;
  else
    update posts set mPath = array_append((select mPath from posts where id = new.parent), id) where id = new.id;
  end if;
  return new;
end;
$$ language plpgsql;

create trigger recount_votes_trigger
  after insert or update on votes
  for each row execute procedure recount_votes();

create trigger new_post_trigger
  after insert on posts
  for each row execute procedure inc_counters();

create trigger new_thread_trigger
  after insert on threads
  for each row execute procedure inc_counters();

create trigger posts_build_path_trigger
  after insert on posts
  for each row execute procedure posts_build_path();