create table if not exists tasks
(
  id serial not null primary key,
  title varchar(50),
  status varchar(10) not null default 'new',
  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);