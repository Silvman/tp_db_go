create extension if not exists citext;
create table users (
  nickname citext collate "ucs_basic" primary key,
  fullname text default '',
  about    text default '',
  email    citext not null,
  id       bigserial,

  constraint email_unique unique (email),
  constraint users_id_uniq unique (id)
);

create index on users (nickname, id);

create table forums (
  slug    citext primary key,
  title   text not null,
  owner   citext references users (nickname),
  posts   bigint default 0,
  threads bigint default 0
);

create table forums_users (
  forum citext collate "ucs_basic",
  uid   bigserial,

  primary key (forum, uid)
);
--
-- -- ?
-- create index on forums_users (uid);
-- -- create index on forums_users(forum);


create table threads (
  id      bigserial primary key,
  title   text not null,
  message text not null,
  votes   bigint                   default 0,
  slug    citext,
  created timestamp with time zone default current_timestamp,
  forum   citext references forums (slug),
  author  citext collate "ucs_basic" references users (nickname),

  constraint slug_unique unique (slug)
);

create table posts (
  id         bigserial primary key,

  parent     bigint                   default 0 ,
  rootParent bigint                   default 0 ,
  mPath      bigint []                not null  default '{}'::bigint[],
  message    text not null,
  isEdit     boolean                  default false,
  forum      citext collate "ucs_basic",
  created    timestamp with time zone default current_timestamp,

  thread     bigint default 0,
  author     citext collate "ucs_basic" references users (nickname)
);

create table votes (
  author citext collate "ucs_basic" references users (nickname),
  thread bigint,
  vote   int default 1,

  constraint author_thread_unique unique (thread, author)
);

create index on threads (created);
create index on threads (forum);
-- create index on threads (author);

-- create index on votes (thread);

-- create index on posts (parent);
create index on posts (thread, id);
create index on posts (thread, mPath);
-- create index on posts (author);
create index on posts (rootParent);

create or replace function establish_forum_users()
  returns trigger as $$
begin
  insert into forums_users (forum, uid)
  values (new.forum, (select id from users where nickname = new.author))
  on conflict do nothing;
  return new;
end;
$$
language plpgsql;

create or replace function recount_votes()
  returns trigger as $$
begin
  if (tg_op = 'INSERT')
  then
    update threads set votes = votes + new.vote where id = new.thread;
  elseif (tg_op = 'UPDATE')
    then
      update threads set votes = votes + (new.vote - old.vote) where id = new.thread;
  end if;
  return new;
end;
$$
language plpgsql;

create or replace function inc_counters()
  returns trigger as $$
begin
  if tg_name = 'new_post_trigger'
  then
    update forums set posts = posts + 1 where slug = new.forum;
  elseif tg_name = 'new_thread_trigger'
    then
      update forums set threads = threads + 1 where slug = new.forum;
  end if;
  return new;
end;
$$
language plpgsql;

create trigger recount_votes_trigger
  after insert or update
  on votes
  for each row execute procedure recount_votes();

create trigger new_thread_trigger
  after insert
  on threads
  for each row execute procedure inc_counters();

create trigger establish_forum_users_threads_trigger
  after insert
  on threads
  for each row execute procedure establish_forum_users();