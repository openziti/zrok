package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type AccountMfa struct {
	Id         int
	AccountId  int
	TotpSecret string
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MfaRecoveryCode struct {
	Id           int
	AccountMfaId int
	CodeHash     string
	Used         bool
	UsedAt       *time.Time
	CreatedAt    time.Time
}

type MfaPendingAuth struct {
	Id           int
	AccountId    int
	PendingToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type MfaChallengeToken struct {
	Id             int
	AccountId      int
	ChallengeToken string
	ExpiresAt      time.Time
	CreatedAt      time.Time
}

// AccountMfa CRUD

func (str *Store) CreateAccountMfa(mfa *AccountMfa, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into account_mfa (account_id, totp_secret, enabled) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing account_mfa insert statement")
	}
	var id int
	if err := stmt.QueryRow(mfa.AccountId, mfa.TotpSecret, mfa.Enabled).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing account_mfa insert statement")
	}
	return id, nil
}

func (str *Store) GetAccountMfa(id int, trx *sqlx.Tx) (*AccountMfa, error) {
	mfa := &AccountMfa{}
	if err := trx.QueryRowx("select * from account_mfa where id = $1", id).StructScan(mfa); err != nil {
		return nil, errors.Wrap(err, "error selecting account_mfa by id")
	}
	return mfa, nil
}

func (str *Store) FindAccountMfaByAccountId(accountId int, trx *sqlx.Tx) (*AccountMfa, error) {
	mfa := &AccountMfa{}
	if err := trx.QueryRowx("select * from account_mfa where account_id = $1", accountId).StructScan(mfa); err != nil {
		return nil, errors.Wrap(err, "error selecting account_mfa by account_id")
	}
	return mfa, nil
}

func (str *Store) UpdateAccountMfa(mfa *AccountMfa, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update account_mfa set totp_secret = $1, enabled = $2, updated_at = current_timestamp where id = $3")
	if err != nil {
		return errors.Wrap(err, "error preparing account_mfa update statement")
	}
	if _, err := stmt.Exec(mfa.TotpSecret, mfa.Enabled, mfa.Id); err != nil {
		return errors.Wrap(err, "error executing account_mfa update statement")
	}
	return nil
}

func (str *Store) DeleteAccountMfa(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from account_mfa where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing account_mfa delete statement")
	}
	if _, err := stmt.Exec(id); err != nil {
		return errors.Wrap(err, "error executing account_mfa delete statement")
	}
	return nil
}

// MfaRecoveryCode CRUD

func (str *Store) CreateMfaRecoveryCodes(accountMfaId int, codeHashes []string, trx *sqlx.Tx) error {
	for _, hash := range codeHashes {
		stmt, err := trx.Prepare("insert into mfa_recovery_codes (account_mfa_id, code_hash) values ($1, $2)")
		if err != nil {
			return errors.Wrap(err, "error preparing mfa_recovery_codes insert statement")
		}
		if _, err := stmt.Exec(accountMfaId, hash); err != nil {
			return errors.Wrap(err, "error executing mfa_recovery_codes insert statement")
		}
	}
	return nil
}

func (str *Store) FindUnusedMfaRecoveryCodes(accountMfaId int, trx *sqlx.Tx) ([]*MfaRecoveryCode, error) {
	rows, err := trx.Queryx("select * from mfa_recovery_codes where account_mfa_id = $1 and not used", accountMfaId)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting unused mfa_recovery_codes")
	}
	defer rows.Close()

	var codes []*MfaRecoveryCode
	for rows.Next() {
		code := &MfaRecoveryCode{}
		if err := rows.StructScan(code); err != nil {
			return nil, errors.Wrap(err, "error scanning mfa_recovery_code")
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (str *Store) FindMfaRecoveryCodeByHash(accountMfaId int, codeHash string, trx *sqlx.Tx) (*MfaRecoveryCode, error) {
	code := &MfaRecoveryCode{}
	if err := trx.QueryRowx("select * from mfa_recovery_codes where account_mfa_id = $1 and code_hash = $2 and not used", accountMfaId, codeHash).StructScan(code); err != nil {
		return nil, errors.Wrap(err, "error selecting mfa_recovery_code by hash")
	}
	return code, nil
}

func (str *Store) MarkMfaRecoveryCodeUsed(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("update mfa_recovery_codes set used = true, used_at = current_timestamp where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing mfa_recovery_codes update statement")
	}
	if _, err := stmt.Exec(id); err != nil {
		return errors.Wrap(err, "error executing mfa_recovery_codes update statement")
	}
	return nil
}

func (str *Store) DeleteMfaRecoveryCodesByAccountMfaId(accountMfaId int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from mfa_recovery_codes where account_mfa_id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing mfa_recovery_codes delete statement")
	}
	if _, err := stmt.Exec(accountMfaId); err != nil {
		return errors.Wrap(err, "error executing mfa_recovery_codes delete statement")
	}
	return nil
}

func (str *Store) CountUnusedMfaRecoveryCodes(accountMfaId int, trx *sqlx.Tx) (int, error) {
	var count int
	if err := trx.QueryRow("select count(*) from mfa_recovery_codes where account_mfa_id = $1 and not used", accountMfaId).Scan(&count); err != nil {
		return 0, errors.Wrap(err, "error counting unused mfa_recovery_codes")
	}
	return count, nil
}

// MfaPendingAuth CRUD

func (str *Store) CreateMfaPendingAuth(pa *MfaPendingAuth, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into mfa_pending_auth (account_id, pending_token, expires_at) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing mfa_pending_auth insert statement")
	}
	var id int
	if err := stmt.QueryRow(pa.AccountId, pa.PendingToken, pa.ExpiresAt).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing mfa_pending_auth insert statement")
	}
	return id, nil
}

func (str *Store) FindMfaPendingAuthByToken(token string, trx *sqlx.Tx) (*MfaPendingAuth, error) {
	pa := &MfaPendingAuth{}
	if err := trx.QueryRowx("select * from mfa_pending_auth where pending_token = $1", token).StructScan(pa); err != nil {
		return nil, errors.Wrap(err, "error selecting mfa_pending_auth by token")
	}
	return pa, nil
}

func (str *Store) DeleteMfaPendingAuth(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from mfa_pending_auth where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing mfa_pending_auth delete statement")
	}
	if _, err := stmt.Exec(id); err != nil {
		return errors.Wrap(err, "error executing mfa_pending_auth delete statement")
	}
	return nil
}

func (str *Store) FindExpiredMfaPendingAuths(before time.Time, limit int, trx *sqlx.Tx) ([]*MfaPendingAuth, error) {
	var sql string
	switch str.cfg.Type {
	case "postgres":
		sql = "select * from mfa_pending_auth where expires_at < $1 limit %d for update"
	case "sqlite3":
		sql = "select * from mfa_pending_auth where expires_at < $1 limit %d"
	default:
		return nil, errors.Errorf("unknown database type '%v'", str.cfg.Type)
	}

	rows, err := trx.Queryx(fmt.Sprintf(sql, limit), before)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting expired mfa_pending_auths")
	}
	defer rows.Close()

	var pas []*MfaPendingAuth
	for rows.Next() {
		pa := &MfaPendingAuth{}
		if err := rows.StructScan(pa); err != nil {
			return nil, errors.Wrap(err, "error scanning mfa_pending_auth")
		}
		pas = append(pas, pa)
	}
	return pas, nil
}

func (str *Store) DeleteMultipleMfaPendingAuths(ids []int, trx *sqlx.Tx) error {
	if len(ids) == 0 {
		return nil
	}

	anyIds := make([]any, len(ids))
	indexes := make([]string, len(ids))

	for i, id := range ids {
		anyIds[i] = id
		indexes[i] = fmt.Sprintf("$%d", i+1)
	}

	stmt, err := trx.Prepare(fmt.Sprintf("delete from mfa_pending_auth where id in (%s)", strings.Join(indexes, ",")))
	if err != nil {
		return errors.Wrap(err, "error preparing mfa_pending_auth delete multiple statement")
	}
	if _, err := stmt.Exec(anyIds...); err != nil {
		return errors.Wrap(err, "error executing mfa_pending_auth delete multiple statement")
	}
	return nil
}

// MfaChallengeToken CRUD

func (str *Store) CreateMfaChallengeToken(ct *MfaChallengeToken, trx *sqlx.Tx) (int, error) {
	stmt, err := trx.Prepare("insert into mfa_challenge_tokens (account_id, challenge_token, expires_at) values ($1, $2, $3) returning id")
	if err != nil {
		return 0, errors.Wrap(err, "error preparing mfa_challenge_tokens insert statement")
	}
	var id int
	if err := stmt.QueryRow(ct.AccountId, ct.ChallengeToken, ct.ExpiresAt).Scan(&id); err != nil {
		return 0, errors.Wrap(err, "error executing mfa_challenge_tokens insert statement")
	}
	return id, nil
}

func (str *Store) FindMfaChallengeTokenByToken(token string, trx *sqlx.Tx) (*MfaChallengeToken, error) {
	ct := &MfaChallengeToken{}
	if err := trx.QueryRowx("select * from mfa_challenge_tokens where challenge_token = $1", token).StructScan(ct); err != nil {
		return nil, errors.Wrap(err, "error selecting mfa_challenge_token by token")
	}
	return ct, nil
}

func (str *Store) DeleteMfaChallengeToken(id int, trx *sqlx.Tx) error {
	stmt, err := trx.Prepare("delete from mfa_challenge_tokens where id = $1")
	if err != nil {
		return errors.Wrap(err, "error preparing mfa_challenge_tokens delete statement")
	}
	if _, err := stmt.Exec(id); err != nil {
		return errors.Wrap(err, "error executing mfa_challenge_tokens delete statement")
	}
	return nil
}

func (str *Store) FindExpiredMfaChallengeTokens(before time.Time, limit int, trx *sqlx.Tx) ([]*MfaChallengeToken, error) {
	var sql string
	switch str.cfg.Type {
	case "postgres":
		sql = "select * from mfa_challenge_tokens where expires_at < $1 limit %d for update"
	case "sqlite3":
		sql = "select * from mfa_challenge_tokens where expires_at < $1 limit %d"
	default:
		return nil, errors.Errorf("unknown database type '%v'", str.cfg.Type)
	}

	rows, err := trx.Queryx(fmt.Sprintf(sql, limit), before)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting expired mfa_challenge_tokens")
	}
	defer rows.Close()

	var cts []*MfaChallengeToken
	for rows.Next() {
		ct := &MfaChallengeToken{}
		if err := rows.StructScan(ct); err != nil {
			return nil, errors.Wrap(err, "error scanning mfa_challenge_token")
		}
		cts = append(cts, ct)
	}
	return cts, nil
}

func (str *Store) DeleteMultipleMfaChallengeTokens(ids []int, trx *sqlx.Tx) error {
	if len(ids) == 0 {
		return nil
	}

	anyIds := make([]any, len(ids))
	indexes := make([]string, len(ids))

	for i, id := range ids {
		anyIds[i] = id
		indexes[i] = fmt.Sprintf("$%d", i+1)
	}

	stmt, err := trx.Prepare(fmt.Sprintf("delete from mfa_challenge_tokens where id in (%s)", strings.Join(indexes, ",")))
	if err != nil {
		return errors.Wrap(err, "error preparing mfa_challenge_tokens delete multiple statement")
	}
	if _, err := stmt.Exec(anyIds...); err != nil {
		return errors.Wrap(err, "error executing mfa_challenge_tokens delete multiple statement")
	}
	return nil
}
