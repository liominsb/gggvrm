# GGGVRM — 全栈博客系统

基于 **Go + Gin + GORM + MySQL + Redis + RabbitMQ + WebSocket** 构建后端服务，搭配 **Vue 3 + TypeScript + Element Plus** 构建前端 SPA 的全栈博客平台。

## 技术栈

### 后端

| 组件 | 技术 |
|------|------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) |
| 数据库 | MySQL |
| 缓存 | Redis |
| 消息队列 | RabbitMQ |
| 认证 | JWT (golang-jwt) |
| WebSocket | gorilla/websocket |
| 配置管理 | Viper |

### 前端

| 组件 | 技术 |
|------|------|
| 框架 | [Vue 3](https://vuejs.org/) + TypeScript |
| 构建工具 | [Vite](https://vitejs.dev/) |
| UI 组件库 | [Element Plus](https://element-plus.org/) |
| 状态管理 | [Pinia](https://pinia.vuejs.org/) |
| 路由 | [Vue Router](https://router.vuejs.org/) |
| HTTP 客户端 | Axios |
| Markdown 编辑器 | md-editor-v3 |
| 样式 | SCSS（小清新 / 像素风双主题） |

## 项目结构

```
gggvrm/
├── config/                # 后端配置（数据库、Redis、RabbitMQ）
├── controllers/           # 控制器层（处理 HTTP 请求）
├── global/                # 全局变量、消息队列接口
├── middlewares/           # 中间件（JWT 认证）
├── models/                # 数据模型
├── mq/                    # RabbitMQ 生产者/消费者
├── repository/            # 数据访问层
├── router/                # 路由定义 & 依赖注入
├── service/               # 业务逻辑层
├── utils/                 # 工具函数（JWT 生成/解析）
├── uploads/               # 上传文件存储目录
├── main.go                # 后端入口
├── go.mod / go.sum        # Go 依赖管理
│
├── frontend/              # Vue 3 前端项目
│   ├── src/
│   │   ├── api/           # API 请求封装
│   │   ├── components/    # 公共组件（AppHeader、AppFooter、ArticleCard 等）
│   │   ├── router/        # Vue Router 路由配置
│   │   ├── stores/        # Pinia 状态管理
│   │   ├── styles/        # 全局样式 & 主题（fresh-theme / pixel-theme）
│   │   ├── types/         # TypeScript 类型定义
│   │   ├── views/         # 页面视图（Home、Feed、Editor、Chat、Profile 等）
│   │   ├── App.vue        # 根组件
│   │   └── main.ts        # 前端入口
│   ├── index.html
│   ├── vite.config.ts     # Vite 配置
│   ├── tsconfig.json      # TypeScript 配置
│   └── package.json       # 前端依赖管理
│
└── config.yaml            # 运行时配置文件（需自行创建）
```

---

## 认证说明

除 `/api/auth/*` 路由外，所有 `/api/v1/*` 路由均需要 JWT 认证。

**请求头格式：**

```
Authorization: Bearer <token>
```

JWT 中包含以下 Claims：
- `AccountID` (uint) — 用户 ID
- `Username` (string) — 用户名
- `SessionID` (string) — 会话 ID（用于单设备登录 / 踢人下线校验）

认证中间件会从 Redis 中校验 `SessionID` 是否与当前活跃会话一致，若不一致则返回 `409 Conflict`，提示用户已在其他设备登录。

---

## 数据模型

### User

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键（GORM 自增） |
| Username | string | 用户名（唯一索引） |
| Password | string | 密码（bcrypt 加密，JSON 不返回） |
| Bio | string | 个人简介（最长 500 字符） |
| Image | string | 头像 URL |
| CreatedAt | time | 创建时间 |
| UpdatedAt | time | 更新时间 |

### Article

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Title | string | 文章标题 |
| Content | string | 文章内容 |
| Preview | string | 文章预览摘要 |
| Likes | int | 点赞数（默认 0） |
| Views | int | 浏览数（默认 0） |
| UserID | uint | 作者 ID（外键） |
| User | User | 作者信息 |
| CoverImg | string | 封面图 URL |
| CategoryID | *uint | 分类 ID |
| Category | Category | 分类信息 |
| Tags | []Tag | 标签列表（多对多） |
| Comments | []Comment | 评论列表 |
| CreatedAt | time | 创建时间 |
| UpdatedAt | time | 更新时间 |

### Comment

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| ArticleID | uint | 所属文章 ID |
| UserID | uint | 评论者 ID |
| Content | string | 评论内容 |
| CreatedAt | time | 创建时间 |

### Tag

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Name | string | 标签名称（唯一） |
| CreatedAt | time | 创建时间 |

### Category

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键 |
| Name | string | 分类名称 |
| CreatedAt | time | 创建时间 |

### UserFollow（关注中间表）

| 字段 | 类型 | 说明 |
|------|------|------|
| FollowerID | uint | 关注者 ID（主键） |
| FolloweeID | uint | 被关注者 ID（主键） |
| CreatedAt | time | 关注时间 |

### UserArticleFavor（收藏中间表）

| 字段 | 类型 | 说明 |
|------|------|------|
| UserID | uint | 用户 ID（主键） |
| ArticleID | uint | 文章 ID（主键） |
| CreatedAt | time | 收藏时间 |

---

## API 接口文档

### 1. 认证模块（无需认证）

#### 1.1 注册

```
POST /api/auth/register
```

**请求体：**

```json
{
  "username": "string (必填)",
  "password": "string (必填)"
}
```

**成功响应 `200 OK`：**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**错误响应 `400 / 500`：**

```json
{
  "error": "错误信息"
}
```

---

#### 1.2 登录

```
POST /api/auth/login
```

**请求体：**

```json
{
  "username": "string (必填)",
  "password": "string (必填)"
}
```

**成功响应 `200 OK`：**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

**错误响应 `500`：**

```json
{
  "error": "错误信息"
}
```

---

#### 1.3 刷新令牌

```
POST /api/auth/refreshTokens
```

**请求体：**

```json
{
  "access_token": "string (必填)",
  "refreshtoken": "string (必填)"
}
```

**成功响应 `200 OK`：**

```json
{
  "token": "新的 access token",
  "refreshToken": "新的 refresh token"
}
```

**错误响应 `400 / 500`：**

```json
{
  "error": "错误信息"
}
```

---

### 2. 用户模块（🔒 需认证）

#### 2.1 获取当前登录用户信息

```
GET /api/v1/user
```

**成功响应 `200 OK`：**

```json
{
  "user": {
    "ID": 1,
    "username": "testuser",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

---

#### 2.2 获取指定用户信息

```
GET /api/v1/user/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 目标用户 ID |

**成功响应 `200 OK`：**

```json
{
  "user": {
    "ID": 5,
    "username": "otheruser",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

**错误响应 `404`：**

```json
{
  "error": "找不到该用户"
}
```

---

#### 2.3 修改个人资料

```
PUT /api/v1/user/profile
```

**请求体：**

```json
{
  "username": "string (必填)",
  "image": "string (可选，头像 URL)",
  "bio": "string (可选，个人简介)"
}
```

**成功响应 `200 OK`：**

```json
{
  "user": {
    "ID": 1,
    "username": "newname",
    "bio": "这是我的个人简介",
    "image": "/uploads/images/avatar.jpg",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

**错误响应 `400`：**

```json
{
  "error": "错误信息"
}
```

---

#### 2.4 修改密码

```
PUT /api/v1/user/password
```

**请求体：**

```json
{
  "old_password": "string (必填)",
  "new_password": "string (必填)"
}
```

**成功响应 `200 OK`：**

```json
{
  "message": "密码修改成功"
}
```

**错误响应 `400`：**

```json
{
  "error": "错误信息"
}
```

---

### 3. 文章模块（🔒 需认证）

#### 3.1 创建文章

```
POST /api/v1/articles
```

**请求体：**

```json
{
  "title": "string (必填)",
  "content": "string (必填)",
  "preview": "string (必填)",
  "category_id": 1,
  "tag_ids": [1, 2, 3],
  "cover_img": "string (可选，封面图 URL)"
}
```

**成功响应 `201 Created`：**

```json
{
  "ID": 1,
  "title": "文章标题",
  "content": "文章内容",
  "preview": "文章预览",
  "likes": 0,
  "views": 0,
  "user_id": 1,
  "cover_img": "/uploads/images/xxx.jpg",
  "category_id": 1,
  "category": null,
  "tags": [],
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

---

#### 3.2 删除文章

```
DELETE /api/v1/article/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
{
  "status": "ok",
  "message": "删除成功"
}
```

**错误响应：**
- `404` — 文章不存在
- `403` — 无权限删除该文章

---

#### 3.3 获取文章列表（分页 + 筛选）

```
GET /api/v1/articles
```

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量 |
| category_id | int | — | 按分类筛选 |
| tag_id | int | — | 按标签筛选 |
| keyword | string | — | 搜索关键词 |

**成功响应 `200 OK`：**

```json
{
  "data": [
    {
      "ID": 1,
      "title": "文章标题",
      "content": "...",
      "preview": "预览内容",
      "likes": 10,
      "views": 100,
      "user_id": 1,
      "user": { "ID": 1, "username": "testuser" },
      "cover_img": "/uploads/images/xxx.jpg",
      "category_id": 1,
      "category": { "ID": 1, "name": "技术" },
      "tags": [{ "ID": 1, "name": "Go" }],
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 10,
  "total_pages": 5
}
```

**错误响应 `400`：**

```json
{
  "error": "请求页码超出最大支持范围"
}
```

---

#### 3.4 获取单篇文章详情

```
GET /api/v1/article/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
{
  "article": {
    "ID": 1,
    "title": "文章标题",
    "content": "...",
    "preview": "预览内容",
    "likes": 10,
    "views": 100,
    "user_id": 1,
    "user": { "ID": 1, "username": "testuser" },
    "cover_img": "/uploads/images/xxx.jpg",
    "category": { "ID": 1, "name": "技术" },
    "tags": [{ "ID": 1, "name": "Go" }],
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  },
  "comments": [
    {
      "ID": 1,
      "article_id": 1,
      "user_id": 2,
      "content": "写得不错！",
      "created_at": "2025-01-02T00:00:00Z"
    }
  ],
  "likes": 10
}
```

**错误响应：**
- `404` — 该文章不存在

---

#### 3.5 更新文章

```
PUT /api/v1/article/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**请求体：**

```json
{
  "title": "string (必填)",
  "content": "string (必填)",
  "preview": "string (必填)",
  "category_id": 1,
  "tag_ids": [1, 2, 3],
  "cover_img": "string (可选)"
}
```

**成功响应 `200 OK`：**

```json
{
  "message": "更新成功",
  "article": { ... }
}
```

**错误响应：**
- `404` — 文章不存在
- `403` — 无权修改他人的文章

---

#### 3.6 游标分页获取文章列表

```
GET /api/v1/articles/cursor
```

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| cursor | uint | 0 | 游标（上次返回的最后一条 ID，首页传 0） |
| limit | int | 10 | 每次获取数量 |

**成功响应 `200 OK`：**

```json
{
  "data": [
    {
      "ID": 95,
      "title": "...",
      "content": "...",
      "preview": "...",
      "likes": 5,
      "views": 50,
      "user_id": 1,
      "user": { ... },
      "category": { ... },
      "tags": [ ... ],
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "next_cursor": 85,
  "has_more": true
}
```

---

#### 3.7 获取个人 Feed 流

```
GET /api/v1/articles/feed
```

> 返回当前登录用户关注的人所发布的文章。

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量 |

**成功响应 `200 OK`：**

```json
{
  "data": [ ... ],
  "total": 20,
  "page": 1,
  "page_size": 10,
  "total_pages": 2
}
```

---

### 4. 点赞模块（🔒 需认证）

#### 4.1 点赞文章

```
POST /api/v1/article/:id/like
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
{
  "message": "点赞成功: 11"
}
```

**错误响应 `404`：**

```json
{
  "error": "文章不存在"
}
```

---

#### 4.2 获取文章点赞数

```
GET /api/v1/article/:id/like
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
{
  "likes": 11
}
```

---

### 5. 评论模块（🔒 需认证）

#### 5.1 创建评论

```
POST /api/v1/article/:id/comment
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**请求体：**

```json
{
  "content": "string (必填)"
}
```

**成功响应 `201 Created`：**

```json
{
  "ID": 1,
  "article_id": 1,
  "user_id": 2,
  "content": "写得不错！",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

---

#### 5.2 删除评论

```
DELETE /api/v1/comment/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 评论 ID |

**成功响应 `200 OK`：**

```json
{
  "message": "删除成功"
}
```

---

#### 5.3 获取文章的所有评论

```
GET /api/v1/article/:id/comments
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
[
  {
    "ID": 1,
    "article_id": 1,
    "user_id": 2,
    "content": "写得不错！",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
]
```

---

### 6. 文件上传（🔒 需认证）

#### 6.1 批量上传图片

```
POST /api/v1/upload
Content-Type: multipart/form-data
```

**表单字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| files | File[] | 图片文件数组（仅接受图片类型） |

**成功响应 `200 OK`：**

```json
{
  "message": "批量上传成功",
  "urls": [
    "/uploads/images/1735689600000000000.jpg",
    "/uploads/images/1735689600000000001.png"
  ]
}
```

> 上传后的文件可通过 `GET /uploads/images/<filename>` 静态访问。

---

### 7. 标签模块（🔒 需认证）

#### 7.1 获取所有标签

```
GET /api/v1/tags
```

**成功响应 `200 OK`：**

```json
[
  {
    "ID": 1,
    "name": "Go",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
]
```

---

#### 7.2 创建标签

```
POST /api/v1/tag
```

**请求体：**

```json
{
  "name": "string (必填)"
}
```

**成功响应 `200 OK`：**

```json
{
  "ID": 2,
  "name": "Rust",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

---

#### 7.3 删除标签

```
DELETE /api/v1/tag/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 标签 ID |

**成功响应 `200 OK`：** 返回 `null`

---

### 8. 分类模块（🔒 需认证）

#### 8.1 获取所有分类

```
GET /api/v1/categories
```

**成功响应 `200 OK`：**

```json
[
  {
    "ID": 1,
    "name": "技术",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z"
  }
]
```

---

#### 8.2 创建分类

```
POST /api/v1/category
```

**请求体：**

```json
{
  "name": "string (必填)"
}
```

**成功响应 `200 OK`：**

```json
{
  "ID": 2,
  "name": "生活",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

---

#### 8.3 删除分类

```
DELETE /api/v1/category/:id
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 分类 ID |

**成功响应 `200 OK`：** 返回 `null`

---

### 9. 收藏模块（🔒 需认证）

#### 9.1 收藏 / 取消收藏文章

```
POST /api/v1/article/:id/favorite
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
// 收藏时
{
  "message": "收藏成功",
  "is_favorited": true
}

// 取消收藏时
{
  "message": "取消收藏成功",
  "is_favorited": false
}
```

---

#### 9.2 获取收藏状态

```
GET /api/v1/article/:id/favorite
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
{
  "is_favorited": true
}
```

---

#### 9.3 获取文章收藏数

```
GET /api/v1/article/:id/favorites/count
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 文章 ID |

**成功响应 `200 OK`：**

```json
{
  "favorites": 42
}
```

---

#### 9.4 获取当前用户收藏列表

```
GET /api/v1/user/favorites
```

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量 |

**成功响应 `200 OK`：**

```json
{
  "data": [ ... ],
  "total": 15,
  "page": 1,
  "page_size": 10,
  "total_pages": 2
}
```

---

### 10. 关注模块（🔒 需认证）

#### 10.1 关注 / 取消关注用户

```
POST /api/v1/user/:id/follow
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 目标用户 ID |

**成功响应 `200 OK`：**

```json
// 关注时
{
  "message": "关注成功",
  "is_following": true
}

// 取消关注时
{
  "message": "取消关注成功",
  "is_following": false
}
```

**错误响应：**
- `400` — 无效的用户 ID / 不能关注自己
- `404` — 用户不存在

---

#### 10.2 获取关注状态

```
GET /api/v1/user/:id/follow
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 目标用户 ID |

**成功响应 `200 OK`：**

```json
{
  "is_following": true
}
```

---

#### 10.3 获取关注数和粉丝数

```
GET /api/v1/user/:id/follow/counts
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 目标用户 ID |

**成功响应 `200 OK`：**

```json
{
  "following_count": 120,
  "followers_count": 3500
}
```

---

#### 10.4 获取关注列表

```
GET /api/v1/user/:id/following
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 目标用户 ID |

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量 |

**成功响应 `200 OK`：**

```json
{
  "data": [ ... ],
  "total": 120,
  "page": 1,
  "page_size": 10,
  "total_pages": 12
}
```

---

#### 10.5 获取粉丝列表

```
GET /api/v1/user/:id/followers
```

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | uint | 目标用户 ID |

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| page | int | 1 | 页码 |
| page_size | int | 10 | 每页数量 |

**成功响应 `200 OK`：**

```json
{
  "data": [ ... ],
  "total": 3500,
  "page": 1,
  "page_size": 10,
  "total_pages": 350
}
```

---

### 11. WebSocket 聊天室（🔒 需认证）

#### 11.1 建立 WebSocket 连接

```
GET /api/v1/ws
Upgrade: websocket
Authorization: Bearer <token>
```

> 升级为 WebSocket 连接后，进入全局聊天室。

**发送消息（客户端 → 服务器）：**

```json
{
  "content": "Hello everyone!"
}
```

> 服务器会自动填充 `username` 字段（从 JWT 中获取），无需客户端传递。

**接收消息（服务器 → 客户端）：**

```json
{
  "username": "testuser",
  "content": "Hello everyone!"
}
```

**心跳机制：**
- 服务器每 **30 秒**发送一次 `Ping` 消息
- 客户端需在 **60 秒**内回复 `Pong`，否则连接将被关闭

---

## API 总览

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/auth/register` | 注册 | ❌ |
| POST | `/api/auth/login` | 登录 | ❌ |
| POST | `/api/auth/refreshTokens` | 刷新令牌 | ❌ |
| GET | `/api/v1/user` | 获取当前用户信息 | ✅ |
| GET | `/api/v1/user/:id` | 获取指定用户信息 | ✅ |
| PUT | `/api/v1/user/profile` | 修改个人资料 | ✅ |
| PUT | `/api/v1/user/password` | 修改密码 | ✅ |
| POST | `/api/v1/articles` | 创建文章 | ✅ |
| DELETE | `/api/v1/article/:id` | 删除文章 | ✅ |
| GET | `/api/v1/articles` | 获取文章列表（分页） | ❌ |
| GET | `/api/v1/article/:id` | 获取文章详情 | ✅ |
| PUT | `/api/v1/article/:id` | 更新文章 | ✅ |
| GET | `/api/v1/articles/cursor` | 游标分页获取文章 | ✅ |
| GET | `/api/v1/articles/feed` | 获取个人 Feed 流 | ✅ |
| POST | `/api/v1/article/:id/like` | 点赞文章 | ✅ |
| GET | `/api/v1/article/:id/like` | 获取文章点赞数 | ✅ |
| POST | `/api/v1/article/:id/comment` | 创建评论 | ✅ |
| DELETE | `/api/v1/comment/:id` | 删除评论 | ✅ |
| GET | `/api/v1/article/:id/comments` | 获取文章评论 | ✅ |
| POST | `/api/v1/upload` | 批量上传图片 | ✅ |
| GET | `/api/v1/tags` | 获取所有标签 | ❌ |
| POST | `/api/v1/tag` | 创建标签 | ✅ |
| DELETE | `/api/v1/tag/:id` | 删除标签 | ✅ |
| GET | `/api/v1/categories` | 获取所有分类 | ❌ |
| POST | `/api/v1/category` | 创建分类 | ✅ |
| DELETE | `/api/v1/category/:id` | 删除分类 | ✅ |
| POST | `/api/v1/article/:id/favorite` | 收藏/取消收藏 | ✅ |
| GET | `/api/v1/article/:id/favorite` | 获取收藏状态 | ✅ |
| GET | `/api/v1/article/:id/favorites/count` | 获取收藏数 | ✅ |
| GET | `/api/v1/user/favorites` | 获取用户收藏列表 | ✅ |
| POST | `/api/v1/user/:id/follow` | 关注/取消关注 | ✅ |
| GET | `/api/v1/user/:id/follow` | 获取关注状态 | ✅ |
| GET | `/api/v1/user/:id/follow/counts` | 获取关注/粉丝数 | ✅ |
| GET | `/api/v1/user/:id/following` | 获取关注列表 | ✅ |
| GET | `/api/v1/user/:id/followers` | 获取粉丝列表 | ✅ |
| GET | `/api/v1/ws` | WebSocket 聊天室 | ✅ |
| GET | `/uploads/*` | 静态文件访问 | ❌ |

---

## 启动方式

### 后端

1. 确保已安装并启动 MySQL、Redis、RabbitMQ
2. 在项目根目录创建 `config/config.yaml` 配置文件，参考 [`config/config.go`](config/config.go) 中的结构填写数据库、Redis、RabbitMQ、JWT 等配置
3. 运行后端：

```bash
go run main.go
```

后端默认启动在 `localhost:8080`（具体端口视配置文件而定）。

### 前端

```bash
cd frontend
npm install    # 安装依赖
npm run dev    # 启动开发服务器（默认 http://localhost:5173）
```

生产构建：

```bash
npm run build  # 输出到 frontend/dist/
```

> 开发模式下前端通过 Vite 代理将 `/api` 请求转发到后端，详见 [`frontend/vite.config.ts`](frontend/vite.config.ts)。