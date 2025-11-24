CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS teams (
  team_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  parent_id TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY, 
  username TEXT,
  name TEXT,
  password TEXT, 
  email TEXT,
  role TEXT,
  color TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS members (
  id BIGSERIAL PRIMARY KEY,
  clickup_id TEXT UNIQUE,
  username TEXT,
  name TEXT,
  email TEXT,
  color TEXT,
  team_id TEXT REFERENCES teams(team_id) ON DELETE SET NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  name TEXT,
  text_content TEXT,
  description TEXT,
  status_id TEXT,
  status_name TEXT,
  status_type TEXT,
  status_color TEXT,
  date_done BIGINT,    
  date_closed BIGINT,  
  assignee_clickup_id TEXT, 
  assignee_user_id BIGINT,  
  assignee_username TEXT,
  assignee_email TEXT,
  assignee_color TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
