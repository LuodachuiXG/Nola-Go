package repository

import (
	"context"
	"errors"
	"fmt"
	"nola-go/internal/db"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/util"
	"time"

	"gorm.io/gorm"
)

// PostRepository 文章 Repo 接口
type PostRepository interface {
	// AddPost 添加文章
	AddPost(ctx context.Context, req *request.PostRequest) (*models.Post, error)
	// DeletePostByIds 根据文章 ID 数组删除文章
	DeletePostByIds(ctx context.Context, ids []uint) (bool, error)
	// UpdatePostStatusToDeleted 修改文章状态为删除
	UpdatePostStatusToDeleted(ctx context.Context, ids []uint) (bool, error)
	// UpdatePostStatusTo 修改文章状态为指定状态
	UpdatePostStatusTo(ctx context.Context, ids []uint, status enum.PostStatus) (bool, error)
	// UpdatePost 修改文章
	UpdatePost(ctx context.Context, req *request.PostRequest) (bool, error)
	// UpdatePostStatus 修改文章状态（状态、可见性、置顶）
	UpdatePostStatus(ctx context.Context, req *request.PostStatusRequest) (bool, error)
	// UpdatePostExcerpt 修改文章摘要
	UpdatePostExcerpt(ctx context.Context, postId uint, excerpt string) (bool, error)
	// UpdatePostLastModifyTime 修改文章最后修改时间
	UpdatePostLastModifyTime(ctx context.Context, postId uint, time *int64) (bool, error)
	// AddPostVisit 增加文章访问量
	AddPostVisit(ctx context.Context, id uint) (bool, error)
	// PostCount 获取文章总数
	PostCount(ctx context.Context) (int64, error)
	// Posts 获取所有文章
	Posts(ctx context.Context, includeTagAndCategory bool) ([]*response.PostResponse, error)
	// PostById 根据文章 ID 获取文章
	PostById(ctx context.Context, id uint, includeTagAndCategory bool) (*response.PostResponse, error)
	// PostByIds 根据文章 ID 数组获取文章
	PostByIds(ctx context.Context, ids []uint, includeTagAndCategory bool) ([]*response.PostResponse, error)
	// PostPager 分页获取所有文章
	PostPager(
		ctx context.Context,
		page int, size int,
		status *enum.PostStatus,
		visible *enum.PostVisible,
		key *string,
		tagId *uint,
		categoryId *uint,
		sort *enum.PostSort,
	) (*models.Pager[response.PostResponse], error)
	// PostBySlug 根据文章别名获取文章
	PostBySlug(ctx context.Context, slug string, includeTagAndCategory bool) (*response.PostResponse, error)
	// PostByKey 根据文章关键字获取文章（标题、别名、摘要、内容）
	PostByKey(ctx context.Context, key string) ([]*response.PostResponse, error)
	// PostApi 获取文章 Api 接口（博客前端）
	PostApi(
		ctx context.Context,
		page int, size int,
		key *string,
		tagId *uint,
		categoryId *uint,
		tag *string,
		category *string,
	) (*models.Pager[response.PostApiResponse], error)
	// PostContents 获取文章所有内容（包括正文和草稿）
	PostContents(ctx context.Context, postId uint) ([]*response.PostContentResponse, error)
	// PostContent 获取文章内容
	PostContent(
		ctx context.Context,
		postId uint,
		status enum.PostContentStatus,
		draftName *string,
	) (*models.PostContent, error)
	// AddPostDraft 添加文章草稿
	AddPostDraft(ctx context.Context, postId uint, content string, draftName string) (*models.PostContent, error)
	// DeletePostContent 删除文章内容
	DeletePostContent(ctx context.Context, postId uint, status enum.PostContentStatus, draftNames []string) (bool, error)
	// UpdatePostContent 修改文章内容
	UpdatePostContent(
		ctx context.Context,
		pc request.PostContentRequest,
		status enum.PostContentStatus,
		draftName *string,
	) (bool, error)
	// UpdatePostDraftName 修改文章草稿名
	UpdatePostDraftName(ctx context.Context, postId uint, oldName string, newName string) (bool, error)

	// UpdatePostDraftToContent 将文章草稿转为文章正文
	//   - ctx: Content
	//   - postId: 文章 ID
	//   - draftName: 草稿名
	//   - deleteContent: 是否删除原来的正文
	//   - contentName: 文章正文名，留 nil 将默认使用被转换为正文的旧草稿名
	UpdatePostDraftToContent(ctx context.Context, postId uint, draftName string, deleteContent bool, contentName *string) (bool, error)
	// IsPostPasswordValid 验证文章密码是否正确
	IsPostPasswordValid(ctx context.Context, postId uint, password string) (bool, error)
	// MostViewedPost 浏览量最多的文章
	MostViewedPost(ctx context.Context) (*response.PostResponse, error)
	// PostVisitCount 文章总浏览量
	PostVisitCount(ctx context.Context) (int64, error)
}

type postRepo struct {
	db           *gorm.DB
	tagRepo      TagRepository
	categoryRepo CategoryRepository
}

func NewPostRepository(db *gorm.DB, tagRepo TagRepository, categoryRepo CategoryRepository) PostRepository {
	return &postRepo{db: db, tagRepo: tagRepo, categoryRepo: categoryRepo}
}

// AddPost 添加文章
func (r *postRepo) AddPost(ctx context.Context, req *request.PostRequest) (*models.Post, error) {
	currentTime := time.Now().UnixMilli()

	// 开启事务
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var pwd *string
	if !util.StringIsNilOrBlank(req.Password) {
		// 密码不为空
		pwd = util.StringPtr(util.GenerateHash(*req.Password))
	} else {
		pwd = nil
	}

	post := models.Post{
		Title:               req.Title,
		AutoGenerateExcerpt: *req.AutoGenerateExcerpt,
		Excerpt:             util.StringDefault(req.Excerpt, ""),
		Slug:                req.Slug,
		Cover:               req.Cover,
		AllowComment:        *req.AllowComment,
		Pinned:              req.Pinned,
		Status:              req.Status,
		Visible:             req.Visible,
		Password:            pwd,
		CreateTime:          currentTime,
		LastModifyTime:      nil,
	}

	// 插入文章
	if err := tx.Model(&models.Post{}).Create(&post).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 新插入的文章 ID
	postId := post.PostId

	if len(req.TagIds) > 0 {
		// 插入文章标签
		var tags []*models.PostTag
		for _, tagId := range req.TagIds {
			tags = append(tags, &models.PostTag{
				PostId: postId,
				TagId:  tagId,
			})
		}

		err := tx.Create(&tags).Error

		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if req.CategoryId != nil {
		// 插入文章分类
		err := tx.Create(&models.PostCategory{
			PostId:     postId,
			CategoryId: *req.CategoryId,
		}).Error

		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	content := util.StringDefault(req.Content, "")

	// 插入文章内容
	err := tx.Create(&models.PostContent{
		PostId:         postId,
		Content:        content,
		HTML:           util.MarkdownToHtml(content),
		Status:         enum.PostContentStatusPublished,
		LastModifyTime: util.Int64Ptr(currentTime),
	}).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &post, nil
}

// DeletePostByIds 根据文章 ID 数组删除文章
func (r *postRepo) DeletePostByIds(ctx context.Context, ids []uint) (success bool, err error) {

	if len(ids) == 0 {
		return false, nil
	}

	tx := r.db.WithContext(ctx).Begin()
	defer handlePanic(tx)(&success, &err)

	// 删除文章内容
	err = tx.Where("post_id IN ?", ids).Delete(&models.PostContent{}).Error
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除文章分类
	err = tx.Where("post_id IN ?", ids).Delete(&models.PostCategory{}).Error
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除文章标签
	err = tx.Where("post_id IN ?", ids).Delete(&models.PostTag{}).Error
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除文章
	ret := tx.Where("post_id IN ?", ids).Delete(&models.Post{})
	if err := ret.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// UpdatePostStatusToDeleted 修改文章状态为删除
func (r *postRepo) UpdatePostStatusToDeleted(ctx context.Context, ids []uint) (bool, error) {
	return r.UpdatePostStatusTo(ctx, ids, enum.PostStatusDeleted)
}

// UpdatePostStatusTo 修改文章状态为指定状态
func (r *postRepo) UpdatePostStatusTo(ctx context.Context, ids []uint, status enum.PostStatus) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}

	ret := r.db.WithContext(ctx).Model(&models.Post{}).Where("post_id IN ?", ids).Update("status", status)
	if err := ret.Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// UpdatePost 修改文章
func (r *postRepo) UpdatePost(ctx context.Context, req *request.PostRequest) (success bool, err error) {

	if req.PostId == nil {
		return false, errors.New("文章 ID 不能为 nil")
	}

	tx := r.db.WithContext(ctx).Begin()
	defer handlePanic(tx)(&success, &err)

	// 删除文章标签
	err = tx.Where("post_id = ?", req.PostId).Delete(&models.PostTag{}).Error
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// 删除文章分类
	err = tx.Where("post_id = ?", req.PostId).Delete(&models.PostCategory{}).Error
	if err != nil {
		tx.Rollback()
		return false, err
	}

	if len(req.TagIds) > 0 {
		// 插入文章标签
		var tags []*models.PostTag
		for _, tagId := range req.TagIds {
			tags = append(tags, &models.PostTag{
				PostId: *req.PostId,
				TagId:  tagId,
			})
		}
		err := tx.Create(&tags).Error
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	if req.CategoryId != nil {
		// 插入文章分类
		err := tx.Create(&models.PostCategory{
			PostId:     *req.PostId,
			CategoryId: *req.CategoryId,
		}).Error
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	// 更新文章
	newPost := map[string]any{
		"title":                 req.Title,
		"auto_generate_excerpt": req.AutoGenerateExcerpt,
		"excerpt":               util.StringDefault(req.Excerpt, ""),
		"slug":                  req.Slug,
		"allow_comment":         req.AllowComment,
		"status":                req.Status,
		"visible":               req.Visible,
		"cover":                 req.Cover,
		"pinned":                req.Pinned,
	}

	if !util.StringIsNilOrBlank(req.Password) && req.Encrypted != nil && *req.Encrypted == true {
		// 设置新密码
		newPost["password"] = util.GenerateHash(*req.Password)
	} else if req.Encrypted != nil && *req.Encrypted == false {
		// 移除旧密码
		newPost["password"] = nil
	}

	ret := tx.Model(&models.Post{}).Where("post_id = ?", req.PostId).Updates(newPost)
	if err := ret.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// UpdatePostStatus 修改文章状态（状态、可见性、置顶）
func (r *postRepo) UpdatePostStatus(ctx context.Context, req *request.PostStatusRequest) (bool, error) {
	if req.Status == nil && req.Visible == nil && &req.Pinned == nil {
		return false, nil
	}

	updates := map[string]any{}

	if req.Status != nil {
		updates["status"] = req.Status
	}
	if req.Visible != nil {
		updates["visible"] = req.Visible
	}
	if req.Pinned != nil {
		updates["pinned"] = req.Pinned
	}

	ret := r.db.WithContext(ctx).Model(&models.Post{}).Where("post_id = ?", req.PostId).Updates(updates)
	if err := ret.Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// UpdatePostExcerpt 修改文章摘要
func (r *postRepo) UpdatePostExcerpt(ctx context.Context, postId uint, excerpt string) (bool, error) {
	ret := r.db.WithContext(ctx).Model(&models.Post{}).Where("post_id = ?", postId).Update("excerpt", excerpt)
	if err := ret.Error; err != nil {
		return false, err
	}
	return ret.RowsAffected > 0, nil
}

// UpdatePostLastModifyTime 修改文章最后修改时间
func (r *postRepo) UpdatePostLastModifyTime(ctx context.Context, postId uint, time *int64) (bool, error) {
	ret := r.db.WithContext(ctx).Model(&models.Post{}).Where("post_id = ?", postId).Update("last_modify_time", time)
	if err := ret.Error; err != nil {
		return false, err
	}
	return ret.RowsAffected > 0, nil
}

// AddPostVisit 增加文章访问量
func (r *postRepo) AddPostVisit(ctx context.Context, id uint) (bool, error) {
	ret := r.db.WithContext(ctx).Model(&models.Post{}).Where("post_id = ?", id).Update("visit", gorm.Expr("visit + ?", 1))
	if err := ret.Error; err != nil {
		return false, err
	}
	return ret.RowsAffected > 0, nil
}

// PostCount 获取文章总数
func (r *postRepo) PostCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Post{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Posts 获取所有文章
func (r *postRepo) Posts(ctx context.Context, includeTagAndCategory bool) ([]*response.PostResponse, error) {

	// 先获取所有文章
	var posts []*models.Post
	err := r.db.WithContext(ctx).Order("create_time DESC").Find(&posts).Error
	if err != nil {
		return nil, err
	}

	// 将文章转为文章响应体
	postRes := response.NewPostResponses(posts)

	if includeTagAndCategory {
		// 获取标签和分类
		if err := r.fillTagAndCategory(ctx, postRes); err != nil {
			return nil, err
		}
	}
	return postRes, nil
}

// PostById 根据文章 ID 获取文章
func (r *postRepo) PostById(ctx context.Context, id uint, includeTagAndCategory bool) (*response.PostResponse, error) {
	var post *models.Post
	err := r.db.WithContext(ctx).Where("post_id = ?", id).First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	postRes := response.NewPostResponse(post)

	if includeTagAndCategory {
		if err := r.fillTagAndCategory(ctx, []*response.PostResponse{postRes}); err != nil {
			return nil, err
		}
	}

	return postRes, nil
}

// PostByIds 根据文章 ID 数组获取文章
func (r *postRepo) PostByIds(ctx context.Context, ids []uint, includeTagAndCategory bool) ([]*response.PostResponse, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var posts []*models.Post
	err := r.db.WithContext(ctx).Where("post_id IN ?", ids).Order("create_time DESC").Find(&posts).Error
	if err != nil {
		return nil, err
	}

	postRes := response.NewPostResponses(posts)

	if includeTagAndCategory {
		if err := r.fillTagAndCategory(ctx, postRes); err != nil {
			return nil, err
		}
	}

	return postRes, nil
}

// PostPager 分页获取所有文章
func (r *postRepo) PostPager(
	ctx context.Context,
	page int, size int,
	status *enum.PostStatus,
	visible *enum.PostVisible,
	key *string,
	tagId *uint,
	categoryId *uint,
	sort *enum.PostSort,
) (*models.Pager[response.PostResponse], error) {

	// 构建查询语句
	base, err := r.sqlQueryPosts(ctx, status, visible, key, tagId, categoryId, nil, nil, sort)

	if err != nil {
		return nil, err
	}

	if page == 0 {
		// 获取所有文章
		var posts []*models.Post
		err := base.Find(&posts).Error
		if err != nil {
			return nil, err
		}

		// 转为文章响应体
		postRes := response.NewPostResponses(posts)

		// 填充标签和分类
		err = r.fillTagAndCategory(ctx, postRes)
		if err != nil {
			return nil, err
		}

		return &models.Pager[response.PostResponse]{
			Page:       0,
			Size:       0,
			Data:       postRes,
			TotalData:  int64(len(postRes)),
			TotalPages: 1,
		}, nil
	}

	pager, err := db.PagerBuilder[models.Post](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return base
	})

	if err != nil {
		return nil, err
	}

	// 转为文章响应体
	postRes := response.NewPostResponses(pager.Data)

	// 填充标签和分类s
	err = r.fillTagAndCategory(ctx, postRes)
	if err != nil {
		return nil, err
	}

	return &models.Pager[response.PostResponse]{
		Page:       pager.Page,
		Size:       pager.Size,
		Data:       postRes,
		TotalData:  pager.TotalData,
		TotalPages: pager.TotalPages,
	}, nil
}

// PostBySlug 根据文章别名获取文章
func (r *postRepo) PostBySlug(ctx context.Context, slug string, includeTagAndCategory bool) (*response.PostResponse, error) {
	var post *models.Post
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	postRes := response.NewPostResponse(post)

	if includeTagAndCategory {
		// 填充标签和分类
		err = r.fillTagAndCategory(ctx, []*response.PostResponse{postRes})
	}

	if err != nil {
		return nil, err
	}

	return postRes, nil
}

// PostByKey 根据文章关键字获取文章（标题、别名、摘要、内容）
func (r *postRepo) PostByKey(ctx context.Context, key string) ([]*response.PostResponse, error) {
	var posts []*models.Post

	query := r.db.WithContext(ctx).
		Table("post p").
		Joins("LEFT JOIN post_content pc ON p.post_id = pc.post_id")
	// 关键词查询
	err := r.sqlQueryKey(query, key).Order("p.create_time DESC").Find(&posts).Error
	if err != nil {
		return nil, err
	}

	return response.NewPostResponses(posts), nil
}

// PostApi 获取文章 Api 接口（博客前端）
func (r *postRepo) PostApi(
	ctx context.Context,
	page int, size int,
	key *string,
	tagId *uint,
	categoryId *uint,
	tag *string,
	category *string,
) (*models.Pager[response.PostApiResponse], error) {

	// 构建查询语句
	base, err := r.sqlQueryPosts(
		ctx,
		enum.PostStatusPtr(enum.PostStatusPublished),
		enum.PostVisiblePtr(enum.PostVisibleVisible),
		key, tagId, categoryId,
		tag, category,
		enum.PostSortPtr(enum.PostSortPinned),
	)

	if err != nil {
		return nil, err
	}

	if page == 0 {
		// 获取所有文章
		var posts []*models.Post
		err := base.Find(&posts).Error
		if err != nil {
			return nil, err
		}

		// 转为文章响应体
		postRes := response.NewPostResponses(posts)

		// 填充标签和分类
		err = r.fillTagAndCategory(ctx, postRes)
		if err != nil {
			return nil, err
		}

		return &models.Pager[response.PostApiResponse]{
			Page:       0,
			Size:       0,
			Data:       response.NewPostApiResponses(postRes, true),
			TotalData:  int64(len(postRes)),
			TotalPages: 1,
		}, nil
	}

	pager, err := db.PagerBuilder[models.Post](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return base
	})

	if err != nil {
		return nil, err
	}

	// 转为文章响应体
	postRes := response.NewPostResponses(pager.Data)

	// 填充标签和分类
	err = r.fillTagAndCategory(ctx, postRes)
	if err != nil {
		return nil, err
	}

	return &models.Pager[response.PostApiResponse]{
		Page:       pager.Page,
		Size:       pager.Size,
		Data:       response.NewPostApiResponses(postRes, true),
		TotalData:  pager.TotalData,
		TotalPages: pager.TotalPages,
	}, nil
}

// PostContents 获取文章所有内容（包括正文和草稿）
func (r *postRepo) PostContents(ctx context.Context, postId uint) ([]*response.PostContentResponse, error) {
	var contents []*models.PostContent
	err := r.db.WithContext(ctx).
		// 不包含内容
		Select("post_content_id, post_id, status, draft_name, last_modify_time").
		Where("post_id = ?", postId).
		Find(&contents).Error

	if err != nil {
		return nil, err
	}

	return response.NewPostContentResponses(contents), nil
}

// PostContent 获取文章内容
func (r *postRepo) PostContent(
	ctx context.Context,
	postId uint,
	status enum.PostContentStatus,
	draftName *string,
) (*models.PostContent, error) {
	var content *models.PostContent
	query := r.db.WithContext(ctx).Where("post_id = ?", postId)
	if status == enum.PostContentStatusDraft {
		query = query.Where("status = ? AND draft_name = ?", enum.PostContentStatusDraft, draftName)
	} else {
		query = query.Where("status = ?", enum.PostContentStatusPublished)
	}
	err := query.First(&content).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return content, nil
}

// AddPostDraft 添加文章草稿
func (r *postRepo) AddPostDraft(ctx context.Context, postId uint, content string, draftName string) (*models.PostContent, error) {

	draft := models.PostContent{
		PostId:         postId,
		Content:        content,
		DraftName:      util.StringPtr(draftName),
		Status:         enum.PostContentStatusDraft,
		LastModifyTime: util.Int64Ptr(time.Now().UnixMilli()),
		HTML:           util.MarkdownToHtml(content),
	}

	err := r.db.WithContext(ctx).Create(&draft).Error

	if err != nil {
		return nil, err
	}

	return &draft, nil
}

// DeletePostContent 删除文章内容
func (r *postRepo) DeletePostContent(ctx context.Context, postId uint, status enum.PostContentStatus, draftNames []string) (bool, error) {
	query := r.db.WithContext(ctx).Where("post_id = ? AND status = ?", postId, status)

	if status == enum.PostContentStatusDraft && draftNames != nil && len(draftNames) > 0 {
		// 删除的是草稿，并且草稿名不为 nil
		query = query.Where("draft_name IN ?", draftNames)
	}

	ret := query.Delete(&models.PostContent{})

	if err := ret.Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// UpdatePostContent 修改文章内容
func (r *postRepo) UpdatePostContent(
	ctx context.Context,
	pc request.PostContentRequest,
	status enum.PostContentStatus,
	draftName *string,
) (bool, error) {
	currentTime := time.Now().UnixMilli()
	query := r.db.WithContext(ctx).
		Model(&models.PostContent{}).
		Where("post_id = ? AND status = ?", pc.PostId, status)

	if status == enum.PostContentStatusDraft && draftName != nil {
		query = query.Where("draft_name = ?", *draftName)
	}

	updates := map[string]any{
		"content":          pc.Content,
		"html":             util.MarkdownToHtml(pc.Content),
		"last_modify_time": currentTime,
	}

	ret := query.Updates(updates)

	if err := ret.Error; err != nil {
		return false, err
	}

	// 文章内容修改成功，并且修改的是正文内容
	if status == enum.PostContentStatusPublished {
		// 修改文章最后修改时间
		_, _ = r.UpdatePostLastModifyTime(ctx, pc.PostId, util.Int64Ptr(currentTime))
	}

	return ret.RowsAffected > 0, nil
}

// UpdatePostDraftName 修改文章草稿名
func (r *postRepo) UpdatePostDraftName(ctx context.Context, postId uint, oldName string, newName string) (bool, error) {
	ret := r.db.WithContext(ctx).
		Model(&models.PostContent{}).
		Where("post_id = ? AND draft_name = ?", postId, oldName).
		Update("draft_name", newName)

	if err := ret.Error; err != nil {
		return false, err
	}
	return ret.RowsAffected > 0, nil
}

// UpdatePostDraftToContent 将文章草稿转为文章正文
//   - ctx: Content
//   - postId: 文章 ID
//   - draftName: 草稿名
//   - deleteContent: 是否删除原来的正文
//   - contentName: 文章正文名，留 nil 将默认使用被转换为正文的旧草稿名
func (r *postRepo) UpdatePostDraftToContent(ctx context.Context, postId uint, draftName string, deleteContent bool, contentName *string) (success bool, err error) {

	// 获取原来的文章正文内容 ID
	var postContentId uint
	err = r.db.WithContext(ctx).
		Model(&models.PostContent{}).
		Select("post_content_id").
		Where("post_id = ? AND status = ?", postId, enum.PostContentStatusPublished).
		Scan(&postContentId).Error

	if err != nil {
		return false, err
	}

	// 获取要转换的草稿的内容 ID
	var postContentDraftId uint
	err = r.db.WithContext(ctx).
		Model(&models.PostContent{}).
		Select("post_content_id").
		Where("post_id = ? AND status = ? AND draft_name = ?", postId, enum.PostContentStatusDraft, draftName).
		Scan(&postContentDraftId).Error

	if err != nil {
		return false, err
	}

	tx := r.db.WithContext(ctx).Begin()
	defer handlePanic(tx)(&success, &err)

	if deleteContent {
		// 删除原来的正文
		err := tx.Where("post_content_id = ?", postContentId).Delete(&models.PostContent{}).Error
		if err != nil {
			tx.Rollback()
			return false, err
		}
	} else {
		// 不删除原来的正文

		updates := map[string]any{
			"status": enum.PostContentStatusDraft,
		}

		if util.StringIsNilOrBlank(contentName) {
			updates["draft_name"] = draftName
		} else {
			updates["draft_name"] = contentName
		}

		// 修改原来的正文内容为草稿，并修改草稿名
		err := tx.Model(&models.PostContent{}).
			Where("post_content_id = ?", postContentId).
			Updates(updates).Error

		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	// 将指定的草稿转为正文
	ret := tx.Model(&models.PostContent{}).
		Where("post_content_id = ?", postContentDraftId).
		Updates(map[string]any{
			"status":     enum.PostContentStatusPublished,
			"draft_name": nil,
		})

	if err := ret.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return ret.RowsAffected > 0, nil
}

// IsPostPasswordValid 验证文章密码是否正确
func (r *postRepo) IsPostPasswordValid(ctx context.Context, postId uint, password string) (bool, error) {
	var count int64
	ret := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("post_id = ? AND password = ?", postId, util.GenerateHash(password)).
		Count(&count)

	if err := ret.Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// MostViewedPost 浏览量最多的文章
func (r *postRepo) MostViewedPost(ctx context.Context) (*response.PostResponse, error) {
	var post *models.Post
	err := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Where("visit >= ?", 0).
		Order("visit DESC").
		Limit(1).
		Scan(&post).Error
	if err != nil {
		return nil, err
	}

	// 转为文章响应类
	res := response.NewPostResponse(post)

	// 填充标签和分类
	err = r.fillTagAndCategory(ctx, []*response.PostResponse{res})

	if err != nil {
		return nil, err
	}

	return res, nil
}

// PostVisitCount 文章总浏览量
func (r *postRepo) PostVisitCount(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Post{}).
		Select("SUM(visit)").
		Scan(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// fillTagAndCategory 给文章填充标签和分类
func (r *postRepo) fillTagAndCategory(ctx context.Context, posts []*response.PostResponse) error {
	if len(posts) == 0 {
		return nil
	}

	for _, post := range posts {
		// 获取文章标签
		tags, err := r.tagRepo.TagByPostId(ctx, post.PostId)
		if err != nil {
			return err
		}
		post.Tags = tags

		// 获取文章分类
		category, err := r.categoryRepo.CategoryByPostId(ctx, post.PostId)
		if err != nil {
			return err
		}
		post.Category = category
	}

	return nil
}

// sqlQueryPosts 构建文章查询 SQL
func (r *postRepo) sqlQueryPosts(
	ctx context.Context,
	status *enum.PostStatus,
	visible *enum.PostVisible,
	key *string,
	tagId *uint,
	categoryId *uint,
	tag *string,
	category *string,
	sort *enum.PostSort,
) (*gorm.DB, error) {
	query := r.db.WithContext(ctx).
		Table("post p").
		Joins("LEFT JOIN post_content pc ON p.post_id = pc.post_id AND pc.status = ?", enum.PostContentStatusPublished)

	// 文章状态
	if status != nil {
		query = query.Where("p.status = ?", status)
	} else {
		// 默认不获取已在回收站中的文章
		query = query.Where("p.status != ?", enum.PostStatusDeleted)
	}

	// 文章可见性
	if visible != nil {
		query = query.Where("p.visible = ?", visible)
	}

	// 关键词
	if !util.StringIsNilOrBlank(key) {
		query = r.sqlQueryKey(query, *key)
	}

	// 文章标签 ID
	if tagId != nil {
		// 获取与当前标签匹配的文章 ID 数组
		var postTags []*models.PostTag
		err := r.db.WithContext(ctx).Where("tag_id = ?", tagId).Find(&postTags).Error
		if err != nil {
			return nil, err
		}
		// 匹配的文章 ID
		postIds := make([]uint, len(postTags))
		for i, postTag := range postTags {
			postIds[i] = postTag.PostId
		}
		// 添加查询条件
		query = query.Where("p.post_id IN ?", postIds)
	}

	// 文章分类
	if categoryId != nil {
		// 获取与当前分类匹配的文章 ID 数组
		var postCategories []*models.PostCategory
		err := r.db.WithContext(ctx).Where("category_id = ?", categoryId).Find(&postCategories).Error
		if err != nil {
			return nil, err
		}
		// 匹配的文章 ID
		postIds := make([]uint, len(postCategories))
		for i, postCategory := range postCategories {
			postIds[i] = postCategory.PostId
		}
		// 添加查询条件
		query = query.Where("p.post_id IN ?", postIds)
	}

	// 标签名或别名
	if !util.StringIsNilOrBlank(tag) {
		// 获取与当前标签匹配的文章 ID 集合
		var postIds []uint
		err := r.db.WithContext(ctx).
			Table("post_tag pt").
			Joins("LEFT JOIN tag t ON pt.tag_id = t.tag_id").
			Where("t.display_name = ? OR (t.slug = ?)", tag, tag).
			Pluck("pt.post_id", &postIds).Error
		if err != nil {
			return nil, err
		}
		query = query.Where("p.post_id IN ?", postIds)
	}

	// 分类名或别名
	if !util.StringIsNilOrBlank(category) {
		// 获取与当前分类匹配的文章 ID 集合
		var postIds []uint
		err := r.db.WithContext(ctx).
			Table("post_category pc").
			Joins("LEFT JOIN category c ON pc.category_id = c.category_id").
			Where("c.display_name = ? OR (c.slug = ?)", category, category).
			Pluck("pc.post_id", &postIds).Error
		if err != nil {
			return nil, err
		}
		query = query.Where("p.post_id IN ?", postIds)
	}

	// 排序
	if sort != nil {
		switch *sort {
		case enum.PostSortCreateDesc:
			query = query.Order("p.create_time DESC")
		case enum.PostSortCreateAsc:
			query = query.Order("p.create_time ASC")
		case enum.PostSortModifyDesc:
			query = query.Order("p.last_modify_time DESC")
		case enum.PostSortModifyAsc:
			query = query.Order("p.last_modify_time ASC")
		case enum.PostSortVisitDesc:
			query = query.Order("p.visit DESC")
		case enum.PostSortVisitAsc:
			query = query.Order("p.visit ASC")
		case enum.PostSortPinned:
			query = query.Order("p.pinned DESC")
		}
	} else {
		// 默认创建时间降序
		query = query.Order("p.create_time DESC")
	}

	return query, nil
}

// sqlQueryKey 给查询条件加上关键字查询
func (r *postRepo) sqlQueryKey(base *gorm.DB, key string) *gorm.DB {
	return base.Where("p.title LIKE %?% OR p.slug LIKE %?% OR p.excerpt LIKE %?% OR pc.content LIKE %?%", key, key, key, key)
}

// handlePanic 处理 Panic
func handlePanic(tx *gorm.DB) func(success *bool, err *error) {
	return func(success *bool, err *error) {
		if r := recover(); r != nil {
			*err = fmt.Errorf("panic %v", r)
			*success = false
			if tx != nil {
				tx.Rollback()
			}
		}
	}
}
