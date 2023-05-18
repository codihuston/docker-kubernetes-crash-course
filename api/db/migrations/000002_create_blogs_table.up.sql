CREATE TABLE IF NOT EXISTS blogs(
   id serial PRIMARY KEY,
   title VARCHAR (50) UNIQUE NOT NULL,
   body TEXT NOT NULL,
   created_at TIMESTAMPTZ,
   updated_at TIMESTAMPTZ,
   deleted_at TIMESTAMPTZ
);
