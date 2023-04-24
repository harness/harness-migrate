// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"           // required for postgres
	_ "github.com/mattn/go-sqlite3" // required for sqlite3
)

var statementBuilder squirrel.StatementBuilderType

// repository is an implementation of the Repository interface that
// provides access to the Drone database using SQLite.
type repository struct {
	db     *sqlx.DB
	dbType string
}

func initStatementBuilder(driver string) {
	switch driver {
	case "postgres":
		statementBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	default:
		statementBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)
	}
}

// NewRepository returns a new Repository that provides access to the Drone
// database using the specified connection string.
func NewRepository(driver, datasource string, db *sqlx.DB) (Repository, error) {
	var err error

	initStatementBuilder(driver)

	// If a DB connection is not provided, create one
	if db == nil {
		db, err = createDBConnection(driver, datasource)
		if err != nil {
			return nil, fmt.Errorf("failed to create DB connection: %w", err)
		}
	}

	// Check the connection using the Ping method
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database %s: %w", datasource, err)
	}

	return &repository{
		db:     db,
		dbType: driver,
	}, nil
}

// createDBConnection creates a new database connection based on the specified driver and datasource.
func createDBConnection(driver, datasource string) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	switch driver {
	case "postgres":
		db, err = sqlx.Open("postgres", datasource)
	case "mysql":
		db, err = sqlx.Open("mysql", datasource)
	case "sqlite3":
		db, err = sqlx.Open("sqlite3", datasource)
	default:
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to datasource %s: %w", datasource, err)
	}

	return db, nil
}

func (r *repository) GetRepos(ctx context.Context, namespace string) ([]*Repo, error) {
	var repos []*Repo
	query, args, err := statementBuilder.
		Select("*").
		From("repos").
		Where(squirrel.Eq{"repo_namespace": namespace}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query: %w", err)
	}
	err = r.db.SelectContext(ctx, &repos, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get repos: %w", err)
	}
	return repos, nil
}

func (r *repository) LatestBuild(ctx context.Context, repoId int64) (*Build, error) {
	var builds Build
	query, args, err := statementBuilder.
		Select("build_id", "build_repo_id", "build_trigger", "build_number", "build_parent", "build_status", "build_error",
			"build_event", "build_action", "build_link", "build_timestamp", "build_title", "build_message", "build_before",
			"build_after", "build_ref", "build_source_repo", "build_source", "build_target", "build_author", "build_author_name",
			"build_author_email", "build_author_avatar", "build_sender", "build_params", "build_cron", "build_deploy", "build_deploy_id",
			"build_debug", "build_started", "build_finished", "build_created", "build_updated").
		From("builds").
		Where("build_repo_id = ?", repoId).
		OrderBy("build_id DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}
	err = r.db.GetContext(ctx, &builds, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &builds, nil
}

func (r *repository) GetSecrets(ctx context.Context, repoID int64) ([]*Secret, error) {
	var secrets []*Secret
	query, args, err := statementBuilder.
		Select("secret_id", "secret_repo_id", "secret_name", "secret_data", "secret_pull_request", "secret_pull_request_push").
		From("secrets").
		Where(squirrel.Eq{"secret_repo_id": repoID}).
		OrderBy("secret_name").ToSql()
	if err != nil {
		return nil, err
	}
	err = r.db.SelectContext(ctx, &secrets, query, args...)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (r *repository) GetOrgSecrets(ctx context.Context, namespace string) ([]*OrgSecret, error) {
	var secrets []*OrgSecret
	query, args, err := statementBuilder.
		Select("orgsecrets.*").
		From("orgsecrets").
		Where(squirrel.Eq{"orgsecrets.secret_namespace": namespace}).
		OrderBy("orgsecrets.secret_name").ToSql()
	if err != nil {
		return nil, err
	}
	err = r.db.SelectContext(ctx, &secrets, query, args...)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}
