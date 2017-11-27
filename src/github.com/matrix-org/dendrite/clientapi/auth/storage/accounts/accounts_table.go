// Copyright 2017 Vector Creations Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matrix-org/dendrite/clientapi/auth/authtypes"
	"github.com/matrix-org/gomatrixserverlib"
)

const insertAccountSQL = "" +
	"INSERT INTO account_accounts(localpart, created_ts, password_hash) VALUES ($1, $2, $3)"

const selectAccountByLocalpartSQL = "" +
	"SELECT localpart FROM account_accounts WHERE localpart = $1"

const selectPasswordHashSQL = "" +
	"SELECT password_hash FROM account_accounts WHERE localpart = $1"

// TODO: Update password

type accountsStatements struct {
	insertAccountStmt            *sql.Stmt
	selectAccountByLocalpartStmt *sql.Stmt
	selectPasswordHashStmt       *sql.Stmt
	serverName                   gomatrixserverlib.ServerName
}

func (s *accountsStatements) prepare(db *sql.DB, server gomatrixserverlib.ServerName) (err error) {
	if s.insertAccountStmt, err = db.Prepare(insertAccountSQL); err != nil {
		return
	}
	if s.selectAccountByLocalpartStmt, err = db.Prepare(selectAccountByLocalpartSQL); err != nil {
		return
	}
	if s.selectPasswordHashStmt, err = db.Prepare(selectPasswordHashSQL); err != nil {
		return
	}
	s.serverName = server
	return
}

// insertAccount creates a new account. 'hash' should be the password hash for this account. If it is missing,
// this account will be passwordless. Returns an error if this account already exists. Returns the account
// on success.
func (s *accountsStatements) insertAccount(
	ctx context.Context, localpart, hash string,
) (*authtypes.Account, error) {
	createdTimeMS := time.Now().UnixNano() / 1000000
	stmt := s.insertAccountStmt
	if _, err := stmt.ExecContext(ctx, localpart, createdTimeMS, hash); err != nil {
		return nil, err
	}
	return &authtypes.Account{
		Localpart:  localpart,
		UserID:     makeUserID(localpart, s.serverName),
		ServerName: s.serverName,
	}, nil
}

func (s *accountsStatements) selectPasswordHash(
	ctx context.Context, localpart string,
) (hash string, err error) {
	err = s.selectPasswordHashStmt.QueryRowContext(ctx, localpart).Scan(&hash)
	return
}

func (s *accountsStatements) selectAccountByLocalpart(
	ctx context.Context, localpart string,
) (*authtypes.Account, error) {
	var acc authtypes.Account
	stmt := s.selectAccountByLocalpartStmt
	err := stmt.QueryRowContext(ctx, localpart).Scan(&acc.Localpart)
	if err != nil {
		acc.UserID = makeUserID(localpart, s.serverName)
		acc.ServerName = s.serverName
	}
	return &acc, err
}

func makeUserID(localpart string, server gomatrixserverlib.ServerName) string {
	return fmt.Sprintf("@%s:%s", localpart, string(server))
}
