-- @up

INSERT INTO users (name, last_name, age)
VALUES ('John', 'Doe', 99),
       ('Jane', 'Doe', 99);

-- @down

DELETE FROM users where last_name = 'Doe';
