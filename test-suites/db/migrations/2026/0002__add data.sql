-- @up

INSERT INTO users (name, last_name, age)
VALUES ('Alice', 'Smith', 30),
       ('Bob', 'Jones', 25);

-- @down

DELETE FROM users;
