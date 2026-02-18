package service

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"oj-system/internal/config"
	"oj-system/internal/model"
	"oj-system/internal/repository"
)

type ProblemService struct {
	repo *repository.ProblemRepository
	submissionRepo *repository.SubmissionRepository
	contestRepo *repository.ContestRepository
	userRepo *repository.UserRepository
	participationRepo *repository.ContestParticipationRepository
}

func NewProblemService() *ProblemService {
	return &ProblemService{
		repo: repository.NewProblemRepository(),
		submissionRepo: repository.NewSubmissionRepository(),
		contestRepo: repository.NewContestRepository(),
		userRepo: repository.NewUserRepository(),
		participationRepo: repository.NewContestParticipationRepository(),
	}
}

// Create 创建题目
func (s *ProblemService) Create(req *model.ProblemCreateRequest, createdBy uint) (*model.Problem, error) {
	fileEnabled, inputName, outputName, err := normalizeFileIO(req)
	if err != nil {
		return nil, err
	}

	problem := &model.Problem{
		Title:         req.Title,
		Description:   req.Description,
		InputFormat:   req.InputFormat,
		OutputFormat:  req.OutputFormat,
		Hint:          req.Hint,
		Samples:       req.Samples,
		TimeLimit:     req.TimeLimit,
		MemoryLimit:   req.MemoryLimit,
		Difficulty:    req.Difficulty,
		Tags:          req.Tags,
		AIJudgeConfig: req.AIJudgeConfig,
		FileIOEnabled: fileEnabled,
		FileInputName: inputName,
		FileOutputName: outputName,
		IsPublic:      req.IsPublic,
		CreatedBy:     createdBy,
	}

	// 设置默认值
	if problem.TimeLimit == 0 {
		problem.TimeLimit = 1000
	}
	if problem.MemoryLimit == 0 {
		problem.MemoryLimit = 256
	}
	if problem.IsPublic == nil {
		t := true
		problem.IsPublic = &t
	}

	if err := s.repo.Create(problem); err != nil {
		return nil, errors.New("创建题目失败")
	}

	// 创建题目数据目录
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problem.ID))
	os.MkdirAll(problemDir, 0755)

	return problem, nil
}

// GetByID 获取题目详情
func (s *ProblemService) GetByID(id uint) (*model.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("题目不存在")
	}
	return problem, nil
}

// GetByIDWithUser 获取题目详情（含通过标识）
func (s *ProblemService) GetByIDWithUser(id uint, userID uint, isAdmin bool) (*model.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("题目不存在")
	}
	isHidden := problem.IsPublic == nil || !*problem.IsPublic
	if isHidden && !isAdmin {
		if !s.canAccessHiddenProblem(problem.ID, userID) {
			return nil, errors.New("无权限访问该题目")
		}
		if s.shouldHideHiddenProblemTags(problem.ID, userID, time.Now()) {
			problem.Tags = model.StringList{}
		}
	}
	if userID > 0 {
		hasAccepted := s.submissionRepo.HasAccepted(userID, id)
		problem.HasAccepted = hasAccepted
	}
	return problem, nil
}

// Update 更新题目
func (s *ProblemService) Update(id uint, req *model.ProblemCreateRequest) (*model.Problem, error) {
	problem, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("题目不存在")
	}

	fileEnabled, inputName, outputName, err := normalizeFileIO(req)
	if err != nil {
		return nil, err
	}

	problem.Title = req.Title
	problem.Description = req.Description
	problem.InputFormat = req.InputFormat
	problem.OutputFormat = req.OutputFormat
	problem.Hint = req.Hint
	problem.Samples = req.Samples
	problem.TimeLimit = req.TimeLimit
	problem.MemoryLimit = req.MemoryLimit
	problem.Difficulty = req.Difficulty
	problem.Tags = req.Tags
	problem.AIJudgeConfig = req.AIJudgeConfig
	problem.FileIOEnabled = fileEnabled
	problem.FileInputName = inputName
	problem.FileOutputName = outputName
	problem.IsPublic = req.IsPublic

	if err := s.repo.Update(problem); err != nil {
		return nil, errors.New("更新题目失败")
	}

	return problem, nil
}

// Delete 删除题目
func (s *ProblemService) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("题目不存在")
	}

	// 删除题目数据目录
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", id))
	os.RemoveAll(problemDir)

	return s.repo.Delete(id)
}

// List 获取题目列表
func (s *ProblemService) List(page, size int, difficulty, tag, keyword string) ([]model.ProblemListItem, int64, error) {
	problems, total, err := s.repo.List(page, size, difficulty, tag, keyword)
	if err != nil {
		return nil, 0, err
	}

	var items []model.ProblemListItem
	for _, p := range problems {
		hasAI := p.AIJudgeConfig != nil && p.AIJudgeConfig.Enabled
		items = append(items, model.ProblemListItem{
			ID:            p.ID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Tags:          p.Tags,
			SubmitCount:   p.SubmitCount,
			AcceptedCount: p.AcceptedCount,
			HasAIJudge:    hasAI,
			HasFileIO:     p.FileIOEnabled,
			IsPublic:      p.IsPublic,
		})
	}

	return items, total, nil
}

// ListWithUser 获取题目列表（含通过标识）
func (s *ProblemService) ListWithUser(page, size int, difficulty, tag, keyword string, userID uint, isAdmin bool) ([]model.ProblemListItem, int64, error) {
	var (
		problems []model.Problem
		total    int64
		err      error
	)
	if isAdmin {
		problems, total, err = s.repo.ListAll(page, size, difficulty, tag, keyword)
	} else {
		problems, total, err = s.repo.List(page, size, difficulty, tag, keyword)
	}
	if err != nil {
		return nil, 0, err
	}

	problemIDs := make([]uint, 0, len(problems))
	for _, p := range problems {
		problemIDs = append(problemIDs, p.ID)
	}

	acceptedSet := map[uint]struct{}{}
	if userID > 0 {
		acceptedIDs, err := s.submissionRepo.GetAcceptedProblemIDs(userID, problemIDs)
		if err == nil {
			for _, id := range acceptedIDs {
				acceptedSet[id] = struct{}{}
			}
		}
	}

	var items []model.ProblemListItem
	for _, p := range problems {
		hasAI := p.AIJudgeConfig != nil && p.AIJudgeConfig.Enabled
		_, hasAccepted := acceptedSet[p.ID]
		items = append(items, model.ProblemListItem{
			ID:            p.ID,
			Title:         p.Title,
			Difficulty:    p.Difficulty,
			Tags:          p.Tags,
			SubmitCount:   p.SubmitCount,
			AcceptedCount: p.AcceptedCount,
			HasAIJudge:    hasAI,
			HasFileIO:     p.FileIOEnabled,
			HasAccepted:   hasAccepted,
			IsPublic:      p.IsPublic,
		})
	}

	return items, total, nil
}

func (s *ProblemService) canAccessHiddenProblem(problemID uint, userID uint) bool {
	if userID == 0 {
		return false
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}
	role := strings.ToLower(strings.TrimSpace(user.Role))
	if role == model.RoleAdmin || role == model.RoleSuperAdmin {
		return true
	}

	contests, err := s.contestRepo.ListAll()
	if err != nil {
		return false
	}
	now := time.Now()
	for _, contest := range contests {
		if now.Before(contest.StartAt) {
			continue
		}
		if !containsUint([]uint(contest.ProblemIDs), problemID) {
			continue
		}
		if !canAccessContest(&contest, userID, user.Group) {
			continue
		}

		if normalizeContestTimingMode(contest.TimingMode) != contestTimingWindow {
			return true
		}
		if !now.Before(contest.EndAt) {
			return true
		}

		participation, err := s.participationRepo.GetByContestAndUser(contest.ID, userID)
		if err == nil && participation != nil && !now.Before(participation.StartAt) {
			return true
		}
	}
	return false
}

func (s *ProblemService) shouldHideHiddenProblemTags(problemID uint, userID uint, now time.Time) bool {
	if userID == 0 {
		return false
	}
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false
	}
	role := strings.ToLower(strings.TrimSpace(user.Role))
	if role == model.RoleAdmin || role == model.RoleSuperAdmin {
		return false
	}

	contests, err := s.contestRepo.ListAll()
	if err != nil {
		return false
	}
	for _, contest := range contests {
		if now.Before(contest.StartAt) || !now.Before(contest.EndAt) {
			continue
		}
		if !containsUint([]uint(contest.ProblemIDs), problemID) {
			continue
		}
		if !canAccessContest(&contest, userID, user.Group) {
			continue
		}
		return true
	}
	return false
}

// GetTestcases 获取测试用例
func (s *ProblemService) GetTestcases(problemID uint) ([]model.Testcase, error) {
	return s.repo.GetTestcases(problemID)
}

// AddTestcase 添加测试用例
func (s *ProblemService) AddTestcase(problemID uint, inputReader, outputReader io.Reader, score int, isSample bool) error {
	// 确保题目存在
	_, err := s.repo.GetByID(problemID)
	if err != nil {
		return errors.New("题目不存在")
	}

	// 获取当前测试用例数量作为序号
	testcases, _ := s.repo.GetTestcases(problemID)
	orderNum := len(testcases) + 1

	// 创建测试数据目录
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID))
	os.MkdirAll(problemDir, 0755)

	// 保存输入文件
	inputFile := filepath.Join(problemDir, fmt.Sprintf("%d.in", orderNum))
	inputData, err := io.ReadAll(inputReader)
	if err != nil {
		return errors.New("读取输入数据失败")
	}
	if err := os.WriteFile(inputFile, inputData, 0644); err != nil {
		return errors.New("保存输入文件失败")
	}

	// 保存输出文件
	outputFile := filepath.Join(problemDir, fmt.Sprintf("%d.out", orderNum))
	outputData, err := io.ReadAll(outputReader)
	if err != nil {
		return errors.New("读取输出数据失败")
	}
	if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
		return errors.New("保存输出文件失败")
	}

	// 创建测试用例记录
	testcase := &model.Testcase{
		ProblemID:  problemID,
		InputFile:  inputFile,
		OutputFile: outputFile,
		Score:      score,
		IsSample:   isSample,
		OrderNum:   orderNum,
	}

	return s.repo.CreateTestcase(testcase)
}

// DeleteTestcases 删除所有测试用例
func (s *ProblemService) DeleteTestcases(problemID uint) error {
	// 删除文件
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID))
	files, _ := filepath.Glob(filepath.Join(problemDir, "*.in"))
	for _, f := range files {
		os.Remove(f)
	}
	files, _ = filepath.Glob(filepath.Join(problemDir, "*.out"))
	for _, f := range files {
		os.Remove(f)
	}

	return s.repo.DeleteTestcases(problemID)
}

func (s *ProblemService) PrepareProblemRejudge(problemID uint) ([]model.Submission, error) {
	if _, err := s.repo.GetByID(problemID); err != nil {
		return nil, errors.New("题目不存在")
	}

	submissions, err := s.submissionRepo.ListRejudgeCandidatesByProblem(problemID)
	if err != nil {
		return nil, errors.New("获取提交记录失败")
	}
	if len(submissions) == 0 {
		return []model.Submission{}, nil
	}

	ids := make([]uint, 0, len(submissions))
	for _, submission := range submissions {
		ids = append(ids, submission.ID)
	}
	if err := s.submissionRepo.ResetForRejudge(ids); err != nil {
		return nil, errors.New("重置提交状态失败")
	}

	return submissions, nil
}

func (s *ProblemService) UploadProblemImage(problemID uint, originalName string, reader io.Reader) (string, string, string, error) {
	if _, err := s.repo.GetByID(problemID); err != nil {
		return "", "", "", errors.New("题目不存在")
	}
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(originalName)))
	if !isAllowedProblemImageExt(ext) {
		return "", "", "", errors.New("仅支持 png/jpg/jpeg/gif/webp/bmp 图片")
	}

	imageDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID), "images")
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return "", "", "", errors.New("创建图片目录失败")
	}

	nameToken, err := randomHex(6)
	if err != nil {
		return "", "", "", errors.New("生成图片文件名失败")
	}
	savedName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), nameToken, ext)
	savePath := filepath.Join(imageDir, savedName)

	output, err := os.Create(savePath)
	if err != nil {
		return "", "", "", errors.New("保存图片失败")
	}
	defer output.Close()

	const maxImageSize = 10 << 20 // 10MB
	written, err := io.Copy(output, io.LimitReader(reader, maxImageSize+1))
	if err != nil {
		_ = os.Remove(savePath)
		return "", "", "", errors.New("写入图片失败")
	}
	if written > maxImageSize {
		_ = os.Remove(savePath)
		return "", "", "", errors.New("图片大小不能超过 10MB")
	}

	url := fmt.Sprintf("/api/v1/problem/%d/image/%s", problemID, savedName)
	alt := strings.TrimSpace(strings.TrimSuffix(filepath.Base(originalName), ext))
	if alt == "" {
		alt = "image"
	}
	markdown := fmt.Sprintf("![%s](%s)", alt, url)
	return url, markdown, savedName, nil
}

func (s *ProblemService) ResolveProblemImagePath(problemID uint, filename string) (string, error) {
	if _, err := s.repo.GetByID(problemID); err != nil {
		return "", errors.New("题目不存在")
	}
	name := strings.TrimSpace(filename)
	if name == "" || name != filepath.Base(name) || strings.Contains(name, "..") {
		return "", errors.New("图片文件名不合法")
	}
	path := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID), "images", name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", errors.New("图片不存在")
	}
	return path, nil
}

func normalizeFileIO(req *model.ProblemCreateRequest) (bool, string, string, error) {
	if req == nil || !req.FileIOEnabled {
		return false, "", "", nil
	}

	inputName := strings.TrimSpace(req.FileInputName)
	outputName := strings.TrimSpace(req.FileOutputName)

	if inputName == "" || outputName == "" {
		return false, "", "", errors.New("文件操作已启用，请填写输入/输出文件名")
	}
	if inputName == outputName {
		return false, "", "", errors.New("输入/输出文件名不能相同")
	}
	if err := validateFileName(inputName, ".in"); err != nil {
		return false, "", "", err
	}
	if err := validateFileName(outputName, ".out"); err != nil {
		return false, "", "", err
	}

	return true, inputName, outputName, nil
}

// UploadTestcaseZip 批量上传测试用例 (Zip)
func (s *ProblemService) UploadTestcaseZip(problemID uint, zipPath string) error {
	// 1. 打开 Zip 文件
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return errors.New("无法打开 zip 文件")
	}
	defer r.Close()

	// 2. 扫描文件，寻找配对
	// Map baseName -> {input: file, output: file}
	type pair struct {
		Input  *zip.File
		Output *zip.File
	}
	pairs := make(map[string]*pair)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// 忽略隐藏文件
		if strings.HasPrefix(f.Name, ".") || strings.HasPrefix(filepath.Base(f.Name), ".") {
			continue
		}

		ext := filepath.Ext(f.Name)
		// 获取去掉扩展名后的路径作为 key
		base := strings.TrimSuffix(f.Name, ext)
		
		if _, ok := pairs[base]; !ok {
			pairs[base] = &pair{}
		}

		if ext == ".in" {
			pairs[base].Input = f
		} else if ext == ".out" || ext == ".ans" {
			pairs[base].Output = f
		}
	}

	// 3. 筛选有效配对
	var validPairs []*pair
	for _, p := range pairs {
		if p.Input != nil && p.Output != nil {
			validPairs = append(validPairs, p)
		}
	}

	if len(validPairs) == 0 {
		return errors.New("未找到匹配的输入输出文件 (.in + .out/.ans)")
	}

	// 4. 排序 (尝试按文件名中的数字排序)
	sort.Slice(validPairs, func(i, j int) bool {
		return compareFileNames(validPairs[i].Input.Name, validPairs[j].Input.Name)
	})

	// 5. 删除旧数据
	if err := s.DeleteTestcases(problemID); err != nil {
		return err
	}

	// 6. 保存新数据
	problemDir := filepath.Join(config.GlobalConfig.Paths.Problems, fmt.Sprintf("%d", problemID))
	os.MkdirAll(problemDir, 0755)

	// 计算每个测试点的分数
	scorePerCase := 100 / len(validPairs)
	if scorePerCase == 0 {
		scorePerCase = 1
	}

	for i, p := range validPairs {
		orderNum := i + 1
		
		// 复制 Input
		inputFile := filepath.Join(problemDir, fmt.Sprintf("%d.in", orderNum))
		if err := extractZipFile(p.Input, inputFile); err != nil {
			return err
		}

		// 复制 Output
		outputFile := filepath.Join(problemDir, fmt.Sprintf("%d.out", orderNum))
		if err := extractZipFile(p.Output, outputFile); err != nil {
			return err
		}

		// 最后一个测试点补齐分数
		score := scorePerCase
		if i == len(validPairs)-1 {
			score = 100 - scorePerCase*(len(validPairs)-1)
		}

		// 创建记录
		testcase := &model.Testcase{
			ProblemID:  problemID,
			InputFile:  inputFile,
			OutputFile: outputFile,
			Score:      score,
			IsSample:   false,
			OrderNum:   orderNum,
		}
		if err := s.repo.CreateTestcase(testcase); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	w, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, rc)
	return err
}

// 简单的文件名数字排序比较
func compareFileNames(a, b string) bool {
	// 提取数字正则
	re := regexp.MustCompile(`(\d+)`)
	numsA := re.FindAllString(a, -1)
	numsB := re.FindAllString(b, -1)

	if len(numsA) > 0 && len(numsB) > 0 {
		// 比较最后一个数字（通常是序号）
		nA := numsA[len(numsA)-1]
		nB := numsB[len(numsB)-1]
		
		// 如果数字长度不同，长的更大
		if len(nA) != len(nB) {
			return len(nA) < len(nB)
		}
		// 长度相同，字典序比较
		if nA != nB {
			return nA < nB
		}
	}
	return a < b
}

func validateFileName(name string, suffix string) error {
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return errors.New("文件名不合法")
	}
	if !strings.HasSuffix(name, suffix) {
		return errors.New("文件名需以 " + suffix + " 结尾")
	}
	return nil
}

func isAllowedProblemImageExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func randomHex(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("invalid random length")
	}
	buffer := make([]byte, n)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
