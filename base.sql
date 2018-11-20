create extension if not exists citext;
create table if not exists users (
  nickname citext collate "ucs_basic" primary key,
  fullname text default '',
  about    text default '',
  email    citext not null,

  constraint email_unique unique (email)
);

cluster users using users_pkey;

create table if not exists forums (
  slug    citext primary key,
  title   text not null,
  owner   citext references users (nickname),
  posts   bigint default 0,
  threads bigint default 0
);

cluster forums using forums_pkey;

create table if not exists forums_users (
  forum    citext collate "ucs_basic",
  nickname citext collate "ucs_basic",
  fullname text default '',
  about    text default '',
  email    citext not null,

  primary key (forum, nickname)
);

cluster forums_users using forums_users_pkey;

create table if not exists threads (
  id      bigserial primary key,
  title   text not null,
  message text not null,
  votes   bigint                   default 0,
  slug    citext,
  created timestamp with time zone default current_timestamp,
  forum   citext,
  author  citext collate "ucs_basic" references users (nickname),

  constraint slug_unique unique (slug)
);

create index if not exists threads_forum_created_idx on threads (forum,created);
cluster threads using threads_forum_created_idx;

create table if not exists posts (
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

create table if not exists votes (
  author citext collate "ucs_basic" references users (nickname),
  thread bigint,
  vote   int default 1,

  constraint author_thread_unique unique (thread, author)
);

create index if not exists posts_thread_id_idx on posts (thread, id);
create index if not exists posts_thread_mpath_idx on posts (thread, mPath);
create index if not exists posts_rootparent_idx on posts (rootParent);

create or replace function establish_forum_users()
  returns trigger as $$
begin
  insert into forums_users (forum, nickname, fullname, about, email)
  select new.forum, nickname, fullname, about, email from users where nickname = new.author
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

drop trigger if exists recount_votes_trigger on votes;
create trigger recount_votes_trigger
  after insert or update
  on votes
  for each row execute procedure recount_votes();

drop trigger if exists new_thread_trigger on threads;
create trigger new_thread_trigger
  after insert
  on threads
  for each row execute procedure inc_counters();

drop trigger if exists establish_forum_users_threads_trigger on threads;
create trigger establish_forum_users_threads_trigger
  after insert
  on threads
  for each row execute procedure establish_forum_users();