DROP TABLE IF EXISTS gatomemes;
CREATE TABLE gatomemes (
  id        INT UNSIGNED AUTO_INCREMENT NOT NULL,
  line1     TEXT(255) NOT NULL,
  line2     TEXT(255) NOT NULL,
  FileName TEXT(255),
  PRIMARY KEY (id)
);

INSERT INTO gatomemes 
  (line1, line2)
VALUES 
  ('LINE1 TEXT1', 'LINE2 TEXT1'),
  ('LINE1 TEXT2', 'LINE2 TEXT2'),
  ('LINE1 TEXT3', 'LINE2 TEXT3');

DROP TABLE IF EXISTS user;
CREATE TABLE user (
  id            INT UNSIGNED AUTO_INCREMENT NOT NULL,
  user_name     VARCHAR(20) NOT NULL,
  password      TEXT(20) NOT NULL,
  is_disabled   BOOL NOT NULL DEFAULT 0,
  PRIMARY KEY(id),
  UNIQUE(user_name)
);

DROP TABLE IF EXISTS user_session;
CREATE TABLE user_session (
  session_key     VARCHAR(255) NOT NULL,
  user_id         INT UNSIGNED NOT NULL,
  login_time      DATETIME NOT NULL,
  last_seen_time  DATETIME NOT NULL,
  PRIMARY KEY     (session_key)
);