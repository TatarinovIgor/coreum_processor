package user

import (
	"coreum_processor/modules/service"
	"coreum_processor/modules/storage"
	"time"
)

type Service struct {
	userStorage *storage.UserPSQL
	merchants   *service.Merchants
}

func (s *Service) AddUser(identity, firstName, lastName string) error {
	return s.userStorage.CreateUser(identity, firstName, lastName)
}

func (s *Service) GetUser(identity string) (*storage.UserStore, error) {
	return s.userStorage.GetUserByIdentity(identity)
}

func (s *Service) GetUserList(merchID string, accessFilter []int, from, to time.Time) ([]storage.UserStore, error) {
	return s.userStorage.GetUserList(merchID, accessFilter, from, to)
}

func (s *Service) GetUserMerchants(identity string) ([]storage.UserMerchant, error) {
	return s.userStorage.GetUserMerchants(identity)
}

func (s *Service) SetUserAccess(identity string, access storage.UserAccess) (*storage.UserStore, error) {
	return s.userStorage.SetUserAccess(identity, access)
}
func (s *Service) UpdateUser(userStore storage.UserStore) error {
	return s.userStorage.UpdateUser(userStore)
}

func (s *Service) LinkUserToMerchant(identity, merchantID string) error {
	return s.userStorage.LinkUserToMerchant(identity, merchantID, storage.MerchantAccess{})
}

func (s *Service) ApproveUserMerchant(identity, merchantID string) error {
	return s.userStorage.ApproveUserMerchant(identity, merchantID)
}

func (s *Service) RequestMerchantForUser(identity, merchantName, merchantEmail string) error {
	return s.userStorage.RequestMerchantForUser(identity, merchantName, merchantEmail)
}

// NewService create a service to process operation with users and merchant settings
func NewService(userStorage *storage.UserPSQL, merchants *service.Merchants) *Service {
	return &Service{
		userStorage: userStorage,
		merchants:   merchants,
	}
}
