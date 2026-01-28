package service

import (
	"context"
	"errors"
	"fmt"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

type PostService struct {
	postRepo        repository.PostRepository
	tagService      *TagService
	categoryService *CategoryService
}

func NewPostService(p repository.PostRepository, tsv *TagService, csv *CategoryService) *PostService {
	return &PostService{postRepo: p, tagService: tsv, categoryService: csv}
}

// AddPost 添加文章
func (s *PostService) AddPost(ctx context.Context, req *request.PostRequest) (*response.PostResponse, error) {
	// 检查别名是否重复
	p, err := s.PostBySlug(ctx, req.Slug, false)

	if err != nil {
		return nil, err
	}

	if p != nil {
		return nil, errors.New("别名 [" + req.Slug + "] 已存在")
	}

	// 检查标签和分类是否存在
	if err := s.checkTagAndCategoryExist(ctx, req); err != nil {
		return nil, err
	}

	// 检查是否需要自动生成摘要
	if err := s.autoGenerateExcerpt(ctx, req, false); err != nil {
		return nil, err
	}

	// 添加文章
	post, err := s.postRepo.AddPost(ctx, req)

	if err != nil {
		logger.Log.Error("添加文章失败", zap.Error(err))
		return nil, response.ServerError
	}

	return response.NewPostResponse(post), nil
}

// AddPostByNamesAndContents 根据名称和内容批量添加文章
//
// Parameters:
//   - ctx: 上下文
//   - names: 文章名称数组
//   - contents: 文章内容数组
//
// Returns:
//   - []*response.PostResponse: 添加成功的文章数组
func (s *PostService) AddPostByNamesAndContents(ctx context.Context, names []string, contents []string) ([]*response.PostResponse, error) {

	if len(names) == 0 || len(contents) == 0 {
		return nil, errors.New("名称或内容不能为空")
	}

	if len(names) != len(contents) {
		return nil, errors.New("名称和内容数组长度不一致")
	}

	var result []*response.PostResponse

	for i, name := range names {
		// 封装文章请求类，用于添加文章
		pr := request.NewPostRequestByNameAndContent(name, contents[i])

		// 检查别名是否重复
		p, err := s.PostBySlug(ctx, pr.Slug, false)
		if err != nil {
			return nil, err
		}

		if p != nil {
			// 别名重复，在别名后面加 _随机六位字符
			pr.Slug = pr.Slug + "_" + util.StringRandom(6)
		}

		// 检查是否需要自动生成摘要
		if err := s.autoGenerateExcerpt(ctx, pr, false); err != nil {
			return nil, err
		}

		// 添加文章
		ret, err := s.postRepo.AddPost(ctx, pr)
		if err != nil {
			logger.Log.Error("添加文章失败", zap.Error(err))
			return nil, response.ServerError
		}
		// 将添加成功的文章加到结果数组
		result = append(result, response.NewPostResponse(ret))
	}

	return result, nil
}

// DeletePosts 根据文章 ID 批量删除文章
func (s *PostService) DeletePosts(ctx context.Context, ids []uint) (bool, error) {

	if len(ids) == 0 {
		return false, nil
	}

	posts, err := s.PostByIds(ctx, ids, false)

	if err != nil {
		return false, err
	}

	for _, post := range posts {
		// 判断给定的文章是否都处于回收状态
		if post.Status != enum.PostStatusDeleted {
			return false, errors.New("只能删除处于回收站的文章")
		}
	}

	// 删除文章
	ret, err := s.postRepo.DeletePostByIds(ctx, ids)

	if err != nil {
		logger.Log.Error("删除文章失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdatePostStatusToDeleted 将文章状态修改为已删除（回收站）
func (s *PostService) UpdatePostStatusToDeleted(ctx context.Context, ids []uint) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	ret, err := s.postRepo.UpdatePostStatusToDeleted(ctx, ids)

	if err != nil {
		logger.Log.Error("修改文章状态失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdatePostStatusTo 将文章转为指定状态
func (s *PostService) UpdatePostStatusTo(ctx context.Context, ids []uint, status enum.PostStatus) (bool, error) {

	if len(ids) == 0 {
		return false, nil
	}

	ret, err := s.postRepo.UpdatePostStatusTo(ctx, ids, status)

	if err != nil {
		logger.Log.Error("修改文章状态失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdatePost 修改文章
func (s *PostService) UpdatePost(ctx context.Context, req *request.PostRequest) (bool, error) {

	if req == nil {
		logger.Log.Error("文章请求体不能为 nil")
		return false, response.ServerError
	}

	// 检查别名是否重复
	p, err := s.PostBySlug(ctx, req.Slug, false)
	if err != nil {
		return false, err
	}

	if p != nil && p.PostId != *req.PostId {
		// 别名重复，且不是相同文章
		return false, errors.New("别名 [" + req.Slug + "] 已存在")
	}

	// 检查标签和分类是否存在
	if err := s.checkTagAndCategoryExist(ctx, req); err != nil {
		return false, err
	}

	// 检查是否要自动生成摘要
	if err := s.autoGenerateExcerpt(ctx, req, true); err != nil {
		return false, err
	}

	// 修改文章
	ret, err := s.postRepo.UpdatePost(ctx, req)

	if err != nil {
		logger.Log.Error("修改文章失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdatePostStatus 修改文章状态（状态、可见性、置顶）
func (s *PostService) UpdatePostStatus(ctx context.Context, req *request.PostStatusRequest) (bool, error) {
	if req.Status == nil && req.Visible == nil && req.Pinned == nil {
		return false, nil
	}

	// 修改状态
	ret, err := s.postRepo.UpdatePostStatus(ctx, req)

	if err != nil {
		logger.Log.Error("修改文章状态失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdatePostExcerpt 修改文章摘要
func (s *PostService) UpdatePostExcerpt(ctx context.Context, postId uint, excerpt string) (bool, error) {
	ret, err := s.postRepo.UpdatePostExcerpt(ctx, postId, excerpt)

	if err != nil {
		logger.Log.Error("修改文章摘要失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// TryUpdatePostExcerptByPostContent 尝试通过文章正文修改文章摘要
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID
func (s *PostService) TryUpdatePostExcerptByPostContent(ctx context.Context, id uint) (bool, error) {
	// 先获取文章
	post, err := s.PostById(ctx, id, false)

	if err != nil {
		return false, err
	}

	// 判断当前文章是否需要自动生成摘要
	if post.AutoGenerateExcerpt {
		// 当前文章需要自动生成摘要
		// 获取文章正文内容
		content, err := s.PostContent(ctx, id, enum.PostContentStatusPublished, nil)

		if err != nil {
			return false, err
		}

		// 根据文章正文内容生成摘要
		excerpt := s.generateExcerptByString(content.Content, nil)

		// 更新文章摘要
		ret, err := s.UpdatePostExcerpt(ctx, id, excerpt)

		if err != nil {
			return false, err
		}

		return ret, nil
	}

	// 当前文章无需自动生成摘要
	return false, nil
}

// AddPostVisit 增加文章访问量
func (s *PostService) AddPostVisit(ctx context.Context, id uint) (bool, error) {
	ret, err := s.postRepo.AddPostVisit(ctx, id)

	if err != nil {
		logger.Log.Error("增加文章访问量失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// PostCount 获取文章总数
func (s *PostService) PostCount(ctx context.Context) (int64, error) {
	ret, err := s.postRepo.PostCount(ctx)

	if err != nil {
		logger.Log.Error("获取文章总数失败", zap.Error(err))
		return 0, response.ServerError
	}

	return ret, nil
}

// Posts 获取所有文章
//
// Parameters:
//   - ctx: 上下文
//   - includeTagAndCategory: 是否包含标签和分类（耗时操作，非必要不包含）
func (s *PostService) Posts(ctx context.Context, includeTagAndCategory bool) ([]*response.PostResponse, error) {
	posts, err := s.postRepo.Posts(ctx, includeTagAndCategory)
	if err != nil {
		logger.Log.Error("获取所有文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return posts, nil
}

// PostById 根据文章 ID 获取文章
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID
//   - includeTagAndCategory: 是否包含标签和分类（耗时操作，非必要不包含）
func (s *PostService) PostById(ctx context.Context, id uint, includeTagAndCategory bool) (*response.PostResponse, error) {
	post, err := s.postRepo.PostById(ctx, id, includeTagAndCategory)
	if err != nil {
		logger.Log.Error("获取文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return post, nil
}

// PostByIds 根据文章 ID 数组获取文章
//
// Parameters:
//   - ctx: 上下文
//   - ids: 文章 ID 数组
//   - includeTagAndCategory: 是否包含标签和分类（耗时操作，非必要不包含）
func (s *PostService) PostByIds(ctx context.Context, ids []uint, includeTagAndCategory bool) ([]*response.PostResponse, error) {

	if len(ids) == 0 {
		return []*response.PostResponse{}, nil
	}

	post, err := s.postRepo.PostByIds(ctx, ids, includeTagAndCategory)
	if err != nil {
		logger.Log.Error("获取文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return post, nil
}

// PostPager 分页获取所有文章
//
// Parameters:
//   - ctx: 上下文
//   - page: 当前页码
//   - size: 每页条数
//   - status: 文章状态
//   - visible: 文章可见性
//   - tagId: 标签 ID
//   - categoryId: 分类 ID
//   - sort: 文章排序
func (s *PostService) PostPager(
	ctx context.Context,
	page int, size int,
	status *enum.PostStatus,
	visible *enum.PostVisible,
	key *string,
	tagId *uint,
	categoryId *uint,
	sort *enum.PostSort,
) (*models.Pager[response.PostResponse], error) {
	pager, err := s.postRepo.PostPager(ctx, page, size, status, visible, key, tagId, categoryId, sort)
	if err != nil {
		logger.Log.Error("分页获取文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return pager, nil
}

// PostBySlug 根据文章别名获取文章
//
// Parameters:
//   - ctx: 上下文
//   - slug: 文章别名
//   - includeTagAndCategory: 是否包含标签和分类（耗时操作，非必要不包含）
func (s *PostService) PostBySlug(ctx context.Context, slug string, includeTagAndCategory bool) (*response.PostResponse, error) {
	post, err := s.postRepo.PostBySlug(ctx, slug, includeTagAndCategory)
	if err != nil {
		logger.Log.Error("获取文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return post, nil
}

// PostByKey 根据关键字获取文章
//
// Parameters:
//   - ctx: 上下文
//   - key: 关键字（标题、别名、摘要、内容）
func (s *PostService) PostByKey(ctx context.Context, key string) ([]*response.PostResponse, error) {
	posts, err := s.postRepo.PostByKey(ctx, key)
	if err != nil {
		logger.Log.Error("获取文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return posts, nil
}

// ApiPosts 获取文章 API 接口，用于博客前端页面，不包含敏感信息
//
// Parameters:
//   - ctx: 上下文
//   - page: 当前页码
//   - size: 每页条数
//   - tagId: 标签 ID
//   - categoryId: 分类 ID
//   - tag: 标签名或别名
//   - category: 分类名或别名
func (s *PostService) ApiPosts(
	ctx context.Context,
	page, size int,
	key *string,
	tagId *uint,
	categoryId *uint,
	tag *string,
	category *string,
) (*models.Pager[response.PostApiResponse], error) {
	posts, err := s.postRepo.PostApi(ctx, page, size, key, tagId, categoryId, tag, category)
	if err != nil {
		logger.Log.Error("获取文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	return posts, nil
}

// PostContents 获取文章所有内容
func (s *PostService) PostContents(ctx context.Context, id uint) ([]*response.PostContentResponse, error) {
	// 判断文章是否存在
	exist, err := s.isPostExist(ctx, id)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errors.New("文章 [" + strconv.Itoa(int(id)) + "] 不存在")
	}

	contents, err := s.postRepo.PostContents(ctx, id)

	if err != nil {
		logger.Log.Error("获取文章内容失败", zap.Error(err))
		return nil, response.ServerError
	}
	return contents, nil
}

// PostContent 获取文章内容
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID
//   - status: 文章内容状态
//   - draftName: 草稿名称
func (s *PostService) PostContent(ctx context.Context, id uint, status enum.PostContentStatus, draftName *string) (*models.PostContent, error) {
	content, err := s.postRepo.PostContent(ctx, id, status, draftName)
	if err != nil {
		logger.Log.Error("获取文章内容失败", zap.Error(err))
		return nil, response.ServerError
	}
	return content, nil
}

// ApiPostContent 获取文章博客 API 接口，用于博客前端页面获取文章内容。
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID（ID 和别名至少存在一个）
//   - slug: 文章别名（ID 和别名至少存在一个）
//   - password: 文章密码（如果有）
func (s *PostService) ApiPostContent(ctx context.Context, id *uint, slug *string, password *string) (*response.PostContentApiResponse, error) {
	if id == nil && slug == nil {
		return nil, nil
	}

	var post *response.PostResponse
	if id != nil {
		// 文章 ID 不为空
		p, err := s.PostById(ctx, *id, true)
		if err != nil {
			return nil, err
		}
		post = p
	} else {
		// 文章别名不为空
		p, err := s.PostBySlug(ctx, *slug, true)
		if err != nil {
			return nil, err
		}
		post = p
	}

	if post == nil {
		return nil, nil
	}

	if post.Status != enum.PostStatusPublished {
		// 文章未发布
		return nil, nil
	}

	// 判断文章是否有密码
	if post.Encrypted {
		// 文章有密码
		if password == nil {
			// 接口没有提供密码
			return nil, nil
		}

		// 验证密码是否正确
		valid, err := s.isPostPasswordValid(ctx, post.PostId, *password)

		if err != nil {
			return nil, err
		}

		if !valid {
			// 密码错误
			return nil, errors.New("文章密码不正确")
		}
	}

	// 获取文章正文
	content, err := s.PostContent(ctx, post.PostId, enum.PostContentStatusPublished, nil)

	if err != nil {
		return nil, err
	}

	// 增加文章浏览量
	go func() {
		bCtx := context.Background()
		_, _ = s.AddPostVisit(bCtx, post.PostId)
	}()

	// 封装博客 API 文章内容响应体
	return &response.PostContentApiResponse{
		Post:    *response.NewPostApiResponse(post, false),
		Content: content.Content,
	}, nil
}

// AddPostDraft 添加文章草稿
func (s *PostService) AddPostDraft(ctx context.Context, req *request.PostDraftRequest) (*models.PostContent, error) {
	// 先判断文章是否存在
	exist, err := s.isPostExist(ctx, req.PostId)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, errors.New("文章 [" + strconv.Itoa(int(req.PostId)) + "] 不存在")
	}

	// 判断草稿名是否已存在
	exist, err = s.isPostDraftNameExist(ctx, req.PostId, req.DraftName)
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, errors.New("草稿名 [" + req.DraftName + "] 已经存在")
	}

	// 添加草稿
	ret, err := s.postRepo.AddPostDraft(ctx, req.PostId, req.Content, req.DraftName)
	if err != nil {
		logger.Log.Error("添加文章草稿失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// DeletePostContent 删除文章内容
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID
//   - status: 文章内容状态
//   - draftNames: 草稿名数组
func (s *PostService) DeletePostContent(
	ctx context.Context,
	id uint,
	status enum.PostContentStatus,
	draftNames []string,
) (bool, error) {
	ret, err := s.postRepo.DeletePostContent(ctx, id, status, draftNames)

	if err != nil {
		logger.Log.Error("删除文章内容失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdatePostContent 修改文章内容
//
// Parameters:
//   - ctx: 上下文
//   - pc: 文章内容请求体
//   - status: 文章内容状态
//   - draftName: 草稿名
func (s *PostService) UpdatePostContent(
	ctx context.Context,
	pc request.PostContentRequest,
	status enum.PostContentStatus,
	draftName *string,
) (bool, error) {
	ret, err := s.postRepo.UpdatePostContent(ctx, pc, status, draftName)
	if err != nil {
		logger.Log.Error("修改文章内容失败", zap.Error(err))
		return false, response.ServerError
	}

	if ret && status == enum.PostContentStatusPublished {
		// 文章内容修改成功，并且当前修改的是文章正文，尝试更新文章摘要
		_, _ = s.TryUpdatePostExcerptByPostContent(ctx, pc.PostId)
	}

	return ret, nil
}

// UpdatePostDraftName 修改文章草稿名
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID
//   - oldName: 旧草稿名
//   - newName: 新草稿名
func (s *PostService) UpdatePostDraftName(ctx context.Context, id uint, oldName string, newName string) (bool, error) {
	if oldName == newName {
		return false, nil
	}

	// 先判断新的草稿名是否已经存在
	exist, err := s.isPostDraftNameExist(ctx, id, newName)
	if err != nil {
		return false, err
	}
	if exist {
		return false, errors.New("草稿名 [" + newName + "] 已存在")
	}

	ret, err := s.postRepo.UpdatePostDraftName(ctx, id, oldName, newName)
	if err != nil {
		logger.Log.Error("修改文章草稿名失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// UpdatePostDraftToContent 将文章草稿转换为文章正文
//
// Parameters:
//   - ctx: 上下文
//   - id: 文章 ID
//   - draftName: 草稿名
//   - deleteContent: 是否删除原来的正文
//   - contentName: 文章正文名，留空将默认使用被转换为正文的旧草稿名。
func (s *PostService) UpdatePostDraftToContent(
	ctx context.Context,
	id uint,
	draftName string,
	deleteContent bool,
	contentName *string,
) (bool, error) {
	// 先判断文章是否存在
	exist, err := s.isPostExist(ctx, id)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, errors.New("文章 [" + strconv.Itoa(int(id)) + "] 不存在")
	}

	// 判断草稿名是否存在
	exist, err = s.isPostDraftNameExist(ctx, id, draftName)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, errors.New("草稿名 [" + draftName + "] 不存在")
	}

	ret, err := s.postRepo.UpdatePostDraftToContent(ctx, id, draftName, deleteContent, contentName)
	if err != nil {
		logger.Log.Error("将文章草稿转换为文章正文失败", zap.Error(err))
		return false, response.ServerError
	}

	if ret {
		// 草稿转换成功，尝试更新文章摘要
		_, _ = s.TryUpdatePostExcerptByPostContent(ctx, id)
		// 修改文章最后修改时间
		_, err = s.postRepo.UpdatePostLastModifyTime(ctx, id, util.Int64Ptr(time.Now().UnixMilli()))
		if err != nil {
			logger.Log.Error("修改文章最后修改时间失败", zap.Error(err))
		}
	}

	return ret, nil
}

// isPostPasswordValid 验证文章密码是否正确
func (s *PostService) isPostPasswordValid(ctx context.Context, id uint, password string) (bool, error) {
	valid, err := s.postRepo.IsPostPasswordValid(ctx, id, password)

	if err != nil {
		logger.Log.Error("验证文章密码失败", zap.Error(err))
		return false, response.ServerError
	}
	return valid, nil
}

// ExportPosts 导出所有文章
func (s *PostService) ExportPosts(ctx context.Context) (*response.ExportPostResponse, error) {
	// 总处理文章数量
	var totalCount = 0
	// 成功的文章数量
	var successCount = 0
	// 失败原因切片
	var failResult []string

	// 当前时间字符串
	nowTime := util.FormatDate(time.Now())

	// 文件夹名前缀
	filePrefix := fmt.Sprintf("%s_Post", nowTime)

	// 临时文件夹地址
	tempDir := fmt.Sprintf("./.nola/temp/%s", filePrefix)

	if util.IsDirExist(tempDir) {
		_ = os.RemoveAll(tempDir)
	}

	// 先获取所有未删除的文章
	posts, err := s.postRepo.Posts(ctx, true)
	if err != nil {
		logger.Log.Error("获取所有文章失败", zap.Error(err))
		return nil, response.ServerError
	}
	posts = util.Filter(posts, func(post *response.PostResponse) bool {
		return post.Status != enum.PostStatusDeleted
	})

	type postResult struct {
		isSuccess bool
		errMsg    string
	}

	resultChan := make(chan postResult, len(posts))
	var wg sync.WaitGroup

	for _, post := range posts {
		wg.Add(1)
		// 每个文章启动一个协程
		go func(p *response.PostResponse) {
			defer wg.Done()

			// 当前文章的所有内容列表（包括正文和草稿）
			items, err := s.postRepo.PostContents(ctx, p.PostId)
			if err != nil {
				logger.Log.Error("获取文章内容列表失败", zap.Error(err))
				resultChan <- postResult{
					isSuccess: false,
					errMsg:    fmt.Sprintf("获取文章 [%s] 内容列表失败", post.Title),
				}
				return
			}

			// 写出文章元数据
			err = s.postMetaDataToTempDir(post, tempDir)
			if err != nil {
				logger.Log.Error("写出文章元数据失败", zap.Error(err))
				resultChan <- postResult{
					isSuccess: false,
					errMsg:    fmt.Sprintf("写出文章 [%s] 元数据失败", post.Title),
				}
				return
			}

			for _, item := range items {
				// 获取当前文章具体内容
				content, err := s.postRepo.PostContent(ctx, item.PostId, *item.Status, item.DraftName)
				if err != nil {
					logger.Log.Error("获取文章内容失败", zap.Error(err))
					resultChan <- postResult{
						isSuccess: false,
						errMsg:    fmt.Sprintf("获取文章 [%s] 内容失败", post.Title),
					}
					continue
				}

				if content == nil {
					if *item.Status == enum.PostContentStatusPublished {
						resultChan <- postResult{
							isSuccess: false,
							errMsg:    fmt.Sprintf("[%s] 文章没有任何内容", post.Title),
						}
					} else {
						resultChan <- postResult{
							isSuccess: false,
							errMsg:    fmt.Sprintf("[%s] 文章的 [%s] 草稿没有任何内容", post.Title, *item.DraftName),
						}
					}
					continue
				}

				// 将当前文章内容写到临时文件夹
				_, err = s.postToTempDir(post, content, tempDir)
				if err != nil {
					resultChan <- postResult{
						isSuccess: false,
						errMsg:    err.Error(),
					}
				} else {
					resultChan <- postResult{
						isSuccess: true,
					}
				}
			}

		}(post)
	}

	// 当所有生产者执行完成后，关闭通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 处理文章导出结果
	for res := range resultChan {
		totalCount++
		if res.isSuccess {
			successCount++
		} else {
			failResult = append(failResult, res.errMsg)
		}
	}

	backupDir := "./.nola/backup"
	// 检查备份目录是否存在
	if !util.IsDirExist(backupDir) {
		err := os.MkdirAll(backupDir, 0755)
		if err != nil {
			logger.Log.Error("创建备份目录失败", zap.Error(err))
			return nil, response.ServerError
		}
	}

	// 将存储文章内容的临时文件夹压缩
	err = util.CreateFolderZip(tempDir, fmt.Sprintf("./.nola/backup/%s.zip", filePrefix))
	if err != nil {
		logger.Log.Error("创建压缩文件失败", zap.Error(err))
		return nil, errors.New("压缩文件失败")
	}

	// 删除临时文件夹
	err = os.RemoveAll(tempDir)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("删除临时文件夹 [%s] 失败", tempDir), zap.Error(err))
	}

	return &response.ExportPostResponse{
		SuccessCount: successCount,
		FailCount:    len(failResult),
		FailResult:   util.DefaultEmptySlice(failResult),
		Path:         fmt.Sprintf("/backup/%s.zip", filePrefix),
		Count:        totalCount,
	}, nil
}

// MostViewedPost 获取浏览量最多的文章
func (s *PostService) MostViewedPost(ctx context.Context) (*response.PostResponse, error) {
	post, err := s.postRepo.MostViewedPost(ctx)
	if err != nil {
		logger.Log.Error("获取浏览量最多的文章失败", zap.Error(err))
		return nil, err
	}
	return post, nil
}

// PostVisitCount 获取文章总浏览量
func (s *PostService) PostVisitCount(ctx context.Context) (int64, error) {
	count, err := s.postRepo.PostVisitCount(ctx)
	if err != nil {
		logger.Log.Error("获取文章总浏览量失败", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// postToTempDir 将文章内容写入临时文件夹
//
// Parameters:
//   - post: 文章信息
//   - content: 文章内容
//   - tempDir: 临时文件夹地址
func (s *PostService) postToTempDir(post *response.PostResponse, content *models.PostContent, tempDir string) (bool, error) {
	// 临时文件夹是否存在
	if !util.IsDirExist(tempDir) {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			logger.Log.Error("创建临时文件夹失败", zap.Error(err))
			return false, errors.New("创建临时文件夹失败")
		}
	}

	// 判断当前文章目录是否存在（文章文件夹名：<文章名>__<文章别名>）
	postDir := filepath.Join(tempDir, fmt.Sprintf("%s__%s", post.Title, post.Slug))
	if !util.IsDirExist(postDir) {
		err := os.MkdirAll(postDir, 0755)
		if err != nil {
			logger.Log.Error("创建文章文件夹失败", zap.Error(err))
			return false, errors.New(fmt.Sprintf("[%s] 创建文章文件夹失败", post.Title))
		}
	}

	// 当前内容输出的文件夹
	outputDir := postDir

	// 判断当前文章内容类型
	switch content.Status {
	case enum.PostContentStatusPublished:
		// 文章正文，检查正文文件夹是否存在
		outputDir = filepath.Join(outputDir, "content")
		if !util.IsDirExist(outputDir) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				logger.Log.Error("创建文章正文内容文件夹失败", zap.Error(err))
				return false, errors.New(fmt.Sprintf("[%s] 创建文章正文内容文件夹失败", post.Title))
			}
		}
	case enum.PostContentStatusDraft:
		// 文章草稿
		outputDir = filepath.Join(outputDir, "draft")
		if !util.IsDirExist(outputDir) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				logger.Log.Error("创建文章草稿内容文件夹失败", zap.Error(err))
				return false, errors.New(fmt.Sprintf("[%s] 创建文章草稿内容文件夹失败", post.Title))
			}
		}
	}

	// 构建文件名
	postFileName := util.StringDefault(content.DraftName, "content")

	// 要写入的文件地址
	writerPath := filepath.Join(outputDir, postFileName+".md")

	// 写入文件
	err := os.WriteFile(writerPath, []byte(content.Content), 0644)
	if err != nil {
		logger.Log.Error("写入临时文件失败", zap.Error(err))
		return false, errors.New("写入临时文件失败")
	}

	return true, nil
}

// postMetaDataToTempDir 将文章元数据写入临时文件夹
//
// Parameters:
//   - post: 文章信息
//   - tempDir: 临时文件夹地址
func (s *PostService) postMetaDataToTempDir(post *response.PostResponse, tempDir string) error {

	if post == nil {
		return errors.New("文章信息不能为空")
	}

	if util.IsDirExist(tempDir) {
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			logger.Log.Error("创建临时文件夹失败", zap.Error(err))
			return errors.New("创建临时文件夹失败")
		}
	}

	// 判断当前文章目录是否存在（文章文件夹名：<文章名>__<文章别名>）
	postDir := filepath.Join(tempDir, fmt.Sprintf("%s__%s", post.Title, post.Slug))
	if !util.IsDirExist(postDir) {
		err := os.MkdirAll(postDir, 0755)
		if err != nil {
			logger.Log.Error("创建文章文件夹失败", zap.Error(err))
			return errors.New(fmt.Sprintf("[%s] 创建文章文件夹失败", post.Title))
		}
	}

	// 输出地址
	outputPath := filepath.Join(postDir, "metadata.json")

	// 文章源数据
	metadata := response.NewPostMetaData(*post)

	json := util.ToJsonString(metadata)

	// 写出元数据
	if err := os.WriteFile(outputPath, []byte(util.StringDefault(json, "")), 0644); err != nil {
		logger.Log.Error("写出文章元数据失败", zap.Error(err))
		return errors.New("写出文章元数据失败")
	}

	return nil
}

// checkTagAndCategoryExist 检查标签和分类是否存在
func (s *PostService) checkTagAndCategoryExist(ctx context.Context, req *request.PostRequest) error {
	// 检查传来的标签 ID 是否都存在
	if len(req.TagIds) > 0 {
		// 检查标签
		nonExistIds, err := s.tagService.isIdsExist(ctx, req.TagIds)
		if err != nil {
			return err
		}
		if len(nonExistIds) > 0 {
			return errors.New(fmt.Sprintf("标签 %v 不存在", nonExistIds))
		}
	}

	if req.CategoryId != nil {
		// 检查分类
		c, err := s.categoryService.CategoryById(ctx, *req.CategoryId)
		if err != nil {
			return err
		}
		if c == nil {
			return errors.New("分类 [" + strconv.Itoa(int(*req.CategoryId)) + "] 不存在")
		}
	}
	return nil
}

// autoGenerateExcerpt 自动生成摘要
//
// Parameters:
//   - ctx: 上下文
//   - req: 文章请求体
//   - isUpdate: 是否是修改文章（true 从数据库获取当前文章内容，false 从 req 获取文章内容）
func (s *PostService) autoGenerateExcerpt(ctx context.Context, req *request.PostRequest, isUpdate bool) error {
	if req == nil || *req.AutoGenerateExcerpt == false {
		// 不用自动生成摘要
		return nil
	}

	var content string
	if isUpdate {
		// 当前是修改文章，就从数据库获取文章内容
		c, err := s.PostContent(ctx, *req.PostId, enum.PostContentStatusPublished, nil)
		if err != nil {
			return err
		}
		content = c.Content
	} else {
		// 当前是添加文章，直接从 pr 对象中获取文章内容
		content = *req.Content
	}

	if util.StringIsBlank(content) {
		// 内容为空，摘要也为空
		req.Excerpt = util.StringPtr("")
	} else {
		// 生成摘要
		req.Excerpt = util.StringPtr(s.generateExcerptByString(content, nil))
	}

	return nil
}

// generateExcerptByString 根据一段 Markdown / PlainText 生成摘要
//
// Parameters:
//   - content: Markdown / PlainText
//   - length: 摘要长度，默认 100
func (s *PostService) generateExcerptByString(content string, length *int) string {
	if util.StringIsBlank(content) {
		return ""
	}

	l := 100
	if length != nil {
		l = *length
	}

	// 将 Markdown 转为纯本文
	excerpt := util.MarkdownToPlainText(content)
	if len(excerpt) > l {
		// 超过长度，截取
		excerpt = excerpt[:l]
	}
	return excerpt
}

// isPostExist 判断文章是否存在
func (s *PostService) isPostExist(ctx context.Context, id uint) (bool, error) {
	post, err := s.PostById(ctx, id, false)
	if err != nil {
		return false, err
	}
	return post != nil, nil
}

// isPostDraftNameExist 判断草稿名是否存在
func (s *PostService) isPostDraftNameExist(ctx context.Context, id uint, draftName string) (bool, error) {
	content, err := s.PostContent(ctx, id, enum.PostContentStatusDraft, util.StringPtr(draftName))
	if err != nil {
		return false, err
	}
	return content != nil, nil
}
