USE hordes_challenge;

SET @open = 0;
SET @request = 1;
SET @invite = 2;

SET @private = 1 << 2;

SET @noapi = 1 << 3;

SET @draft = 0 << 4;
SET @review = 1 << 4;
SET @invitation = 2 << 4;

INSERT INTO user VALUES(1251, "Farijo", "farijo", null);
INSERT INTO user VALUES(1252, "Rima", "rima", "https:\/\/myhordes.eu\/cdn\/avatars\/1252\/e545256c415fb25322c351946a1ec724.jpeg");
INSERT INTO user VALUES(1253, "Odium", "odium", "https:\/\/myhordes.eu\/cdn\/avatars\/1253\/c79b9cdd2bd6e6989d7c70478f9c335c.png");
INSERT INTO user VALUES(1254, "Loki", "loki", null);
INSERT INTO user VALUES(1255, "LaCreme", "lacreme", "https:\/\/myhordes.eu\/cdn\/avatars\/1255\/9ab5b46b542ef1f3e0687faa51a180b7.jpeg");
INSERT INTO user VALUES(1256, "kikess", "kikess", null);

INSERT INTO challenge(name, creator, flags, start_date, end_date) VALUES("ch1", 1251, 0, NULL, NULL);
INSERT INTO challenge(name, creator, flags, start_date, end_date) VALUES("ch1", 1251, 0, '2023-01-01', NULL);
INSERT INTO challenge(name, creator, flags, start_date, end_date) VALUES("ch2", 1251, @request | @invitation, '2023-01-01', NULL);
INSERT INTO challenge(name, creator, flags, start_date, end_date) VALUES("ch3", 1251, @request | @private | @invitation, '2023-01-01', NULL);
INSERT INTO challenge(name, creator, flags, start_date, end_date) VALUES("ch3", 1252, @request | @private | @invitation, '2023-01-01', NULL);

INSERT INTO participant VALUES(1252, 5);
INSERT INTO participant VALUES(1252, 2);
INSERT INTO participant VALUES(1252, 3);
INSERT INTO participant VALUES(1252, 4);
INSERT INTO participant VALUES(1253, 4);
INSERT INTO participant VALUES(1253, 2);
INSERT INTO participant VALUES(1253, 3);
INSERT INTO participant VALUES(1254, 4);
INSERT INTO participant VALUES(1254, 2);
INSERT INTO participant VALUES(1254, 3);
INSERT INTO participant VALUES(1255, 3);
INSERT INTO participant VALUES(1255, 2);

INSERT INTO validator VALUES(1251, 2, 0);
INSERT INTO validator VALUES(1251, 3, 0);
INSERT INTO validator VALUES(1251, 4, 0);
INSERT INTO validator VALUES(1255, 2, 0);
INSERT INTO validator VALUES(1255, 4, 0);

INSERT INTO goal(challenge, typ, entity) VALUES(2, 1, 0);
INSERT INTO goal(challenge, typ, entity) VALUES(2, 1, 3);
INSERT INTO goal(challenge, typ, entity) VALUES(2, 2, 3);
INSERT INTO goal(challenge, typ, entity) VALUES(3, 1, 0);
INSERT INTO goal(challenge, typ, entity) VALUES(3, 1, 3);
INSERT INTO goal(challenge, typ, entity) VALUES(3, 2, 3);
INSERT INTO goal(challenge, typ, entity) VALUES(3, 0, 40);
INSERT INTO goal(challenge, typ, entity) VALUES(4, 1, 0);
INSERT INTO goal(challenge, typ, entity) VALUES(4, 1, 3);
INSERT INTO goal(challenge, typ, entity) VALUES(4, 2, 3);

SET @t = UTC_TIMESTAMP();
SET @t2 = UTC_TIMESTAMP(2);

INSERT INTO milestone(user, dt) VALUES(1252, @t2);
INSERT INTO milestone(user, dt) VALUES(1252, @t);
INSERT INTO milestone(user, dt) VALUES(1253, @t2);
INSERT INTO milestone(user, dt) VALUES(1254, @t2);
INSERT INTO milestone(user, dt) VALUES(1255, @t2);

INSERT INTO success(user, goal, accomplished, amount) VALUES(1252, 4, @t2, 5);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1252, 4, @t, 1);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1253, 4, @t2, 2);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1254, 4, @t2, 3);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1255, 4, @t2, 4);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1252, 5, @t2, 6);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1253, 5, @t2, 2);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1254, 5, @t2, 3);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1253, 6, @t2, 0);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1254, 6, @t2, 1);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1252, 7, @t, 2);
INSERT INTO success(user, goal, accomplished, amount) VALUES(1252, 7, @t2, 26);