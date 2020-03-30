package adapter

import (
	"context"
	"fmt"

	"github.com/LyricTian/gin-admin/internal/app/model"
	"github.com/LyricTian/gin-admin/internal/app/schema"
	"github.com/LyricTian/gin-admin/pkg/logger"
	casbinModel "github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/google/wire"
)

// CasbinAdapterSet 注入CasbinAdapter
var CasbinAdapterSet = wire.NewSet(wire.Struct(new(CasbinAdapter), "*"))

// CasbinAdapter casbin适配器
type CasbinAdapter struct {
	RoleModel         model.IRole
	RoleMenuModel     model.IRoleMenu
	MenuResourceModel model.IMenuActionResource
	UserModel         model.IUser
	UserRoleModel     model.IUserRole
}

// LoadPolicy loads all policy rules from the storage.
func (a *CasbinAdapter) LoadPolicy(model casbinModel.Model) error {
	ctx := context.Background()
	err := a.loadRolePolicy(ctx, model)
	if err != nil {
		logger.Errorf(ctx, "Load casbin role policy error: %s", err.Error())
		return err
	}

	err = a.loadUserPolicy(ctx, model)
	if err != nil {
		logger.Errorf(ctx, "Load casbin user policy error: %s", err.Error())
		return err
	}
	return nil
}

// 加载角色策略(p,role_id,path,method)
func (a *CasbinAdapter) loadRolePolicy(ctx context.Context, m casbinModel.Model) error {
	roleResult, err := a.RoleModel.Query(ctx, schema.RoleQueryParam{
		Status: 1,
	})
	if err != nil {
		return err
	} else if len(roleResult.Data) == 0 {
		return nil
	}

	roleMenuResult, err := a.RoleMenuModel.Query(ctx, schema.RoleMenuQueryParam{})
	if err != nil {
		return err
	}
	mRoleMenus := roleMenuResult.Data.ToRoleIDMap()

	menuResourceResult, err := a.MenuResourceModel.Query(ctx, schema.MenuActionResourceQueryParam{})
	if err != nil {
		return err
	}
	mMenuResources := menuResourceResult.Data.ToActionIDMap()

	mcache := make(map[string]struct{})
	for _, item := range roleResult.Data {
		if rms, ok := mRoleMenus[item.RecordID]; ok {
			for _, actionID := range rms.ToActionIDs() {
				if mrs, ok := mMenuResources[actionID]; ok {
					for _, mr := range mrs {
						if mr.Path == "" || mr.Method == "" {
							continue
						} else if _, ok := mcache[mr.Path+mr.Method]; ok {
							continue
						}
						mcache[mr.Path+mr.Method] = struct{}{}
						line := fmt.Sprintf("p,%s,%s,%s", item.RecordID, mr.Path, mr.Method)
						persist.LoadPolicyLine(line, m)
					}
				}
			}
		}
	}

	return nil
}

// 加载用户策略(g,user_id,role_id)
func (a *CasbinAdapter) loadUserPolicy(ctx context.Context, m casbinModel.Model) error {
	userResult, err := a.UserModel.Query(ctx, schema.UserQueryParam{
		Status: 1,
	})
	if err != nil {
		return err
	} else if len(userResult.Data) == 0 {
		return nil
	}

	userRoleResult, err := a.UserRoleModel.Query(ctx, schema.UserRoleQueryParam{})
	if err != nil {
		return err
	}

	mUserRoles := userRoleResult.Data.ToUserIDMap()
	for _, uitem := range userResult.Data {
		if urs, ok := mUserRoles[uitem.RecordID]; ok {
			for _, ur := range urs {
				line := fmt.Sprintf("g,%s,%s", ur.UserID, ur.RoleID)
				persist.LoadPolicyLine(line, m)
			}
		}
	}

	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *CasbinAdapter) SavePolicy(model casbinModel.Model) error {
	return nil
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (a *CasbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (a *CasbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// This is part of the Auto-Save feature.
func (a *CasbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}
