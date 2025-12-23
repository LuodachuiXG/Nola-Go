package repository

import (
	"context"
	"errors"
	"nola-go/internal/db"
	"nola-go/internal/logger"
	"nola-go/internal/models"
	"nola-go/internal/models/enum"
	"nola-go/internal/models/request"
	"nola-go/internal/util"
	"time"

	"gorm.io/gorm"
)

// MenuRepository 菜单 Repo 接口
type MenuRepository interface {
	// AddMenu 添加菜单
	AddMenu(ctx context.Context, menu *request.MenuRequest) (*models.Menu, error)
	// DeleteMenus 删除菜单
	DeleteMenus(ctx context.Context, menuIds []uint) (bool, error)
	// UpdateMenu 修改菜单
	UpdateMenu(ctx context.Context, menu *request.MenuRequest) (bool, error)
	// MenuById 获取菜单 - 菜单 ID
	MenuById(ctx context.Context, menuId uint) (*models.Menu, error)
	// MenuByDisplayName 获取菜单 - 菜单名
	MenuByDisplayName(ctx context.Context, displayName string) (*models.Menu, error)
	// Menus 获取所有菜单
	Menus(ctx context.Context) ([]*models.Menu, error)
	// MenusPager 分页获取菜单
	MenusPager(ctx context.Context, page, size int) (*models.Pager[models.Menu], error)
	// AddMenuItem 添加菜单项
	AddMenuItem(ctx context.Context, menuItem *request.MenuItemRequest) (*models.MenuItem, error)
	// DeleteMenuItems 删除菜单项
	DeleteMenuItems(ctx context.Context, menuItemIds []uint) (bool, error)
	// UpdateMenuItem 修改菜单项
	UpdateMenuItem(ctx context.Context, menuItem *request.MenuItemRequest) (bool, error)
	// MenuItemById 获取菜单项 - 菜单项 ID
	MenuItemById(ctx context.Context, menuItemId uint) (*models.MenuItem, error)
	// MenuItemsByMenuId 获取所有菜单项 - 菜单 ID
	MenuItemsByMenuId(ctx context.Context, menuId uint) ([]*models.MenuItem, error)
	// MenuItems 获取所有菜单项
	MenuItems(ctx context.Context) ([]*models.MenuItem, error)
	// MainMenuItems 获取主菜单菜单项
	MainMenuItems(ctx context.Context) ([]*models.MenuItem, error)
	// MenuCount 获取菜单数量
	MenuCount(ctx context.Context) (int64, error)
	// MainMenuItemCount 获取主菜单菜单项数量
	MainMenuItemCount(ctx context.Context) (int64, error)
}

type menuRepo struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepo{
		db: db,
	}
}

// AddMenu 添加菜单
func (r *menuRepo) AddMenu(ctx context.Context, menu *request.MenuRequest) (*models.Menu, error) {

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	add := &models.Menu{
		DisplayName: menu.DisplayName,
		IsMain:      menu.IsMain,
		CreateTime:  time.Now().UnixMilli(),
	}

	ret := tx.WithContext(ctx).Create(add)
	if ret.Error != nil {
		tx.Rollback()
		return nil, ret.Error
	}

	if menu.IsMain && ret.RowsAffected > 0 {
		// 当前菜单被设为了主菜单，将其他所有菜单设为非主菜单
		_, err := r.setMainMenu(ctx, add.MenuId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return add, nil
}

// DeleteMenus 删除菜单
func (r *menuRepo) DeleteMenus(ctx context.Context, menuIds []uint) (bool, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除被删除的菜单下的菜单项
	err := tx.Where("parent_menuId IN ?", menuIds).Delete(&models.MenuItem{}).Error
	if err != nil {
		tx.Rollback()
		return false, nil
	}

	// 删除菜单
	deleteRet := tx.Where("menu_id IN ?", menuIds).Delete(&models.Menu{})
	if err := deleteRet.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	ret := tx.Commit()

	if err := ret.Error; err != nil {
		return false, ret.Error
	}

	return deleteRet.RowsAffected > 0, nil
}

// UpdateMenu 修改菜单
func (r *menuRepo) UpdateMenu(ctx context.Context, menu *request.MenuRequest) (bool, error) {

	if menu.MenuId == nil {
		logger.Log.Error("菜单 ID 不能为 nil")
		return false, nil
	}

	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	update := map[string]any{
		"display_name":     menu.DisplayName,
		"is_main":          menu.IsMain,
		"last_modify_time": util.Int64Ptr(time.Now().UnixMilli()),
	}

	updateRet := tx.Model(&models.Menu{}).
		Where("menu_id = ?", *menu.MenuId).
		Updates(update)

	if err := updateRet.Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if menu.IsMain && updateRet.RowsAffected > 0 {
		// 当前菜单被设为了主菜单，将其他所有菜单设为非主菜单
		_, err := r.setMainMenu(ctx, *menu.MenuId)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return updateRet.RowsAffected > 0, nil
}

// MenuById 获取菜单 - 菜单 ID
func (r *menuRepo) MenuById(ctx context.Context, menuId uint) (*models.Menu, error) {
	var menu *models.Menu
	err := r.db.WithContext(ctx).Where("menu_id = ?", menuId).First(&menu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return menu, nil
}

// MenuByDisplayName 获取菜单 - 菜单名
func (r *menuRepo) MenuByDisplayName(ctx context.Context, displayName string) (*models.Menu, error) {
	var menu *models.Menu
	err := r.db.WithContext(ctx).Where("display_name = ?", displayName).First(&menu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return menu, nil
}

// Menus 获取所有菜单
func (r *menuRepo) Menus(ctx context.Context) ([]*models.Menu, error) {
	var menus []*models.Menu
	err := r.db.WithContext(ctx).Model(&models.Menu{}).Find(&menus).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return menus, nil
}

// MenusPager 分页获取菜单
func (r *menuRepo) MenusPager(ctx context.Context, page, size int) (*models.Pager[models.Menu], error) {
	pager, err := db.PagerBuilder[models.Menu](ctx, r.db, page, size, func(query *gorm.DB) *gorm.DB {
		return query.Model(&models.Menu{})
	})

	if err != nil {
		return nil, err
	}
	return pager, nil
}

// AddMenuItem 添加菜单项
func (r *menuRepo) AddMenuItem(ctx context.Context, menuItem *request.MenuItemRequest) (*models.MenuItem, error) {
	// 先留出空位给新添加的菜单项
	err := r.sqlUpdateMenuItemIndex(ctx, menuItem, true)
	if err != nil {
		return nil, err
	}
	add := &models.MenuItem{
		DisplayName:      menuItem.DisplayName,
		Href:             menuItem.Href,
		ParentMenuId:     &menuItem.ParentMenuId,
		ParentMenuItemId: menuItem.ParentMenuItemId,
		Index:            menuItem.Index,
		CreateTime:       time.Now().UnixMilli(),
	}

	if menuItem.Target == nil {
		add.Target = enum.MenuTargetBlank
	}

	err = r.db.WithContext(ctx).Create(add).Error
	if err != nil {
		return nil, err
	}
	return add, nil
}

// DeleteMenuItems 删除菜单项
func (r *menuRepo) DeleteMenuItems(ctx context.Context, menuItemIds []uint) (bool, error) {
	if len(menuItemIds) == 0 {
		return false, nil
	}

	ret := r.db.WithContext(ctx).Where("menu_item_id IN ?", menuItemIds).Delete(&models.MenuItem{})
	if ret.Error != nil {
		return false, ret.Error
	}

	return ret.RowsAffected > 0, nil
}

// UpdateMenuItem 修改菜单项
func (r *menuRepo) UpdateMenuItem(ctx context.Context, menuItem *request.MenuItemRequest) (bool, error) {
	// 先留出空位给新修改的菜单项
	err := r.sqlUpdateMenuItemIndex(ctx, menuItem, false)

	if err != nil {
		return false, err
	}

	updates := map[string]any{
		"display_name":        menuItem.DisplayName,
		"href":                menuItem.Href,
		"target":              util.DefaultPtr(menuItem.Target, enum.MenuTargetBlank),
		"parent_menuId":       menuItem.ParentMenuId,
		"parent_menu_item_id": menuItem.ParentMenuItemId,
		"index":               menuItem.Index,
		"last_modify_time":    time.Now().UnixMilli(),
	}

	ret := r.db.WithContext(ctx).
		Where("menu_item_id = ?", menuItem.MenuItemId).
		Model(&models.MenuItem{}).
		Updates(updates)

	if ret.Error != nil {
		return false, ret.Error
	}

	return ret.RowsAffected > 0, nil
}

// MenuItemById 获取菜单项 - 菜单 ID
func (r *menuRepo) MenuItemById(ctx context.Context, menuItemId uint) (*models.MenuItem, error) {
	var menuItem *models.MenuItem
	err := r.db.WithContext(ctx).
		Where("menu_item_id = ?", menuItemId).
		Model(&models.MenuItem{}).
		First(&menuItem).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return menuItem, nil
}

// MenuItemsByMenuId 获取所有菜单项 - 菜单 ID
func (r *menuRepo) MenuItemsByMenuId(ctx context.Context, menuId uint) ([]*models.MenuItem, error) {
	var menuItems []*models.MenuItem

	err := r.db.WithContext(ctx).
		Model(&models.MenuItem{}).
		Where("parent_menuId = ?", menuId).
		Order("`index` ASC").
		Find(&menuItems).Error

	if err != nil {
		return nil, err
	}
	return menuItems, nil
}

// MenuItems 获取所有菜单项
func (r *menuRepo) MenuItems(ctx context.Context) ([]*models.MenuItem, error) {
	var menuItems []*models.MenuItem
	err := r.db.WithContext(ctx).
		Model(&models.MenuItem{}).
		Order("`index` ASC").
		Find(&menuItems).Error
	if err != nil {
		return nil, err
	}
	return menuItems, nil
}

// MainMenuItems 获取主菜单菜单项
func (r *menuRepo) MainMenuItems(ctx context.Context) ([]*models.MenuItem, error) {
	var menuItems []*models.MenuItem
	err := r.db.WithContext(ctx).
		Table("menu_item mt").
		Joins("LEFT JOIN menu m ON m.menu_id = mt.parent_menuId").
		Where("m.is_main = 1").
		Order("mt.index ASC").
		Find(&menuItems).Error

	if err != nil {
		return nil, err
	}
	return menuItems, nil
}

// MenuCount 获取菜单数量
func (r *menuRepo) MenuCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Menu{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// MainMenuItemCount 获取主菜单菜单项数量
func (r *menuRepo) MainMenuItemCount(ctx context.Context) (int64, error) {
	// 先获取主菜单 ID
	var mainMenu *models.Menu
	err := r.db.WithContext(ctx).
		Where("is_main = 1").
		Model(&models.Menu{}).
		First(&mainMenu).Error

	if err != nil {
		return 0, err
	}

	if mainMenu == nil {
		return 0, nil
	}

	// 获取主菜单的菜单项数量
	var count int64
	err = r.db.WithContext(ctx).
		Where("parent_menuId = ?", mainMenu.MenuId).
		Model(&models.MenuItem{}).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// setMainMenu 将非此菜单 ID 的菜单设置为非主菜单
// - Parameters:
//   - ctx: 上下文
//   - menuId: 菜单 ID（此菜单外的所有菜单都会被设为非主菜单，即 isMain = false）
func (r *menuRepo) setMainMenu(ctx context.Context, menuId uint) (int, error) {
	ret := r.db.WithContext(ctx).
		Model(&models.Menu{}).
		Where("menu_id != ?", menuId).
		Update("is_main", false)

	if ret.Error != nil {
		return 0, ret.Error
	}

	return int(ret.RowsAffected), nil
}

// sqlUpdateMenuItemIndex 重新设置和新添加的菜单同级的菜单的 index
// 需要在添加（或更新）当前给定的菜单前，调用此方法，用于空出当前菜单的 index 位置
// - Parameters:
//   - ctx: 上下文
//   - newMenuItem: 新添加的菜单项
//   - isAddMenu: 是否是添加菜单，修改菜单的话赋 false
func (r *menuRepo) sqlUpdateMenuItemIndex(
	ctx context.Context,
	newMenuItem *request.MenuItemRequest,
	isAddMenu bool,
) error {

	if newMenuItem == nil {
		return nil
	}

	tx := r.db.WithContext(ctx).Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取所有同级菜单
	var sameLevelMenuItems []*models.MenuItem
	err := tx.WithContext(ctx).
		Model(&models.MenuItem{}).
		Where("parent_menuId = ? AND parent_menu_item_id = ?",
			newMenuItem.ParentMenuId, newMenuItem.ParentMenuItemId).
		Order("`index` ASC").
		Find(&sameLevelMenuItems).Error

	if err != nil {
		return err
	}

	if len(sameLevelMenuItems) == 0 {
		return nil
	}

	var newIndex uint = 0

	for _, menuItem := range sameLevelMenuItems {
		if isAddMenu {
			// 当前是添加菜单，只需留出空位即可
			if newIndex == newMenuItem.Index {
				newIndex++
			}
			if newIndex != newMenuItem.Index {
				err := tx.Where("menu_item_id = ?", menuItem.MenuItemId).
					Model(&models.MenuItem{}).
					Update("index", newIndex).Error
				if err != nil {
					tx.Rollback()
					return err
				}
			}
			newIndex++
		} else {
			// 当前是修改菜单，sameLevelMenuItems 中已经包含了新菜单
			if menuItem.MenuItemId != *newMenuItem.MenuItemId {
				// 如果当前位置是新菜单要用的位置，这里留出空位
				if newIndex == newMenuItem.Index {
					newIndex++
				}

				err := tx.Where("menu_item_id = ?", menuItem.MenuItemId).
					Model(&models.MenuItem{}).
					Update("index", newIndex).Error
				if err != nil {
					tx.Rollback()
					return err
				}
				newIndex++
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
