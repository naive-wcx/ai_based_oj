package repository

import (
	"oj-system/internal/model"

	"gorm.io/gorm"
)

type ProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository() *ProblemRepository {
	return &ProblemRepository{db: DB}
}

// Create 创建题目
func (r *ProblemRepository) Create(problem *model.Problem) error {
	return r.db.Create(problem).Error
}

// GetByID 根据 ID 获取题目
func (r *ProblemRepository) GetByID(id uint) (*model.Problem, error) {
	var problem model.Problem
	if err := r.db.First(&problem, id).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

// GetByIDs 根据 ID 列表获取题目
func (r *ProblemRepository) GetByIDs(ids []uint) ([]model.Problem, error) {
	if len(ids) == 0 {
		return []model.Problem{}, nil
	}
	var problems []model.Problem
	if err := r.db.Where("id IN ?", ids).Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

// Update 更新题目
func (r *ProblemRepository) Update(problem *model.Problem) error {
	return r.db.Save(problem).Error
}

// Delete 删除题目
func (r *ProblemRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联的测试用例
		if err := tx.Where("problem_id = ?", id).Delete(&model.Testcase{}).Error; err != nil {
			return err
		}
		// 删除题目
		return tx.Delete(&model.Problem{}, id).Error
	})
}

// List 获取题目列表
func (r *ProblemRepository) List(page, size int, difficulty string, tag string, keyword string) ([]model.Problem, int64, error) {
	var problems []model.Problem
	var total int64

	query := r.db.Model(&model.Problem{}).Where("is_public = ?", true)

	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id ASC").Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, total, nil
}

// ListAll 获取题目列表（包含隐藏题）
func (r *ProblemRepository) ListAll(page, size int, difficulty string, tag string, keyword string) ([]model.Problem, int64, error) {
	var problems []model.Problem
	var total int64

	query := r.db.Model(&model.Problem{})

	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id ASC").Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, total, nil
}

// GetTestcases 获取题目的测试用例
func (r *ProblemRepository) GetTestcases(problemID uint) ([]model.Testcase, error) {
	var testcases []model.Testcase
	if err := r.db.Where("problem_id = ?", problemID).Order("order_num ASC").Find(&testcases).Error; err != nil {
		return nil, err
	}
	return testcases, nil
}

// CreateTestcase 创建测试用例
func (r *ProblemRepository) CreateTestcase(testcase *model.Testcase) error {
	return r.db.Create(testcase).Error
}

// DeleteTestcases 删除题目的所有测试用例
func (r *ProblemRepository) DeleteTestcases(problemID uint) error {
	return r.db.Where("problem_id = ?", problemID).Delete(&model.Testcase{}).Error
}

// IncrementSubmitCount 增加题目提交数
func (r *ProblemRepository) IncrementSubmitCount(problemID uint) error {
	return r.db.Model(&model.Problem{}).Where("id = ?", problemID).
		UpdateColumn("submit_count", gorm.Expr("submit_count + 1")).Error
}

// IncrementAcceptedCount 增加题目通过数
func (r *ProblemRepository) IncrementAcceptedCount(problemID uint) error {
	return r.db.Model(&model.Problem{}).Where("id = ?", problemID).
		UpdateColumn("accepted_count", gorm.Expr("accepted_count + 1")).Error
}

// CountPublic 获取公开题目数量
func (r *ProblemRepository) CountPublic() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Problem{}).Where("is_public = ?", true).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
