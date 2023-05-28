package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type UserAccess uint
type MerchantAccess json.RawMessage

const (
	UserBlocked    UserAccess = 0x0000
	UserRegistered UserAccess = 0x0001
	UserOnboarded  UserAccess = 0x0002
	UserSysAdmin   UserAccess = 0x1000
)

type UserStore struct {
	Id                 int             `json:"-"`
	Identity           string          `json:"identity"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          *time.Time      `json:"-"`
	FirstName          string          `json:"first_name"`
	LastName           string          `json:"last_name"`
	TermsAndConditions *bool           `json:"terms_and_conditions"`
	Access             UserAccess      `json:"access"`
	MetaData           json.RawMessage `json:"meta_data"`
}

type UserMerchant struct {
	MerchantID       string          `json:"merchant_id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"-"`
	MerchantName     string          `json:"company_name"`
	Email            string          `json:"email"`
	MerchantAccess   MerchantAccess  `json:"merchant_access"`
	IsBlocked        bool            `json:"is_blocked"`
	UserMetaData     json.RawMessage `json:"user_meta_data"`
	MerchantMetaData json.RawMessage `json:"merchant_meta_data"`
}

type UserPSQL struct {
	db                     *sql.DB
	userNamespace          string
	merchantUsersNamespace string
	merchantListNamespace  string
	defaultAccess          UserAccess
}

// GetUserByIdentity find a user in the store by unique user identity
func (s *UserPSQL) GetUserByIdentity(identity string) (*UserStore, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE identity = '%s'", s.userNamespace, identity)
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	userStore, err := rowsToUser(rows)
	if err != nil {
		return nil, err
	}
	if len(userStore) == 1 {
		return &userStore[0], nil
	}
	return nil, fmt.Errorf("unexpected number of users: %v, for identity: %v", len(userStore), identity)
}

// GetUserList get a filtered list of users
func (s *UserPSQL) GetUserList(merchID string, accessFilter []int, from, to time.Time) ([]UserStore, error) {
	un := s.userNamespace
	query := fmt.Sprintf("SELECT %s.id, %s.created_at, %s.updated_at, %s.deleted_at, "+
		" %s.identity, %s.first_name, %s.last_name, %s.terms_and_conditions, %s.access, %s.meta_data FROM %s ",
		un, un, un, un, un, un, un, un, un, un, un)

	query += fmt.Sprintf(" join %s mu on %s.id = mu.user_id ",
		s.merchantUsersNamespace, un)
	query += fmt.Sprintf("join %s ml on mu.merchant_list_id = ml.id ", s.merchantListNamespace)
	if merchID != "" {
		query += fmt.Sprintf(" where merchant_id = '%s' and ", merchID)
	} else {
		query += fmt.Sprintf(" where merchant_id IS NULL AND ")
	}

	query += fmt.Sprintf(" %s.deleted_at IS NULL and %s.created_at > '%v' and %s.created_at < '%v' ",
		un, un, from.Format(time.RFC3339), un, to.Format(time.RFC3339))
	if accessFilter != nil && len(accessFilter) > 0 {
		query += fmt.Sprintf(" and %s.access in (", un)
		for i, val := range accessFilter {
			if i > 0 {
				query += ", "
			}
			query += fmt.Sprintf("%v", val)
		}
		query += ") "
	}
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	userStore, err := rowsToUser(rows)
	if err != nil {
		return nil, fmt.Errorf("could not execute queryL: %w", err)
	}
	return userStore, nil
}

// GetUserMerchants find all merchant linked with the user in the store by unique user identity
func (s *UserPSQL) GetUserMerchants(identity string) ([]UserMerchant, error) {
	query := fmt.Sprintf("select ml.merchant_id, ml.created_at, ml.updated_at, ml.deleted_at, "+
		"ml.company_name, ml.email, ml.is_blocked, mu.access, mu.meta_data, ml.meta_data "+
		"from %s join %s mu on %s.id = mu.user_id join %s ml on ml.id = mu.merchant_list_id where identity = '%s'",
		s.userNamespace, s.merchantUsersNamespace, s.userNamespace, s.merchantListNamespace, identity)
	rows, err := s.db.Query(query)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}

	return rowsToMerchants(rows)
}

// CreateUser makes a new record in the user store by unique user identity
func (s *UserPSQL) CreateUser(identity, firstName, lastName string) error {
	query := fmt.Sprintf("INSERT INTO %s (identity, created_at, updated_at, first_name, last_name, terms_and_conditions, access)",
		s.userNamespace)
	query += "VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	_, err := s.db.Query(query,
		identity, time.Now().UTC(), time.Now().UTC(), firstName, lastName, false, s.defaultAccess)
	if err != nil {
		return err
	}
	return nil
}

// SetUserAccess sets the user  access in store by unique user identity and access
func (s *UserPSQL) SetUserAccess(identity string, access UserAccess) (*UserStore, error) {
	query := fmt.Sprintf(
		"UPDATE %s SET updated_at = $2, terms_and_conditions = $3, access = $4 WHERE identity = $1 RETURNING *",
		s.userNamespace)
	rows, err := s.db.Query(query,
		identity, time.Now().UTC(), true, access)
	if err != nil {
		return nil, err
	}
	userStore, err := rowsToUser(rows)
	if err != nil {
		return nil, err
	}
	return &userStore[0], nil
}

// UpdateUser updates the user in store by unique user identity
func (s *UserPSQL) UpdateUser(user UserStore) error {
	query := fmt.Sprintf(
		"UPDATE %s SET updated_at = $2, first_name = $3, last_name = $4, terms_and_conditions = $5, access = $6, meta_data = $7 WHERE identity = $1",
		s.userNamespace)
	_, err := s.db.Query(query,
		user.Identity, time.Now().UTC(), user.FirstName, user.LastName, user.TermsAndConditions, user.Access, user.MetaData)
	if err != nil {
		return err
	}
	return nil
}

// LinkUserToMerchant link the user to a merchant store by unique user identity and merchant id
func (s *UserPSQL) LinkUserToMerchant(identity, merchantID string, merchantAccess MerchantAccess) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (created_at, updated_at, deleted_at, user_id, merchant_list_id)"+
			"values (now(), now(), null, (select id from %s where identity = '%s'), "+
			"(select id from %s where merchant_id = '%s'))",
		s.merchantUsersNamespace, s.userNamespace, identity, s.merchantListNamespace, merchantID)
	_, err := s.db.Query(query)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserPSQL) ApproveUserMerchant(identity, merchantID string) error {
	query := fmt.Sprintf(
		"WITH user_id_var AS (SELECT id FROM %s WHERE identity = $1), "+
			"merchant_id_var AS (SELECT merchant_list_id FROM %s JOIN %s ml ON merchant_list_id = ml.id "+
			"WHERE user_id IN (SELECT id FROM user_id_var) AND ml.merchant_id IS NULL) "+
			"UPDATE %s SET updated_at = $2, merchant_id = $3 WHERE id IN (SELECT merchant_list_id FROM merchant_id_var)",
		s.userNamespace, s.merchantUsersNamespace, s.merchantListNamespace, s.merchantListNamespace)

	_, err := s.db.Query(query,
		identity, time.Now().UTC(), merchantID)
	if err != nil {
		return err
	}
	return nil
}

// RequestMerchantForUser link the user to a merchant store by unique user identity and merchant id
func (s *UserPSQL) RequestMerchantForUser(identity, merchantName, merchantEmail string) error {
	query := fmt.Sprintf(
		"WITH merchantID AS (INSERT INTO %s (created_at, updated_at, deleted_at, email, company_name) "+
			"values (now(), now(), null, '%s', '%s') returning id) "+
			"INSERT INTO %s (created_at, updated_at, deleted_at, user_id, merchant_list_id) "+
			"values (now(), now(), null, (select id from %s where identity = '%s'), "+
			"(select id from merchantID))",
		s.merchantListNamespace, merchantEmail, merchantName, s.merchantUsersNamespace, s.userNamespace, identity)
	_, err := s.db.Query(query)
	if err != nil {
		return err
	}
	return nil
}
func NewUserStorage(defaultAccess UserAccess, userNamespace, merchantUsersNamespace, merchantListNamespace string,
	db *sql.DB) (*UserPSQL, error) {
	s := UserPSQL{
		db:                     db,
		userNamespace:          userNamespace,
		merchantUsersNamespace: merchantUsersNamespace,
		merchantListNamespace:  merchantListNamespace,
		defaultAccess:          defaultAccess,
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to DB: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", userNamespace)); err != nil {
		return nil, fmt.Errorf("could not connect to user storage: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", merchantUsersNamespace)); err != nil {
		return nil, fmt.Errorf("could not connect to merchant's users storage: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %q LIMIT 1", merchantListNamespace)); err != nil {
		return nil, fmt.Errorf("could not connect to merchant list: %v", err)
	}
	return &s, nil
}

func rowsToUser(rows *sql.Rows) ([]UserStore, error) {
	var users []UserStore
	if rows == nil {
		return users, nil
	}
	for rows.Next() {
		user := UserStore{}
		if err := rows.Scan(
			&user.Id,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
			&user.Identity, &user.FirstName, &user.LastName,
			&user.TermsAndConditions,
			&user.Access,
			&user.MetaData,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
func rowsToMerchants(rows *sql.Rows) ([]UserMerchant, error) {
	var merchants []UserMerchant
	if rows == nil {
		return merchants, nil
	}
	for rows.Next() {
		merchant := UserMerchant{}
		if err := rows.Scan(
			&merchant.MerchantID,
			&merchant.CreatedAt, &merchant.UpdatedAt, &merchant.DeletedAt,
			&merchant.MerchantName, &merchant.Email, &merchant.IsBlocked,
			&merchant.MerchantAccess,
			&merchant.UserMetaData, &merchant.MerchantMetaData,
		); err != nil {
			return nil, err
		}
		merchants = append(merchants, merchant)
	}
	return merchants, nil
}
