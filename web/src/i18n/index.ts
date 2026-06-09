import { ref, computed } from 'vue'

type Messages = Record<string, string>

const zh: Messages = {
  // Nav
  'nav.home': '首页',
  'nav.snippets': '片段',
  'nav.settings': '设置',
  'nav.online': '在线',

  // Login
  'login.title': '登录 cb',
  'login.register': '注册账号',
  'login.username': '用户名',
  'login.password': '密码',
  'login.placeholder.user': '请输入用户名',
  'login.placeholder.pass': '至少 6 个字符',
  'login.submit': '登录',
  'login.submitting': '请稍候...',
  'login.register_btn': '注册',
  'login.no_account': '没有账号？去注册',
  'login.has_account': '已有账号？去登录',

  // Home
  'home.title': '最近片段',
  'home.view_all': '查看全部 →',
  'home.empty': '还没有片段',
  'home.empty_hint': '使用 cb send 或 cb save 创建第一个片段',

  // Snippets
  'snippets.title': '所有片段',
  'snippets.new': '+ 新建片段',
  'snippets.cancel': '取消',
  'snippets.filter_category': '按分类筛选...',
  'snippets.filter_tag': '按标签筛选...',
  'snippets.filter': '筛选',
  'snippets.loading': '加载中...',
  'snippets.encrypted': '已加密',

  // Editor
  'editor.title': '新建片段',
  'editor.content': '粘贴片段内容...',
  'editor.content_required': '内容不能为空',
  'editor.create_failed': '创建失败',
  'editor.alias': '别名（可选）',
  'editor.desc': '描述（可选）',
  'editor.category': '分类（可选）',
  'editor.language': '语言（如 python, bash）',
  'editor.tags': '标签（逗号分隔）',
  'editor.ttl': '过期时间（如 1h, 30m, 7d）',
  'editor.create': '创建片段',
  'editor.creating': '创建中...',

  // Detail
  'detail.copy': '复制',
  'detail.copied': '已复制！',
  'detail.edit': '编辑',
  'detail.cancel': '取消',
  'detail.save': '保存更改',
  'detail.category': '分类',
  'detail.versions': '版本历史',
  'detail.not_found': '片段不存在',
  'detail.loading': '加载中...',

  // Settings
  'settings.title': '设置',
  'settings.admin': '管理员',
  'settings.tab.account': '账户',
  'settings.tab.system': '系统信息',
  'settings.tab.server': '服务配置',
  'settings.tab.users': '用户管理',
  'settings.tab.webhooks': 'Webhooks',

  // Account
  'account.profile': '个人信息',
  'account.username': '用户名',
  'account.user_id': '用户 ID',
  'account.role': '角色',
  'account.role_admin': '管理员',
  'account.role_user': '普通用户',
  'account.change_pass': '修改密码',
  'account.current_pass': '当前密码',
  'account.new_pass': '新密码（至少 6 位）',
  'account.submit_pass': '确认修改',
  'account.pass_success': '密码修改成功！请重新登录。',
  'account.pass_error': '密码修改失败',
  'account.pass_min': '密码至少需要 6 个字符',
  'account.logout': '退出登录',

  // System
  'system.users': '用户数',
  'system.snippets': '片段数',
  'system.devices': '设备数',
  'system.db_size': '数据库大小',
  'system.uptime': '运行时间',
  'system.started': '启动时间',

  // Server
  'server.status': '服务器状态',
  'server.tls': 'TLS',
  'server.enabled': '已启用',
  'server.disabled': '未启用',
  'server.cors': 'CORS 源',
  'server.save': '保存',
  'server.saved': '已保存',
  'server.tls_cert': 'TLS 证书',
  'server.tls_desc': '上传或生成 TLS 证书。更改立即生效，无需重启服务器。',
  'server.upload': '上传并应用',
  'server.generate': '生成自签证书',
  'server.tls_applied': 'TLS 证书已应用',
  'server.tls_generated': 'TLS 证书已生成并应用',

  // Users
  'users.username': '用户名',
  'users.role': '角色',
  'users.created': '注册时间',
  'users.actions': '操作',
  'users.you': '（你）',
  'users.admin': '管理员',
  'users.user': '用户',
  'users.make_admin': '设为管理员',
  'users.revoke_admin': '撤销管理员',
  'users.confirm': '确认',
  'users.reset_pass': '重置密码',
  'users.delete': '删除',
  'users.confirm_delete': '确认删除',
  'users.pass_reset_ok': '密码重置成功',

  // Webhooks
  'webhooks.title': 'Webhook 管理',
  'webhooks.add': '添加 Webhook',
  'webhooks.name': '名称',
  'webhooks.url': '回调 URL',
  'webhooks.events': '事件',
  'webhooks.template': 'Payload 模板（可选）',
  'webhooks.template_hint': '留空则发送默认 JSON。用 {{.变量名}} 插入动态值，用 {{json .变量名}} 安全转义内容（自动处理换行、引号等）',
  'webhooks.template_vars': '查看所有可用变量',
  'webhooks.template_placeholder': '{"text":"[{{.Event}}] {{.Snippet.Content}}"}',
  'webhooks.example': '填写示例',
  'webhooks.example_dingtalk': '钉钉机器人',
  'webhooks.empty': '还没有 Webhook',
  'webhooks.active': '启用',
  'webhooks.inactive': '禁用',
  'webhooks.toggle': '切换',
  'webhooks.logs': '日志',
  'webhooks.delete': '删除',
  'webhooks.confirm_delete': '确认删除',
  'webhooks.no_logs': '暂无投递日志',
  'webhooks.status_ok': '成功',
  'webhooks.status_fail': '失败',
  'webhooks.var.event': '事件类型',
  'webhooks.var.datetime': '时间',
  'webhooks.var.id': 'ID',
  'webhooks.var.userid': '用户ID',
  'webhooks.var.alias': '别名',
  'webhooks.var.desc': '描述',
  'webhooks.var.content': '内容',
  'webhooks.var.encrypted': '是否加密',
  'webhooks.var.category': '分类',
  'webhooks.var.lang': '语言',
  'webhooks.var.expires': '过期时间',
  'webhooks.var.created': '创建时间',
  'webhooks.var.updated': '更新时间',

  // Common
  'common.loading': '加载中...',
  'common.error': '错误',
}

const en: Messages = {
  // Nav
  'nav.home': 'Home',
  'nav.snippets': 'Snippets',
  'nav.settings': 'Settings',
  'nav.online': 'online',

  // Login
  'login.title': 'Login to cb',
  'login.register': 'Create Account',
  'login.username': 'Username',
  'login.password': 'Password',
  'login.placeholder.user': 'Enter username',
  'login.placeholder.pass': 'At least 6 characters',
  'login.submit': 'Login',
  'login.submitting': 'Please wait...',
  'login.register_btn': 'Register',
  'login.no_account': "Don't have an account? Register",
  'login.has_account': 'Already have an account? Login',

  // Home
  'home.title': 'Recent Snippets',
  'home.view_all': 'View all →',
  'home.empty': 'No snippets yet',
  'home.empty_hint': 'Use cb send or cb save to create your first snippet',

  // Snippets
  'snippets.title': 'All Snippets',
  'snippets.new': '+ New Snippet',
  'snippets.cancel': 'Cancel',
  'snippets.filter_category': 'Filter by category...',
  'snippets.filter_tag': 'Filter by tag...',
  'snippets.filter': 'Filter',
  'snippets.loading': 'Loading...',
  'snippets.encrypted': 'encrypted',

  // Editor
  'editor.title': 'New Snippet',
  'editor.content': 'Paste your snippet content here...',
  'editor.content_required': 'Content cannot be empty',
  'editor.create_failed': 'Failed to create snippet',
  'editor.alias': 'Alias (optional)',
  'editor.desc': 'Description (optional)',
  'editor.category': 'Category (optional)',
  'editor.language': 'Language (e.g. python, bash)',
  'editor.tags': 'Tags (comma-separated)',
  'editor.ttl': 'TTL (e.g. 1h, 30m, 7d)',
  'editor.create': 'Create Snippet',
  'editor.creating': 'Creating...',

  // Detail
  'detail.copy': 'Copy',
  'detail.copied': 'Copied!',
  'detail.edit': 'Edit',
  'detail.cancel': 'Cancel',
  'detail.save': 'Save Changes',
  'detail.category': 'Category',
  'detail.versions': 'Version History',
  'detail.not_found': 'Snippet not found',
  'detail.loading': 'Loading...',

  // Settings
  'settings.title': 'Settings',
  'settings.admin': 'Admin',
  'settings.tab.account': 'Account',
  'settings.tab.system': 'System Info',
  'settings.tab.server': 'Server Config',
  'settings.tab.users': 'Users',
  'settings.tab.webhooks': 'Webhooks',

  // Account
  'account.profile': 'Profile',
  'account.username': 'Username',
  'account.user_id': 'User ID',
  'account.role': 'Role',
  'account.role_admin': 'Administrator',
  'account.role_user': 'User',
  'account.change_pass': 'Change Password',
  'account.current_pass': 'Current password',
  'account.new_pass': 'New password (min 6 chars)',
  'account.submit_pass': 'Change Password',
  'account.pass_success': 'Password changed! Please login again.',
  'account.pass_error': 'Failed to change password',
  'account.pass_min': 'Password must be at least 6 characters',
  'account.logout': 'Logout',

  // System
  'system.users': 'Users',
  'system.snippets': 'Snippets',
  'system.devices': 'Devices',
  'system.db_size': 'Database Size',
  'system.uptime': 'Uptime',
  'system.started': 'Started',

  // Server
  'server.status': 'Server Status',
  'server.tls': 'TLS',
  'server.enabled': 'Enabled',
  'server.disabled': 'Disabled',
  'server.cors': 'CORS Origin',
  'server.save': 'Save',
  'server.saved': 'Saved',
  'server.tls_cert': 'TLS Certificate',
  'server.tls_desc': 'Upload or generate a TLS certificate. Changes apply immediately without restarting the server.',
  'server.upload': 'Upload & Apply',
  'server.generate': 'Generate Self-Signed Certificate',
  'server.tls_applied': 'TLS certificate applied',
  'server.tls_generated': 'TLS certificate generated and applied',

  // Users
  'users.username': 'Username',
  'users.role': 'Role',
  'users.created': 'Created',
  'users.actions': 'Actions',
  'users.you': '(you)',
  'users.admin': 'Admin',
  'users.user': 'User',
  'users.make_admin': 'Make Admin',
  'users.revoke_admin': 'Revoke Admin',
  'users.confirm': 'Confirm',
  'users.reset_pass': 'Reset Password',
  'users.delete': 'Delete',
  'users.confirm_delete': 'Confirm Delete',
  'users.pass_reset_ok': 'Password reset successfully',

  // Common
  'common.loading': 'Loading...',
  'common.error': 'Error',

  // Webhooks
  'webhooks.title': 'Webhook Management',
  'webhooks.add': 'Add Webhook',
  'webhooks.name': 'Name',
  'webhooks.url': 'Callback URL',
  'webhooks.events': 'Events',
  'webhooks.template': 'Payload Template (optional)',
  'webhooks.template_hint': 'Leave empty for default JSON. Use {{.Var}} to insert values, {{json .Var}} to safely escape content (handles newlines, quotes, etc.)',
  'webhooks.template_vars': 'View all available variables',
  'webhooks.template_placeholder': '{"text":"[{{.Event}}] {{.Snippet.Content}}"}',
  'webhooks.example': 'Examples',
  'webhooks.example_dingtalk': 'DingTalk Robot',
  'webhooks.empty': 'No webhooks yet',
  'webhooks.active': 'Active',
  'webhooks.inactive': 'Inactive',
  'webhooks.toggle': 'Toggle',
  'webhooks.logs': 'Logs',
  'webhooks.delete': 'Delete',
  'webhooks.confirm_delete': 'Confirm',
  'webhooks.no_logs': 'No delivery logs',
  'webhooks.status_ok': 'OK',
  'webhooks.status_fail': 'Failed',
  'webhooks.var.event': 'event type',
  'webhooks.var.datetime': 'timestamp',
  'webhooks.var.id': 'ID',
  'webhooks.var.userid': 'user ID',
  'webhooks.var.alias': 'alias',
  'webhooks.var.desc': 'description',
  'webhooks.var.content': 'content',
  'webhooks.var.encrypted': 'encrypted',
  'webhooks.var.category': 'category',
  'webhooks.var.lang': 'language',
  'webhooks.var.expires': 'expiry time',
  'webhooks.var.created': 'created at',
  'webhooks.var.updated': 'updated at',
}

const messages: Record<string, Messages> = { zh, en }

const currentLang = ref(localStorage.getItem('cb_lang') || 'zh')

export function useI18n() {
  function t(key: string): string {
    return messages[currentLang.value]?.[key] || messages['en']?.[key] || key
  }

  function setLang(lang: string) {
    currentLang.value = lang
    localStorage.setItem('cb_lang', lang)
  }

  function toggleLang() {
    setLang(currentLang.value === 'zh' ? 'en' : 'zh')
  }

  const lang = computed(() => currentLang.value)
  const langLabel = computed(() => currentLang.value === 'zh' ? '中文' : 'EN')

  return { t, lang, langLabel, setLang, toggleLang }
}
