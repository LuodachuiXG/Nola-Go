package service

import (
	"context"
	"errors"
	"fmt"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/request"
	"nola-go/internal/models/response"
	"nola-go/internal/repository"
	"nola-go/internal/util"

	"go.uber.org/zap"
)

type MenuService struct {
	menuRepo repository.MenuRepository
}

// NewMenuService 创建菜单 Service
func NewMenuService(menuRepo repository.MenuRepository) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
	}
}

// AddMenu 添加菜单
func (s *MenuService) AddMenu(c context.Context, menu *request.MenuRequest) (*models.Menu, error) {

	if menu == nil {
		logger.Log.Error("添加菜单失败，菜单为 nil")
		return nil, nil
	}

	// 先检查菜单名是否已经存在
	m, err := s.MenuByDisplayName(c, menu.DisplayName)
	if err != nil {
		return nil, err
	}

	if m != nil {
		// 菜单名已经存在
		return nil, errors.New("菜单名 [" + menu.DisplayName + "] 已存在")
	}

	// 判断昂钱菜单是否是第一个菜单
	count, err := s.MenuCount(c)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		// 将第一个菜单默认设置成主菜单
		menu.IsMain = true
	}

	ret, err := s.menuRepo.AddMenu(c, menu)

	if err != nil {
		logger.Log.Error("添加菜单失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(c context.Context, menuIds []uint) (bool, error) {

	if len(menuIds) == 0 {
		return false, nil
	}

	ret, err := s.menuRepo.DeleteMenus(c, menuIds)
	if err != nil {
		logger.Log.Error("删除菜单失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdateMenu 修改菜单
func (s *MenuService) UpdateMenu(c context.Context, menu *request.MenuRequest) (bool, error) {
	if menu == nil || menu.MenuId == nil {
		logger.Log.Error("修改菜单失败，菜单为 nil")
		return false, nil
	}

	// 判断菜单名是否重复
	m, err := s.MenuByDisplayName(c, menu.DisplayName)
	if err != nil {
		return false, err
	}

	if m != nil && m.MenuId != *menu.MenuId {
		return false, errors.New("菜单名 [" + menu.DisplayName + "] 已存在")
	}

	ret, err := s.menuRepo.UpdateMenu(c, menu)

	if err != nil {
		logger.Log.Error("修改菜单失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// MenuCount 获取菜单数量
func (s *MenuService) MenuCount(c context.Context) (int64, error) {
	count, err := s.menuRepo.MenuCount(c)
	if err != nil {
		logger.Log.Error("获取菜单数量失败", zap.Error(err))
		return 0, response.ServerError
	}
	return count, nil
}

// MenuById 获取菜单 - 菜单 ID
func (s *MenuService) MenuById(c context.Context, menuId uint) (*models.Menu, error) {
	ret, err := s.menuRepo.MenuById(c, menuId)
	if err != nil {
		logger.Log.Error("获取菜单失败", zap.Error(err))
		return nil, response.ServerError
	}

	return ret, nil
}

// MenuByDisplayName 获取菜单 - 菜单名
func (s *MenuService) MenuByDisplayName(c context.Context, displayName string) (*models.Menu, error) {
	ret, err := s.menuRepo.MenuByDisplayName(c, displayName)
	if err != nil {
		logger.Log.Error("获取菜单失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// Menus 获取所有菜单
func (s *MenuService) Menus(c context.Context) ([]*models.Menu, error) {
	ret, err := s.menuRepo.Menus(c)
	if err != nil {
		logger.Log.Error("获取所有菜单失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// MenusPager 分页获取所有菜单
func (s *MenuService) MenusPager(c context.Context, page, size int) (*models.Pager[models.Menu], error) {
	if page == 0 {
		// 获取所有菜单
		all, err := s.Menus(c)
		if err != nil {
			return nil, response.ServerError
		}
		return &models.Pager[models.Menu]{
			Page:       0,
			Size:       0,
			Data:       all,
			TotalData:  int64(len(all)),
			TotalPages: 1,
		}, nil
	}

	// 分页获取菜单
	pager, err := s.menuRepo.MenusPager(c, page, size)
	if err != nil {
		logger.Log.Error("分页获取菜单失败", zap.Error(err))
		return nil, response.ServerError
	}
	return pager, nil
}

// AddMenuItem 添加菜单项
func (s *MenuService) AddMenuItem(c context.Context, menuItem *request.MenuItemRequest) (*models.MenuItem, error) {
	// 添加前先检查父菜单是否存在
	m, err := s.MenuById(c, menuItem.ParentMenuId)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, errors.New(fmt.Sprintf("父菜单 [%d] 不存在", menuItem.ParentMenuId))
	}

	// 父菜单存在，再检查父菜单项是否存在
	if menuItem.ParentMenuItemId != nil {
		mt, err := s.MenuItem(c, *menuItem.ParentMenuItemId)
		if err != nil {
			return nil, err
		}
		if mt == nil {
			return nil, errors.New(fmt.Sprintf("父菜单项 [%d] 不存在", *menuItem.ParentMenuItemId))
		}

		// 检查和父菜单项是否有相同祖先
		if *mt.ParentMenuId != menuItem.ParentMenuId {
			return nil, errors.New("父菜单项必须有相同的父菜单")
		}
	}

	ret, err := s.menuRepo.AddMenuItem(c, menuItem)
	if err != nil {
		logger.Log.Error("添加菜单项失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// DeleteMenuItems 删除菜单项
// - Parameters:
//   - c: 上下文
//   - menuItemIds: 菜单项 ID 列表
func (s *MenuService) DeleteMenuItems(c context.Context, menuItemIds []uint) (bool, error) {
	// 先获取所有菜单项
	menuItems, err := s.menuRepo.MenuItems(c)
	if err != nil {
		logger.Log.Error("获取所有菜单项失败", zap.Error(err))
		return false, response.ServerError
	}

	// 获取要删除的所有菜单项的所有子菜单项 ID
	var childrenIds []uint
	for _, id := range menuItemIds {
		ids := s.findChildrenIds(c, id, menuItems)
		childrenIds = append(childrenIds, ids...)
	}

	// 将要删除的菜单项 ID 列表也加入子菜单项 ID 列表中，一起删除
	childrenIds = append(childrenIds, menuItemIds...)

	// 删除菜单项
	ret, err := s.menuRepo.DeleteMenuItems(c, childrenIds)
	if err != nil {
		logger.Log.Error("删除菜单项失败", zap.Error(err))
		return false, response.ServerError
	}
	return ret, nil
}

// UpdateMenuItem 修改菜单项
func (s *MenuService) UpdateMenuItem(c context.Context, menuItem *request.MenuItemRequest) (bool, error) {
	if menuItem == nil {
		logger.Log.Error("修改菜单项失败，菜单项为 nil")
		return false, nil
	}

	// 判断是否将自己设为父菜单项
	if menuItem.ParentMenuItemId == menuItem.MenuItemId {
		return false, errors.New("不能将自己设为父菜单项")
	}

	// 修改前先检查父菜单是否存在
	pm, err := s.MenuById(c, menuItem.ParentMenuId)
	if err != nil {
		return false, err
	}
	if pm == nil {
		return false, errors.New(fmt.Sprintf("父菜单 [%d] 不存在", menuItem.ParentMenuId))
	}

	// 父菜单项不为空
	if menuItem.ParentMenuItemId != nil {
		// 父菜单项是否存在
		pmt, err := s.MenuItem(c, *menuItem.ParentMenuItemId)
		if err != nil {
			return false, err
		}

		if pmt == nil {
			return false, errors.New(fmt.Sprintf("父菜单项 [%d] 不存在", menuItem.ParentMenuItemId))
		}

		// 判断父菜单项和当前修改的菜单项是否在同一个父菜单下
		if *pmt.ParentMenuId != menuItem.ParentMenuId {
			return false, errors.New("父菜单项和和当前菜单项不在相同父菜单")
		}

		// 如果父菜单项为当前修改菜单的子菜单项（循环设置父菜单，即父菜单项的子菜单项又被设为父菜单项的父菜单项）
		if pmt.ParentMenuItemId == menuItem.MenuItemId {
			// 删除原先的子菜单项的父菜单项
			_, err := s.UpdateMenuItem(c, &request.MenuItemRequest{
				MenuItemId:   &pmt.MenuItemId,
				DisplayName:  pmt.DisplayName,
				Href:         pmt.Href,
				Target:       &pmt.Target,
				ParentMenuId: *pmt.ParentMenuId,
				// 将原先的子菜单的（当前要修改的菜单项的父菜单项）的父菜单项设置为 nil
				ParentMenuItemId: nil,
				Index:            pmt.Index,
			})

			if err != nil {
				return false, err
			}
		}
	}

	ret, err := s.menuRepo.UpdateMenuItem(c, menuItem)
	if err != nil {
		logger.Log.Error("修改菜单项失败", zap.Error(err))
		return false, response.ServerError
	}

	return ret, nil
}

// MenuItem 获取菜单项 - 菜单项 ID
func (s *MenuService) MenuItem(c context.Context, menuItemId uint) (*models.MenuItem, error) {
	ret, err := s.menuRepo.MenuItemById(c, menuItemId)
	if err != nil {
		logger.Log.Error("获取菜单项失败", zap.Error(err))
		return nil, response.ServerError
	}
	return ret, nil
}

// MenuItems 获取菜单项
// - Parameters:
//   - c: 上下文
//   - menuId: 菜单 ID
//   - buildTree: 是否构建菜单树
func (s *MenuService) MenuItems(c context.Context, menuId uint, buildTree bool) ([]*response.MenuItemResponse, error) {
	// 先获取当前菜单的所有菜单项
	menuItems, err := s.menuRepo.MenuItemsByMenuId(c, menuId)
	if err != nil {
		logger.Log.Error("获取菜单项失败", zap.Error(err))
		return nil, response.ServerError
	}

	// 是否构建菜单项树
	if buildTree {
		// 构建菜单项树
		return s.findMenuItemChildren(nil, menuItems), nil
	}

	// 不构建菜单项树
	return util.Map(menuItems, func(mt *models.MenuItem) *response.MenuItemResponse {
		return &response.MenuItemResponse{
			MenuItemId:       mt.MenuItemId,
			DisplayName:      mt.DisplayName,
			Href:             mt.Href,
			Target:           mt.Target,
			ParentMenuId:     *mt.ParentMenuId,
			ParentMenuItemId: mt.ParentMenuItemId,
			Children:         []*response.MenuItemResponse{},
			Index:            mt.Index,
			LastModifyTime:   mt.LastModifyTime,
			CreateTime:       mt.CreateTime,
		}
	}), nil

}

// MainMenu 获取主菜单的菜单项
// - Parameters:
//   - c: 上下文
//   - buildTree: 是否构建菜单树
func (s *MenuService) MainMenu(c context.Context, buildTree bool) ([]*response.MenuItemResponse, error) {
	// 获取主菜单的菜单项
	menuItems, err := s.menuRepo.MainMenuItems(c)
	if err != nil {
		logger.Log.Error("获取主菜单的菜单项失败", zap.Error(err))
		return nil, response.ServerError
	}
	if buildTree {
		// 构建菜单项树
		return s.findMenuItemChildren(nil, menuItems), nil
	}

	// 不构建菜单项树
	return util.Map(menuItems, func(mt *models.MenuItem) *response.MenuItemResponse {
		return &response.MenuItemResponse{
			MenuItemId:       mt.MenuItemId,
			DisplayName:      mt.DisplayName,
			Href:             mt.Href,
			Target:           mt.Target,
			ParentMenuId:     *mt.ParentMenuId,
			ParentMenuItemId: mt.ParentMenuItemId,
			Index:            mt.Index,
			LastModifyTime:   mt.LastModifyTime,
			CreateTime:       mt.CreateTime,
			Children:         []*response.MenuItemResponse{},
		}
	}), nil
}

// MainMenuItemCount 获取主菜单的菜单项数量
func (s *MenuService) MainMenuItemCount(c context.Context) (int64, error) {
	count, err := s.MainMenuItemCount(c)
	if err != nil {
		logger.Log.Error("获取主菜单的菜单项数量失败", zap.Error(err))
		return 0, response.ServerError
	}
	return count, nil
}

// findChildrenIds 查找子菜单项 ID
// - Parameters:
//   - c: 上下文
//   - parentMenuItemId: 父菜单项 ID
//   - menuItems: 菜单项列表
func (s *MenuService) findChildrenIds(c context.Context, parentMenuItemId uint, menuItems []*models.MenuItem) []uint {
	var childrenIds []uint
	for _, menuItem := range menuItems {
		if *menuItem.ParentMenuItemId == parentMenuItemId {
			childrenIds = append(childrenIds, menuItem.MenuItemId)
			// 再递归查找子菜单的子菜单
			childrenIds = append(childrenIds, s.findChildrenIds(c, menuItem.MenuItemId, menuItems)...)
		}
	}

	return childrenIds
}

// findMenuItemChildren 递归函数，查找并构建子菜单项
// - Parameters:
//   - parentMenuItemId: 父菜单项 ID
//   - menuItems: 菜单项列表
func (s *MenuService) findMenuItemChildren(parentMenuItemId *uint, items []*models.MenuItem) []*response.MenuItemResponse {
	filter := util.Filter(items, func(mt *models.MenuItem) bool {
		return mt.ParentMenuItemId == parentMenuItemId
	})

	return util.Map(filter, func(mt *models.MenuItem) *response.MenuItemResponse {
		return &response.MenuItemResponse{
			MenuItemId:       mt.MenuItemId,
			DisplayName:      mt.DisplayName,
			Href:             mt.Href,
			Target:           mt.Target,
			ParentMenuId:     *mt.ParentMenuId,
			ParentMenuItemId: mt.ParentMenuItemId,
			Index:            mt.Index,
			LastModifyTime:   mt.LastModifyTime,
			CreateTime:       mt.CreateTime,
			// 递归查找子菜单
			Children: s.findMenuItemChildren(&mt.MenuItemId, items),
		}
	})
}
