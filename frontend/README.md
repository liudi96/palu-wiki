# 幻兽帕鲁攻略网站 - 管理后台前端

基于 Next.js + TypeScript + Tailwind CSS 构建的现代化管理后台界面。

## 🚀 快速开始

### 1. 安装依赖

```bash
# 如果网络有问题，使用代理
HTTP_PROXY=http://127.0.0.1:15236 HTTPS_PROXY=http://127.0.0.1:15236 npm install

# 或者使用cnpm
npm install -g cnpm --registry=https://registry.npmmirror.com
cnpm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

前端将在 http://localhost:3001 启动

### 3. 访问管理后台

1. 确保后端服务已启动 (http://localhost:8080)
2. 访问 http://localhost:3001
3. 使用以下账户登录:
   - 管理员: admin / admin123 (需要先在数据库中设置角色为admin)
   - 普通用户: testuser / 123456

## 📁 项目结构

```
frontend/
├── src/
│   ├── app/                # Next.js App Router页面
│   │   ├── admin/         # 管理后台页面
│   │   ├── login/         # 登录页面
│   │   └── layout.tsx     # 根布局
│   ├── components/        # 共享组件
│   │   └── AdminLayout.tsx # 管理后台布局
│   └── lib/              # 工具库
│       └── api.ts        # API客户端
├── package.json
├── next.config.js        # Next.js配置
└── tailwind.config.js    # Tailwind CSS配置
```

## 🔧 功能特性

### ✅ 已实现功能
- 🔐 用户登录/退出
- 📊 管理后台仪表板
- 📈 数据统计展示
- 🎨 响应式设计
- 🔒 权限验证

### 🚧 待开发功能
- 📝 文章管理页面
- 👥 用户管理页面
- ⚙️ 系统设置页面
- 📊 数据统计页面

## 🔗 API 集成

前端通过 axios 与后端 API 通信:
- 基础URL: http://localhost:8080
- 认证方式: JWT Bearer Token
- 自动token管理和刷新

## 🎨 设计系统

- **UI框架**: Tailwind CSS
- **图标**: Heroicons
- **字体**: Inter
- **主色调**: Blue (#3b82f6)

## 📝 开发说明

### 环境变量

创建 `.env.local` 文件:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 代理配置

如果遇到CORS问题，已在 `next.config.js` 中配置了API代理。

### 权限控制

管理后台需要管理员权限，普通用户将被重定向到登录页面。

## 🚀 部署

### 开发环境
```bash
npm run dev
```

### 生产环境
```bash
npm run build
npm start
```

## 🔍 故障排除

1. **依赖安装失败**: 使用代理或cnpm
2. **API请求失败**: 检查后端服务是否启动
3. **登录失败**: 确认用户在数据库中存在且密码正确
4. **权限不足**: 确认用户角色为admin

## 📞 支持

如有问题，请检查:
1. 后端服务是否正常运行
2. 数据库连接是否正常
3. 用户权限是否正确设置