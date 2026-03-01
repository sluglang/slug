-- @up

CREATE TABLE users
(
    id        INTEGER PRIMARY KEY,
    name      TEXT,
    last_name TEXT,
    age       INTEGER
);

-- @down

drop table users;
