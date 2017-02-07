package db

import (
	"github.com/blendlabs/spiffy"
	"github.com/blendlabs/spiffy/migration"
)

// Migrate migrates the database.
func Migrate() error {
	schema := migration.New(
		"chatbus",
		migration.New(
			"users",
			migration.Step(
				migration.CreateTable,
				migration.Body(
					"CREATE TABLE users (id serial not null, uuid varchar(64) not null, display_name varchar(256));",
					"ALTER TABLE users ADD CONSTRAINT pk_users_id PRIMARY KEY (id);",
					"ALTER TABLE users ADD CONSTRAINT uk_users_uuid UNIQUE (uuid);",
				),
				"users",
			),
		),
		migration.New(
			"sessions",
			migration.Step(
				migration.CreateTable,
				migration.Body(
					"CREATE TABLE sessions (uuid varchar(64) not null, created_utc timestamp not null, last_active_utc timestamp not null, user_id int not null);",
					"ALTER TABLE sessions ADD CONSTRAINT pk_sessions_uuid PRIMARY KEY (uuid);",
					"ALTER TABLE sessions ADD CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id);",
				),
				"sessions",
			),
		),
		migration.New(
			"contacts",
			migration.Step(
				migration.CreateTable,
				migration.Body(
					"CREATE TABLE contacts (sender int not null, receiver int not null);",
					"ALTER TABLE contacts ADD CONSTRAINT pk_sessions_sender_receiver PRIMARY KEY (sender, receiver);",
					"ALTER TABLE contacts ADD CONSTRAINT fk_contacts_sender FOREIGN KEY (sender) REFERENCES users(id);",
					"ALTER TABLE contacts ADD CONSTRAINT fk_contacts_receiver FOREIGN KEY (receiver) REFERENCES users(id);",
				),
				"contacts",
			),
		),
		migration.New(
			"messages",
			migration.Step(
				migration.CreateTable,
				migration.Body(
					"CREATE TABLE messages (uuid varchar(64) not null, created_utc timestamp not null, sender int not null, receiver int not null, body varchar(1024) not null, attachments json);",
					"ALTER TABLE messages ADD CONSTRAINT pk_messages_uuid PRIMARY KEY (uuid);",
					"ALTER TABLE messages ADD CONSTRAINT fk_messages_sender FOREIGN KEY (sender) REFERENCES users(id);",
					"ALTER TABLE messages ADD CONSTRAINT fk_messages_receiver FOREIGN KEY (receiver) REFERENCES users(id);",
				),
				"messages",
			),
		),
	)
	schema.Logged(migration.NewLogger())
	return schema.Apply(spiffy.DefaultDb())
}
