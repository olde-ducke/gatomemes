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
  identity        VARCHAR(36) NOT NULL,
  user_name       VARCHAR(20) NOT NULL,
  password        VARCHAR(20) NOT NULL,
  session_key     VARCHAR(36),
  is_disabled     BOOL NOT NULL DEFAULT 0,
  is_admin        BOOL NOT NULL DEFAULT 0,
  is_root         BOOL NOT NULL DEFAULT 0,
  reg_time        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_seen_time  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY(identity),
  UNIQUE(user_name),
  UNIQUE(session_key)
);