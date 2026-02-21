package users

import (
	"context"
	"lazeez-core/config"
	"lazeez-core/internal/branch"
	"lazeez-core/internal/common"
)

type UserService interface {
	CreateUser(ctx context.Context, req UserRequest) error
	CreateSuperAdminUser(ctx context.Context, req SuperAdminUserRequest) error
	GetUserByID(ctx context.Context, id string) (UserDTO, error)
	UpdateUser(ctx context.Context, id string, req UserRequest) error
	DeleteUser(ctx context.Context, id string) error
	UserLookUp(ctx context.Context, phoneNumber string) (UserDTO, error)
	GetAllUsers(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*UserDTO], error)
	GetUserByBranchID(ctx context.Context, branchID string) (UserDTO, error)
	UnDeleteUser(ctx context.Context, id string) error
	// Internal methods for auth module
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*User, error)
	GetUserDTOWithMerchant(ctx context.Context, user *User) (UserDTO, error)
	GetUserByIDForLogin(ctx context.Context, userID string) (UserDTO, error)
	UpdateLoggingAttempts(ctx context.Context, id string, attempts int) error
	ResetLoggingAttempts(ctx context.Context, id string) error
	LockUser(ctx context.Context, id string) error
	SetPassword(ctx context.Context, phoneNumber, id, password string, isFirstLogin bool) error
}

type userService struct {
	userRepository UserRepository
	branchService  branch.BranchService
	logger         config.Logger
}

func NewUserService(userRepository UserRepository, branchService branch.BranchService, logger config.Logger) UserService {
	return &userService{
		userRepository: userRepository,
		branchService:  branchService,
		logger:         logger,
	}
}

func (s *userService) CreateUser(ctx context.Context, req UserRequest) error {
	// req.PhoneNumber is already validated and normalized by the handler
	if err := s.userRepository.CheckUserExistsByPhoneNumber(ctx, req.PhoneNumber); err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return err
	}
	if err := s.validateBranch(ctx, req.BranchID, req.MerchantID); err != nil {
		s.logger.Error("Failed to validate branch", "error", err)
		return err
	}
	newUser := s.createUserDefaultData(&req)
	err := s.userRepository.CreateUser(ctx, *newUser)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return err
	}
	return nil
}

func (s *userService) CreateSuperAdminUser(ctx context.Context, req SuperAdminUserRequest) error {
	// req.PhoneNumber is already validated and normalized by the handler
	if err := s.userRepository.CheckUserExistsByPhoneNumber(ctx, req.PhoneNumber); err != nil {
		s.logger.Error("Failed to check user exists by phone number", "error", err)
		return err
	}
	newUser := s.createSuperAdminUserDefaultData(&req)
	err := s.userRepository.CreateUser(ctx, *newUser)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return err
	}
	return nil
}


func (s *userService) GetUserByID(ctx context.Context, id string) (UserDTO, error) {
	s.logger.Info("Getting user by ID", "id", id)
	user, err := s.userRepository.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", "error", err)
		return UserDTO{}, err
	}
	result := user.ToDTO()
	return result, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req UserRequest) error {
	existingUser, err := s.userRepository.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", "error", err)
		return err
	}

	// req.PhoneNumber, when set, is already validated and normalized by the handler
	user := s.updateUserDefaultData(&req, &existingUser)
	return s.userRepository.UpdateUser(ctx, *user)
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Deleting user", "id", id)
	return s.userRepository.DeleteUser(ctx, id)
}

func (s *userService) UserLookUp(ctx context.Context, phoneNumber string) (UserDTO, error) {
	// phoneNumber is already validated and normalized by the handler
	s.logger.Info("Looking up user by phone number", "phone number", phoneNumber)
	user, err := s.userRepository.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return UserDTO{}, err
	}
	return user.ToDTO(), nil
}

func (s *userService) GetAllUsers(ctx context.Context, filter common.Filter) (*common.PaginatedResponse[[]*UserDTO], error) {
	s.logger.Info("Getting all users")
	result, err := s.userRepository.GetAllUsers(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get all users", "error", err)
		return nil, err
	}
	userDTOs := make([]*UserDTO, len(result.Data))
	for i, user := range result.Data {
		dto := user.ToDTO()
		userDTOs[i] = &dto
	}
	return &common.PaginatedResponse[[]*UserDTO]{
		Data: userDTOs,
		Meta: result.Meta,
	}, nil
}

func (s *userService) GetUserByBranchID(ctx context.Context, branchID string) (UserDTO, error) {
	s.logger.Info("Getting user by branch ID", "branchID", branchID)
	user, err := s.userRepository.GetUserByBranchID(ctx, branchID)
	if err != nil {
		s.logger.Error("Failed to get user by branch ID", "error", err)
		return UserDTO{}, err
	}
	result := user.ToDTO()
	return result, nil
}

func (s *userService) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*User, error) {
	user, err := s.userRepository.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		s.logger.Error("Failed to get user by phone number", "error", err)
		return nil, err
	}
	return &user, nil
}

// GetUserDTOWithMerchant returns UserDTO with MerchantID populated from branch (for branch_manager and super_branch_admin).
func (s *userService) GetUserDTOWithMerchant(ctx context.Context, user *User) (UserDTO, error) {
	dto := user.ToDTO()
	if user.BranchID.Valid && user.BranchID.String != "" {
		branch, err := s.branchService.Get(ctx, user.BranchID.String)
		if err == nil {
			dto.MerchantID = branch.MerchantID
		}
	}
	return dto, nil
}

// GetUserByIDForLogin returns UserDTO with MerchantID for login/refresh response.
func (s *userService) GetUserByIDForLogin(ctx context.Context, userID string) (UserDTO, error) {
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user by ID for login", "error", err)
		return UserDTO{}, err
	}
	return s.GetUserDTOWithMerchant(ctx, &user)
}

func (s *userService) UpdateLoggingAttempts(ctx context.Context, id string, attempts int) error {
	return s.userRepository.UpdateLoggingAttempts(ctx, id, attempts)
}

func (s *userService) ResetLoggingAttempts(ctx context.Context, id string) error {
	return s.userRepository.ResetLoggingAttempts(ctx, id)
}

func (s *userService) LockUser(ctx context.Context, id string) error {
	return s.userRepository.LockUser(ctx, id)
}

func (s *userService) SetPassword(ctx context.Context, phoneNumber, id, password string, isFirstLogin bool) error {
	return s.userRepository.SetPassword(ctx, phoneNumber, id, password, isFirstLogin)
}

func (s *userService) UnDeleteUser(ctx context.Context, id string) error {
	s.logger.Info("Undeleting user", "id", id)
	err := s.userRepository.UnDeleteUser(ctx, id)
	if err != nil {
		s.logger.Error("Failed to undelete user", "error", err)
		return err
	}
	return nil
}