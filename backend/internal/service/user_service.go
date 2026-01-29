package service

import (
	"errors"
	"net/mail"
	"strings"

	"gorm.io/gorm"
	"oj-system/internal/model"
	"oj-system/internal/repository"
	"oj-system/internal/utils"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		repo: repository.NewUserRepository(),
	}
}

// CreateUserByAdmin 管理员创建用户
func (s *UserService) CreateUserByAdmin(req *model.AdminCreateUserRequest) (*model.User, error) {
	if err := normalizeAndValidateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	if s.repo.ExistsByUsername(req.Username) {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在（仅在邮箱非空时）
	if req.Email != "" && s.repo.ExistsByEmail(req.Email) {
		return nil, errors.New("邮箱已被使用")
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	role := req.Role
	if role == "" {
		role = "user"
	}
	if role != "user" && role != "admin" {
		return nil, errors.New("无效的角色")
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		StudentID:    req.StudentID,
		Role:         role,
		Group:        req.Group,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	return user, nil
}

// CreateUsersBatch 管理员批量创建用户
func (s *UserService) CreateUsersBatch(req *model.AdminCreateUsersRequest) (int, []map[string]interface{}) {
	if req == nil || len(req.Users) == 0 {
		return 0, []map[string]interface{}{
			{"index": 0, "error": "用户列表不能为空"},
		}
	}

	created := 0
	var errorsList []map[string]interface{}

	for i := range req.Users {
		item := req.Users[i]
		if err := normalizeAndValidateCreateUserRequest(&item); err != nil {
			errorsList = append(errorsList, map[string]interface{}{
				"index":    i,
				"username": item.Username,
				"error":    err.Error(),
			})
			continue
		}

		if s.repo.ExistsByUsername(item.Username) {
			errorsList = append(errorsList, map[string]interface{}{
				"index":    i,
				"username": item.Username,
				"error":    "用户名已存在",
			})
			continue
		}

		if item.Email != "" && s.repo.ExistsByEmail(item.Email) {
			errorsList = append(errorsList, map[string]interface{}{
				"index":    i,
				"username": item.Username,
				"error":    "邮箱已被使用",
			})
			continue
		}

		hashedPassword, err := utils.HashPassword(item.Password)
		if err != nil {
			errorsList = append(errorsList, map[string]interface{}{
				"index":    i,
				"username": item.Username,
				"error":    "密码加密失败",
			})
			continue
		}

		role := item.Role
		if role == "" {
			role = "user"
		}
		if role != "user" && role != "admin" {
			errorsList = append(errorsList, map[string]interface{}{
				"index":    i,
				"username": item.Username,
				"error":    "无效的角色",
			})
			continue
		}

		user := &model.User{
			Username:     item.Username,
			Email:        item.Email,
			PasswordHash: hashedPassword,
			StudentID:    item.StudentID,
			Role:         role,
			Group:        item.Group,
		}

		if err := s.repo.Create(user); err != nil {
			errorsList = append(errorsList, map[string]interface{}{
				"index":    i,
				"username": item.Username,
				"error":    "创建用户失败",
			})
			continue
		}
		created++
	}

	return created, errorsList
}

// Login 用户登录
func (s *UserService) Login(req *model.UserLoginRequest) (*model.UserLoginResponse, error) {
	// 查找用户
	user, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, errors.New("登录失败")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成 Token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	return &model.UserLoginResponse{
		Token: token,
		User:  user.ToUserInfo(),
	}, nil
}

// GetProfile 获取用户信息
func (s *UserService) GetProfile(userID uint) (*model.UserInfo, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return user.ToUserInfo(), nil
}

// UpdateProfile 更新用户信息
func (s *UserService) UpdateProfile(userID uint, email, studentID string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 检查新邮箱是否与其他用户冲突
	if email != "" && email != user.Email {
		if s.repo.ExistsByEmail(email) {
			return errors.New("邮箱已被使用")
		}
		user.Email = email
	}

	if studentID != "" {
		user.StudentID = studentID
	}

	return s.repo.Update(user)
}

// GetRankList 获取排行榜
func (s *UserService) GetRankList(page, size int) ([]model.UserInfo, int64, error) {
	users, total, err := s.repo.List(page, size)
	if err != nil {
		return nil, 0, err
	}

	var result []model.UserInfo
	for _, u := range users {
		result = append(result, *u.ToUserInfo())
	}

	return result, total, nil
}

// GetUserList 获取用户列表（管理员）
func (s *UserService) GetUserList(page, size int) ([]model.User, int64, error) {
	return s.repo.List(page, size)
}

// SetUserRole 设置用户角色（管理员）
func (s *UserService) SetUserRole(userID uint, role string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if role != "user" && role != "admin" {
		return errors.New("无效的角色")
	}

	user.Role = role
	return s.repo.Update(user)
}

// UpdateUserByAdmin 管理员更新用户信息
func (s *UserService) UpdateUserByAdmin(userID uint, req *model.AdminUpdateUserRequest) (*model.User, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if len(username) < 3 || len(username) > 20 {
			return nil, errors.New("用户名长度应为 3-20")
		}
		if username != user.Username && s.repo.ExistsByUsername(username) {
			return nil, errors.New("用户名已存在")
		}
		user.Username = username
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email != "" {
			if _, err := mail.ParseAddress(email); err != nil {
				return nil, errors.New("邮箱格式不正确")
			}
			if email != user.Email && s.repo.ExistsByEmail(email) {
				return nil, errors.New("邮箱已被使用")
			}
		}
		user.Email = email
	}

	if req.StudentID != nil {
		user.StudentID = strings.TrimSpace(*req.StudentID)
	}

	if req.Group != nil {
		user.Group = strings.TrimSpace(*req.Group)
	}

	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		if role != "user" && role != "admin" {
			return nil, errors.New("无效的角色")
		}
		user.Role = role
	}

	if req.Password != nil {
		password := *req.Password
		if len(password) < 6 || len(password) > 20 {
			return nil, errors.New("密码长度应为 6-20")
		}
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return nil, errors.New("密码加密失败")
		}
		user.PasswordHash = hashedPassword
	}

	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("更新用户失败")
	}

	return user, nil
}

func normalizeAndValidateCreateUserRequest(req *model.AdminCreateUserRequest) error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.StudentID = strings.TrimSpace(req.StudentID)
	req.Role = strings.TrimSpace(req.Role)
	req.Group = strings.TrimSpace(req.Group)

	if len(req.Username) < 3 || len(req.Username) > 20 {
		return errors.New("用户名长度应为 3-20")
	}
	if len(req.Password) < 6 || len(req.Password) > 20 {
		return errors.New("密码长度应为 6-20")
	}
	if req.Email != "" {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			return errors.New("邮箱格式不正确")
		}
	}
	return nil
}
