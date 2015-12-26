
--CREATE DOMAIN uid NOT NULL CHECK(VALUE ~ '^[a-fA-F0-9]{32}$');

CREATE TABLE users(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  email TEXT NOT NULL UNIQUE CHECK(email ~ '^[a-zA-Z0-9_%+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9][a-zA-Z0-9]+$'),
  password VARCHAR(128) NOT NULL
);

CREATE UNIQUE INDEX idx_users_email ON users(lower(email));

CREATE TABLE access_tokens(
  secret VARCHAR(256) PRIMARY KEY,
  user_id uuid REFERENCES users(id) NOT NULL,
  expires TIMESTAMP,
  access INT NOT NULL
);
CREATE INDEX idx_access_tokens_user_id ON access_tokens(user_id);
CREATE INDEX idx_access_tokens_expires ON access_tokens(expires) WHERE expires IS NOT NULL; -- index to use for deleting expired tokenss

CREATE TABLE folders(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  owner_id uuid REFERENCES users(id) NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT timeofday()::TIMESTAMP,
  name VARCHAR(1024) NOT NULL CHECK (name != '')
);
CREATE INDEX idx_folders_owner_id ON folders(owner_id);
CREATE INDEX idx_folders_id_owner_id ON folders(id, owner_id);
CREATE UNIQUE INDEX idx_folders_owner_id_name ON folders(owner_id, lower(name));


CREATE TABLE torrents(
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  folder VARCHAR(1024) REFERENCES folders(name) ON DELETE CASCADE NOT NULL,
  owner_id uuid REFERENCES users(id) NOT NULL,
  date_created TIMESTAMP NOT NULL DEFAULT timeofday()::TIMESTAMP,
  info_hash TEXT NOT NULL CHECK (info_hash ~ '^[A-F0-9]{40}$'),
  source_url TEXT,
  data BYTEA
);
CREATE INDEX idx_torrents_folder ON torrents(folder);
CREATE INDEX idx_torrents_id_owner_id ON torrents(id, owner_id);
CREATE UNIQUE INDEX idx_torrents_id_info_hash ON torrents(id, info_hash);
