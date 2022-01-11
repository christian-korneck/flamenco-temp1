CREATE TABLE jobs (
  uuid CHAR(36) PRIMARY KEY NOT NULL,
  name VARCHAR(255) NOT NULL,
  jobType VARCHAR(255) NOT NULL,
  priority INT NOT NULL,
  created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE job_settings (
  job_id CHAR(36),
  key VARCHAR(255),
  value TEXT,
  FOREIGN KEY(job_id) REFERENCES jobs(uuid)
);
CREATE UNIQUE INDEX job_settings_index ON job_settings(job_id, key);

CREATE TABLE job_metadata (
  job_id CHAR(36),
  key VARCHAR(255),
  value TEXT,
  FOREIGN KEY(job_id) REFERENCES jobs(uuid)
);
CREATE UNIQUE INDEX job_metadata_index ON job_metadata(job_id, key);
