/*
 * Copyright 2023 FormulaGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package admin

import (
	"context"
	"formulago/pkg/encrypt"
	"github.com/casbin/casbin/v2"
	"sync"

	"formulago/data/ent"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cockroachdb/errors"
)

type InitDatabase struct {
	DB  *ent.Client
	Csb *casbin.Enforcer
	Mu  *sync.Mutex
}

var (
	mutex              = new(sync.Mutex)
	InitDatabaseStatus bool
)

func NewInitDatabase(db *ent.Client, csb *casbin.Enforcer) *InitDatabase {
	return &InitDatabase{
		DB:  db,
		Csb: csb,
		Mu:  mutex,
	}
}

func (I *InitDatabase) InitDatabase(ctx context.Context) error {
	// add lock to avoid duplicate initialization
	I.Mu.Lock()
	defer I.Mu.Unlock()

	// judge if the initialization had been done
	check, err := I.DB.API.Query().Count(ctx)
	if InitDatabaseStatus || check > 0 {
		return errors.New("Database had been initialized")
	}

	// insert init data
	err = I.insertUserData(ctx)
	if err != nil {
		hlog.Error("insert user data failed", err)
		err = errors.Wrap(err, "insert user data failed")
		return err
	}

	err = I.insertRoleData(ctx)
	if err != nil {
		hlog.Error("insert role data failed", err)
		err = errors.Wrap(err, "insert role data failed")
		return err
	}

	err = I.insertMenuData(ctx)
	if err != nil {
		hlog.Error("insert menu data failed", err)
		err = errors.Wrap(err, "insert menu data failed")
		return err
	}

	err = I.insertApiData(ctx)
	if err != nil {
		hlog.Error("insert api data failed", err)
		err = errors.Wrap(err, "insert api data failed")
		return err
	}
	err = I.insertRoleMenuAuthorityData(ctx)
	if err != nil {
		hlog.Error("insert role menu authority data failed", err)
		err = errors.Wrap(err, "insert role menu authority data failed")
		return err
	}
	err = I.insertCasbinPoliciesData(ctx)
	if err != nil {
		hlog.Error("insert casbin policies data failed", err)
		err = errors.Wrap(err, "insert casbin policies data failed")
		return err
	}

	err = I.insertProviderData(ctx)
	if err != nil {
		hlog.Error("insert provider data failed", err)
		err = errors.Wrap(err, "insert provider data failed")
		return err
	}

	// set init status
	InitDatabaseStatus = true
	return nil
}

// insert init user data
func (I *InitDatabase) insertUserData(ctx context.Context) error {
	var users []*ent.UserCreate
	password, _ := encrypt.BcryptEncrypt("admin123")
	users = append(users, I.DB.User.Create().
		SetUsername("admin").
		SetNickname("admin").
		SetPassword(password).
		SetEmail("admin@gmail.com").
		SetMobile("12345678901").
		SetRoleID(1).
		SetWecom("admin"),
	)

	err := I.DB.User.CreateBulk(users...).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}
	return nil
}

// insert init apis data
func (I *InitDatabase) insertRoleData(ctx context.Context) error {
	var roles []*ent.RoleCreate
	roles = make([]*ent.RoleCreate, 3)
	roles[0] = I.DB.Role.Create().
		SetName("role.admin").
		SetValue("admin").
		SetRemark("超级管理员").
		SetOrderNo(1)

	roles[1] = I.DB.Role.Create().
		SetName("role.stuff").
		SetValue("stuff").
		SetRemark("普通员工").
		SetOrderNo(2)

	roles[2] = I.DB.Role.Create().
		SetName("role.member").
		SetValue("member").
		SetRemark("注册会员").
		SetOrderNo(3)

	err := I.DB.Role.CreateBulk(roles...).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}
	return nil
}

// insert init API data
func (I *InitDatabase) insertApiData(ctx context.Context) error {
	var apis []*ent.APICreate
	apis = make([]*ent.APICreate, 57)
	// USER
	apis[0] = I.DB.API.Create().
		SetPath("/api/admin/user/login").
		SetDescription("apiDesc.userLogin").
		SetAPIGroup("user").
		SetMethod("POST")

	apis[1] = I.DB.API.Create().
		SetPath("/api/admin/user/register").
		SetDescription("apiDesc.userRegister").
		SetAPIGroup("user").
		SetMethod("POST")

	apis[2] = I.DB.API.Create().
		SetPath("/api/admin/user/create").
		SetDescription("apiDesc.createUser").
		SetAPIGroup("user").
		SetMethod("POST")

	apis[3] = I.DB.API.Create().
		SetPath("/api/admin/user/update").
		SetDescription("apiDesc.updateUser").
		SetAPIGroup("user").
		SetMethod("POST")

	apis[4] = I.DB.API.Create().
		SetPath("/api/admin/user/change-password").
		SetDescription("apiDesc.userChangePassword").
		SetAPIGroup("user").
		SetMethod("POST")

	apis[5] = I.DB.API.Create().
		SetPath("/api/admin/user/info").
		SetDescription("apiDesc.OauthUserInfo").
		SetAPIGroup("user").
		SetMethod("GET")

	apis[6] = I.DB.API.Create().
		SetPath("/api/admin/user/list").
		SetDescription("apiDesc.userList").
		SetAPIGroup("user").
		SetMethod("GET")

	apis[7] = I.DB.API.Create().
		SetPath("/api/admin/user").
		SetDescription("apiDesc.deleteUser").
		SetAPIGroup("user").
		SetMethod("DELETE")

	apis[8] = I.DB.API.Create().
		SetPath("/api/admin/user/perm").
		SetDescription("apiDesc.userPermissions").
		SetAPIGroup("user").
		SetMethod("GET")

	apis[9] = I.DB.API.Create().
		SetPath("/api/admin/user/profile").
		SetDescription("apiDesc.userProfile").
		SetAPIGroup("user").
		SetMethod("GET")

	apis[10] = I.DB.API.Create().
		SetPath("/api/admin/user/profile").
		SetDescription("apiDesc.updateProfile").
		SetAPIGroup("user").
		SetMethod("POST")

	apis[11] = I.DB.API.Create().
		SetPath("/api/admin/user/logout").
		SetDescription("apiDesc.logout").
		SetAPIGroup("user").
		SetMethod("GET")

	apis[12] = I.DB.API.Create().
		SetPath("/api/admin/user/status").
		SetDescription("apiDesc.updateUserStatus").
		SetAPIGroup("user").
		SetMethod("POST")

	// ROLE
	apis[13] = I.DB.API.Create().
		SetPath("/api/admin/role/create").
		SetDescription("apiDesc.createRole").
		SetAPIGroup("role").
		SetMethod("POST")

	apis[14] = I.DB.API.Create().
		SetPath("/api/admin/role/update").
		SetDescription("apiDesc.updateRole").
		SetAPIGroup("role").
		SetMethod("POST")

	apis[15] = I.DB.API.Create().
		SetPath("/api/admin/role").
		SetDescription("apiDesc.deleteRole").
		SetAPIGroup("role").
		SetMethod("DELETE")

	apis[16] = I.DB.API.Create().
		SetPath("/api/admin/role/list").
		SetDescription("apiDesc.roleList").
		SetAPIGroup("role").
		SetMethod("GET")

	apis[17] = I.DB.API.Create().
		SetPath("/api/admin/role/status").
		SetDescription("apiDesc.setRoleStatus").
		SetAPIGroup("role").
		SetMethod("POST")

	// MENU
	apis[18] = I.DB.API.Create().
		SetPath("/api/admin/menu/create").
		SetDescription("apiDesc.createMenu").
		SetAPIGroup("menu").
		SetMethod("POST")

	apis[19] = I.DB.API.Create().
		SetPath("/api/admin/menu/update").
		SetDescription("apiDesc.updateMenu").
		SetAPIGroup("menu").
		SetMethod("POST")

	apis[20] = I.DB.API.Create().
		SetPath("/api/admin/menu").
		SetDescription("apiDesc.deleteMenu").
		SetAPIGroup("menu").
		SetMethod("DELETE")

	apis[21] = I.DB.API.Create().
		SetPath("/api/admin/menu/list").
		SetDescription("apiDesc.menuList").
		SetAPIGroup("menu").
		SetMethod("GET")

	apis[22] = I.DB.API.Create().
		SetPath("/api/admin/menu/role").
		SetDescription("apiDesc.roleMenu").
		SetAPIGroup("menu").
		SetMethod("GET")

	apis[23] = I.DB.API.Create().
		SetPath("/api/admin/menu/param/create").
		SetDescription("apiDesc.createMenuParam").
		SetAPIGroup("menu").
		SetMethod("POST")

	apis[24] = I.DB.API.Create().
		SetPath("/api/admin/menu/param/update").
		SetDescription("apiDesc.updateMenuParam").
		SetAPIGroup("menu").
		SetMethod("POST")

	apis[25] = I.DB.API.Create().
		SetPath("/api/admin/menu/param/list").
		SetDescription("apiDesc.menuParamListByMenuID").
		SetAPIGroup("menu").
		SetMethod("GET")

	apis[26] = I.DB.API.Create().
		SetPath("/api/admin/menu/param").
		SetDescription("apiDesc.deleteMenuParam").
		SetAPIGroup("menu").
		SetMethod("DELETE")

	// CAPTCHA
	apis[27] = I.DB.API.Create().
		SetPath("/api/admin/captcha").
		SetDescription("apiDesc.captcha").
		SetAPIGroup("captcha").
		SetMethod("GET")

	// AUTHORIZATION
	apis[28] = I.DB.API.Create().
		SetPath("/api/admin/authority/api/create").
		SetDescription("apiDesc.createApiAuthority").
		SetAPIGroup("authority").
		SetMethod("POST")

	apis[29] = I.DB.API.Create().
		SetPath("/api/admin/authority/api/update").
		SetDescription("apiDesc.updateApiAuthority").
		SetAPIGroup("authority").
		SetMethod("POST")

	apis[30] = I.DB.API.Create().
		SetPath("/api/admin/authority/api/role").
		SetDescription("apiDesc.APIAuthorityOfRole").
		SetAPIGroup("authority").
		SetMethod("POST")

	apis[31] = I.DB.API.Create().
		SetPath("/api/admin/authority/menu/create").
		SetDescription("apiDesc.createMenuAuthority").
		SetAPIGroup("authority").
		SetMethod("POST")

	apis[32] = I.DB.API.Create().
		SetPath("/api/admin/authority/menu/update").
		SetDescription("apiDesc.updateMenuAuthority").
		SetAPIGroup("authority").
		SetMethod("POST")

	apis[33] = I.DB.API.Create().
		SetPath("/api/admin/authority/menu/role").
		SetDescription("apiDesc.menuAuthorityOfRole").
		SetAPIGroup("authority").
		SetMethod("POST")

	// API
	apis[34] = I.DB.API.Create().
		SetPath("/api/admin/api/create").
		SetDescription("apiDesc.createApi").
		SetAPIGroup("api").
		SetMethod("POST")

	apis[35] = I.DB.API.Create().
		SetPath("/api/admin/api/update").
		SetDescription("apiDesc.updateApi").
		SetAPIGroup("api").
		SetMethod("POST")

	apis[36] = I.DB.API.Create().
		SetPath("/api/admin/api").
		SetDescription("apiDesc.deleteAPI").
		SetAPIGroup("api").
		SetMethod("DELETE")

	apis[37] = I.DB.API.Create().
		SetPath("/api/admin/api/list").
		SetDescription("apiDesc.APIList").
		SetAPIGroup("api").
		SetMethod("GET")

	// DICTIONARY
	apis[38] = I.DB.API.Create().
		SetPath("/api/admin/dict/create").
		SetDescription("apiDesc.createDictionary").
		SetAPIGroup("dictionary").
		SetMethod("POST")

	apis[39] = I.DB.API.Create().
		SetPath("/api/admin/dict/update").
		SetDescription("apiDesc.updateDictionary").
		SetAPIGroup("dictionary").
		SetMethod("POST")

	apis[40] = I.DB.API.Create().
		SetPath("/api/admin/dict").
		SetDescription("apiDesc.deleteDictionary").
		SetAPIGroup("dictionary").
		SetMethod("DELETE")

	apis[41] = I.DB.API.Create().
		SetPath("/api/admin/dict/detail").
		SetDescription("apiDesc.deleteDictionaryDetail").
		SetAPIGroup("dictionary").
		SetMethod("DELETE")

	apis[42] = I.DB.API.Create().
		SetPath("/api/admin/dict/detail/create").
		SetDescription("apiDesc.createDictionaryDetail").
		SetAPIGroup("dictionary").
		SetMethod("POST")

	apis[43] = I.DB.API.Create().
		SetPath("/api/admin/dict/detail/update").
		SetDescription("apiDesc.updateDictionaryDetail").
		SetAPIGroup("dictionary").
		SetMethod("POST")

	apis[44] = I.DB.API.Create().
		SetPath("/api/admin/dict/detail/list").
		SetDescription("apiDesc.getDictionaryListDetail").
		SetAPIGroup("dictionary").
		SetMethod("GET")

	apis[45] = I.DB.API.Create().
		SetPath("/api/admin/dict/list").
		SetDescription("apiDesc.getDictionaryList").
		SetAPIGroup("dictionary").
		SetMethod("GET")

	// OAUTH
	apis[46] = I.DB.API.Create().
		SetPath("/api/admin/oauth/provider/create").
		SetDescription("apiDesc.createProvider").
		SetAPIGroup("oauth").
		SetMethod("POST")

	apis[47] = I.DB.API.Create().
		SetPath("/api/admin/oauth/provider/update").
		SetDescription("apiDesc.updateProvider").
		SetAPIGroup("oauth").
		SetMethod("POST")

	apis[48] = I.DB.API.Create().
		SetPath("/api/admin/oauth/provider").
		SetDescription("apiDesc.deleteProvider").
		SetAPIGroup("oauth").
		SetMethod("DELETE")

	apis[49] = I.DB.API.Create().
		SetPath("/api/admin/oauth/provider/list").
		SetDescription("apiDesc.geProviderList").
		SetAPIGroup("oauth").
		SetMethod("GET")

	apis[50] = I.DB.API.Create().
		SetPath("/api/admin/oauth/login").
		SetDescription("apiDesc.oauthLogin").
		SetAPIGroup("oauth").
		SetMethod("POST")

	// TOKEN
	apis[51] = I.DB.API.Create().
		SetPath("/api/admin/token/create").
		SetDescription("apiDesc.createToken").
		SetAPIGroup("token").
		SetMethod("POST")

	apis[52] = I.DB.API.Create().
		SetPath("/api/admin/token/update").
		SetDescription("apiDesc.updateToken").
		SetAPIGroup("token").
		SetMethod("POST")

	apis[53] = I.DB.API.Create().
		SetPath("/api/admin/token").
		SetDescription("apiDesc.deleteToken").
		SetAPIGroup("token").
		SetMethod("DELETE")

	apis[54] = I.DB.API.Create().
		SetPath("/api/admin/token/list").
		SetDescription("apiDesc.getTokenList").
		SetAPIGroup("token").
		SetMethod("GET")

	apis[55] = I.DB.API.Create().
		SetPath("/api/admin/token/status").
		SetDescription("apiDesc.setTokenStatus").
		SetAPIGroup("token").
		SetMethod("POST")

	apis[56] = I.DB.API.Create().
		SetPath("/api/admin/token/logout").
		SetDescription("user.forceLoggingOut").
		SetAPIGroup("token").
		SetMethod("POST")

	err := I.DB.API.CreateBulk(apis...).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}
	return nil
}

// init menu data
func (I *InitDatabase) insertMenuData(ctx context.Context) error {
	var menus []*ent.MenuCreate
	menus = make([]*ent.MenuCreate, 19)
	menus[0] = I.DB.Menu.Create().
		SetMenuLevel(0).
		SetMenuType(0).
		SetParentID(1).
		SetPath("").
		SetName("root").
		SetComponent("").
		SetOrderNo(0).
		SetTitle("").
		SetIcon("").
		SetHideMenu(false)

	menus[1] = I.DB.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(1).
		SetParentID(1).
		SetPath("/api/admin/dashboard").
		SetName("Dashboard").
		SetComponent("/dashboard/workbench/index").
		SetOrderNo(0).
		SetTitle("route.dashboard").
		SetIcon("ant-design:home-outlined").
		SetHideMenu(false)

	menus[2] = I.DB.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(0).
		SetParentID(1).
		SetPath("").
		SetName("System Management").
		SetComponent("LAYOUT").
		SetOrderNo(1).
		SetTitle("route.systemManagementTitle").
		SetIcon("ant-design:tool-outlined").
		SetHideMenu(false)

	menus[3] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/menu").
		SetName("MenuManagement").
		SetComponent("/sys/menu/index").
		SetOrderNo(1).
		SetTitle("route.menuManagementTitle").
		SetIcon("ant-design:bars-outlined").
		SetHideMenu(false)

	menus[4] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/role").
		SetName("Role Management").
		SetComponent("/sys/role/index").
		SetOrderNo(2).
		SetTitle("route.roleManagementTitle").
		SetIcon("ant-design:user-outlined").
		SetHideMenu(false)

	menus[5] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/api").
		SetName("API Management").
		SetComponent("/sys/api/index").
		SetOrderNo(4).
		SetTitle("route.apiManagementTitle").
		SetIcon("ant-design:api-outlined").
		SetHideMenu(false)

	menus[6] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/user").
		SetName("User Management").
		SetComponent("/sys/user/index").
		SetOrderNo(3).
		SetTitle("route.userManagementTitle").
		SetIcon("ant-design:user-outlined").
		SetHideMenu(false)

	menus[7] = I.DB.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(1).
		SetParentID(1).
		SetPath("/api/admin/file").
		SetName("File Management").
		SetComponent("/file/index").
		SetOrderNo(2).
		SetTitle("route.fileManagementTitle").
		SetIcon("ant-design:folder-open-outlined").
		SetHideMenu(true)

	menus[8] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/dictionary").
		SetName("Dictionary Management").
		SetComponent("/sys/dictionary/index").
		SetOrderNo(5).
		SetTitle("route.dictionaryManagementTitle").
		SetIcon("ant-design:book-outlined").
		SetHideMenu(false)

	menus[9] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(2).
		SetParentID(3).
		SetPath("/api/admin/dictionary/detail").
		SetName("Dictionary Detail").
		SetComponent("/sys/dictionary/detail").
		SetOrderNo(1).
		SetTitle("route.dictionaryDetailManagementTitle").
		SetIcon("ant-design:align-left-outlined").
		SetHideMenu(true)

	menus[10] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/oauth").
		SetName("Oauth Management").
		SetComponent("/sys/oauth/index").
		SetOrderNo(6).
		SetTitle("route.oauthManagement").
		SetIcon("ant-design:unlock-filled").
		SetHideMenu(false)

	menus[11] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/api/admin/token").
		SetName("Token Management").
		SetComponent("/sys/token/index").
		SetOrderNo(7).
		SetTitle("route.tokenManagement").
		SetIcon("ant-design:lock-outlined").
		SetHideMenu(false)

	menus[12] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(3).
		SetPath("/sys/logs/index").
		SetName("Logs Management").
		SetComponent("/sys/logs/index").
		SetOrderNo(8).
		SetTitle("日志管理").
		SetIcon("ant-design:profile-twotone").
		SetHideMenu(false)

	menus[13] = I.DB.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(0).
		SetParentID(1).
		SetPath("").
		SetName("Other Pages").
		SetComponent("LAYOUT").
		SetOrderNo(4).
		SetTitle("route.otherPages").
		SetIcon("ant-design:question-circle-outlined").
		SetHideMenu(true)

	menus[14] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(13).
		SetPath("/sys/oauth/callback").
		SetName("OauthCallbackPage").
		SetComponent("/sys/oauth/callback").
		SetOrderNo(2).
		SetTitle("回调页面").
		SetIcon("ant-design:android-filled").
		SetHideMenu(false)

	menus[15] = I.DB.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(1).
		SetParentID(13).
		SetPath("/api/admin/profile").
		SetName("Profile").
		SetComponent("/sys/profile/index").
		SetOrderNo(3).
		SetTitle("route.userProfileTitle").
		SetIcon("ant-design:profile-outlined").
		SetHideMenu(true)

	menus[16] = I.DB.Menu.Create().
		SetMenuLevel(1).
		SetMenuType(0).
		SetParentID(1).
		SetPath("").
		SetName("Dev Tool").
		SetComponent("LAYOUT").
		SetOrderNo(5).
		SetTitle("开发工具").
		SetIcon("ant-design:api-filled").
		SetHideMenu(true)

	menus[17] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(16).
		SetPath("/devtool/structToProto").
		SetName("StructToProto").
		SetComponent("/devtool/structToProto").
		SetOrderNo(1).
		SetTitle("StructToProto").
		SetIcon("ant-design:disconnect-outlined").
		SetHideMenu(false)

	menus[18] = I.DB.Menu.Create().
		SetMenuLevel(2).
		SetMenuType(1).
		SetParentID(16).
		SetPath("/devtool/structTag").
		SetName("DeleteStructTag").
		SetComponent("/devtool/structTag").
		SetOrderNo(2).
		SetTitle("DeleteStructTag").
		SetIcon("ant-design:disconnect-outlined").
		SetHideMenu(false)

	err := I.DB.Menu.CreateBulk(menus...).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}
	return nil
}

// insert admin menu authority

func (I *InitDatabase) insertRoleMenuAuthorityData(ctx context.Context) error {
	count, err := I.DB.Menu.Query().Count(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}

	var menuIDs []uint64
	menuIDs = make([]uint64, count)

	for i := range menuIDs {
		menuIDs[i] = uint64(i + 1)
	}

	err = I.DB.Role.Update().AddMenuIDs(menuIDs...).Exec(ctx)

	if err != nil {
		return errors.Wrap(err, "db failed")
	}
	return nil
}

// init casbin policies

func (I *InitDatabase) insertCasbinPoliciesData(ctx context.Context) error {
	apis, err := I.DB.API.Query().All(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}

	var policies [][]string
	for _, v := range apis {
		policies = append(policies, []string{"1", v.Path, v.Method})
	}

	addResult, err := I.Csb.AddPolicies(policies)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}

	if !addResult {
		return errors.Wrap(err, "db failed")
	}
	return nil
}

func (I *InitDatabase) insertProviderData(ctx context.Context) error {
	var providers []*ent.OauthProviderCreate
	providers = make([]*ent.OauthProviderCreate, 3)

	providers[0] = I.DB.OauthProvider.Create().
		SetName("google").
		SetClientID("your client id").
		SetClientSecret("your client secret").
		SetRedirectURL("http://localhost:3100/oauth/login/callback").
		SetScopes("email openid").
		SetAuthURL("https://accounts.google.com/o/oauth2/auth").
		SetTokenURL("https://oauth2.googleapis.com/token").
		SetAuthStyle(1).
		SetInfoURL("https://www.googleapis.com/oauth2/v2/userinfo?access_token=")

	providers[1] = I.DB.OauthProvider.Create().
		SetName("github").
		SetClientID("your client id").
		SetClientSecret("your client secret").
		SetRedirectURL("http://localhost:3100/oauth/login/callback").
		SetScopes("email openid").
		SetAuthURL("https://github.com/login/oauth/authorize").
		SetTokenURL("https://github.com/login/oauth/access_token").
		SetAuthStyle(2).
		SetInfoURL("https://api.github.com/user")

	providers[2] = I.DB.OauthProvider.Create().
		SetName("wecom").
		SetAppID("your app id").
		SetClientID("your client id").
		SetClientSecret("your client secret").
		SetRedirectURL("http://localhost:3100/oauth/login/callback").
		SetScopes("email openid").
		SetAuthURL("https://open.work.weixin.qq.com/wwopen/sso/qrConnect").
		SetAuthStyle(2).
		SetInfoURL("https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo").
		SetAppID("your app id")

	err := I.DB.OauthProvider.CreateBulk(providers...).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "db failed")
	}
	return nil
}
