package service

import (
	"errors"

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

// Register 用户注册
func (s *UserService) Register(req *model.UserRegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if s.repo.ExistsByUsername(req.Username) {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if s.repo.ExistsByEmail(req.Email) {
		return nil, errors.New("邮箱已被注册")
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		StudentID:    req.StudentID,
		Role:         "user",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	return user, nil
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
