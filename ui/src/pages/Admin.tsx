import React, { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  BarChart3,
  ChevronLeft,
  Check,
  Database,
  Edit2,
  Plus,
  Rss,
  Shield,
  Trash2,
  UserPlus,
  Users,
} from 'lucide-react'
import {
  useAdminSettings,
  useAdminCreateFeed,
  useAdminCreateRecommendedFeed,
  useAdminCreateUser,
  useAdminDeleteFeed,
  useAdminDeleteRecommendedFeed,
  useAdminDeleteUser,
  useAdminFeeds,
  useAdminRecommendedFeeds,
  useAdminStats,
  useAdminUpdateFeed,
  useAdminUpdateRecommendedFeed,
  useAdminUpdateSettings,
  useAdminUpdateUser,
  useAdminUsers,
} from '../api/hooks'
import { showToast } from '../components/Toast'
import type { Feed, RecommendedFeed, User } from '../types'

type MenuKey = 'overview' | 'users' | 'feeds' | 'sources'

function StatCard({ title, value, icon: Icon, color }: { title: string; value: number; icon: React.ElementType; color: string }) {
  return (
    <div className="rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)] p-5">
      <div className="flex items-center gap-4">
        <div className={`rounded-xl p-3 ${color}`}>
          <Icon className="h-5 w-5 text-white" />
        </div>
        <div>
          <p className="text-sm text-[var(--text-secondary)]">{title}</p>
          <p className="text-2xl font-semibold text-[var(--text-primary)]">{value}</p>
        </div>
      </div>
    </div>
  )
}

function UserModal({ user, onClose }: { user?: User; onClose: () => void }) {
  const createUser = useAdminCreateUser()
  const updateUser = useAdminUpdateUser()
  const [username, setUsername] = useState(user?.username ?? '')
  const [email, setEmail] = useState(user?.email ?? '')
  const [password, setPassword] = useState('')
  const [isAdmin, setIsAdmin] = useState(Boolean(user?.is_admin))
  const [status, setStatus] = useState<number>(user?.status ?? 1)

  const isEdit = Boolean(user)
  const pending = createUser.isPending || updateUser.isPending

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (isEdit && user) {
        await updateUser.mutateAsync({
          id: user.id,
          email: email || undefined,
          password: password || undefined,
          is_admin: isAdmin,
          status,
        })
        showToast('success', `用户 ${user.username} 已更新`)
        onClose()
        return
      }
      await createUser.mutateAsync({
        username: username.trim(),
        email: email.trim(),
        password,
        is_admin: isAdmin,
      })
      showToast('success', `用户 ${username.trim()} 创建成功`)
      onClose()
    } catch (err) {
      showToast('error', err instanceof Error ? err.message : '保存用户失败')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <form onSubmit={submit} className="w-full max-w-lg rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)] p-6">
        <h3 className="mb-4 text-lg font-semibold text-[var(--text-primary)]">{isEdit ? '编辑用户' : '新增用户'}</h3>
        {!isEdit && (
          <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} required />
        )}
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" placeholder="邮箱" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" type="password" placeholder={isEdit ? '新密码（可空）' : '密码'} value={password} onChange={(e) => setPassword(e.target.value)} required={!isEdit} />
        {isEdit && (
          <select className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={status} onChange={(e) => setStatus(Number(e.target.value))}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
        )}
        <label className="mb-5 flex items-center gap-2 text-sm text-[var(--text-secondary)]">
          <input type="checkbox" checked={isAdmin} onChange={(e) => setIsAdmin(e.target.checked)} />
          管理员权限
        </label>
        <div className="flex gap-3">
          <button type="button" className="flex-1 rounded-lg border border-[var(--border-color)] py-2 text-sm" onClick={onClose}>取消</button>
          <button type="submit" disabled={pending} className="flex-1 rounded-lg bg-[var(--accent)] py-2 text-sm text-white disabled:opacity-60">{pending ? '保存中...' : '保存'}</button>
        </div>
      </form>
    </div>
  )
}

function FeedModal({ feed, onClose }: { feed?: Feed; onClose: () => void }) {
  const updateFeed = useAdminUpdateFeed()
  const createFeed = useAdminCreateFeed()
  const [userID, setUserID] = useState<number>(feed?.user_id ?? 1)
  const [url, setURL] = useState(feed?.url ?? '')
  const [title, setTitle] = useState(feed?.title ?? '')
  const [siteURL, setSiteURL] = useState(feed?.site_url ?? '')
  const [description, setDescription] = useState(feed?.description ?? '')
  const [disabled, setDisabled] = useState(Boolean(feed?.disabled))
  const [errorCount, setErrorCount] = useState(feed?.error_count ?? 0)

  const isEdit = Boolean(feed)
  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (isEdit && feed) {
        await updateFeed.mutateAsync({
          id: feed.id,
          user_id: userID,
          url,
          title,
          site_url: siteURL,
          description,
          disabled,
          error_count: errorCount,
        })
        showToast('success', `Feed #${feed.id} 已更新`)
        onClose()
        return
      }
      await createFeed.mutateAsync({
        user_id: userID,
        url,
        title,
        site_url: siteURL,
        description,
      })
      showToast('success', 'Feed 创建成功')
      onClose()
    } catch (err) {
      showToast('error', err instanceof Error ? err.message : '保存 Feed 失败')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <form onSubmit={submit} className="w-full max-w-2xl rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)] p-6">
        <h3 className="mb-4 text-lg font-semibold text-[var(--text-primary)]">{isEdit ? '编辑订阅数据（SQLite）' : '新增订阅数据（SQLite）'}</h3>
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" type="number" value={userID} onChange={(e) => setUserID(Number(e.target.value))} placeholder="user_id" required />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={url} onChange={(e) => setURL(e.target.value)} placeholder="url" required />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="title" />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={siteURL} onChange={(e) => setSiteURL(e.target.value)} placeholder="site_url" />
        <textarea className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" rows={3} value={description} onChange={(e) => setDescription(e.target.value)} placeholder="description" />
        {isEdit && (
          <>
            <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" type="number" value={errorCount} onChange={(e) => setErrorCount(Number(e.target.value) || 0)} placeholder="error_count" />
            <label className="mb-4 flex items-center gap-2 text-sm text-[var(--text-secondary)]">
              <input type="checkbox" checked={disabled} onChange={(e) => setDisabled(e.target.checked)} />
              disabled
            </label>
          </>
        )}
        <div className="flex gap-3">
          <button type="button" className="flex-1 rounded-lg border border-[var(--border-color)] py-2 text-sm" onClick={onClose}>取消</button>
          <button type="submit" className="flex-1 rounded-lg bg-[var(--accent)] py-2 text-sm text-white">保存</button>
        </div>
      </form>
    </div>
  )
}

function SourceModal({ item, onClose }: { item?: RecommendedFeed; onClose: () => void }) {
  const createSource = useAdminCreateRecommendedFeed()
  const updateSource = useAdminUpdateRecommendedFeed()
  const [name, setName] = useState(item?.name ?? '')
  const [url, setURL] = useState(item?.url ?? '')
  const [description, setDescription] = useState(item?.description ?? '')
  const [category, setCategory] = useState(item?.category ?? '')
  const [icon, setIcon] = useState(item?.icon ?? '')

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (item) {
        await updateSource.mutateAsync({ ...item, name, url, description, category, icon })
        showToast('success', `数据源 ${name} 已更新`)
      } else {
        await createSource.mutateAsync({ name, url, description, category, icon })
        showToast('success', `数据源 ${name} 已创建`)
      }
      onClose()
    } catch (err) {
      showToast('error', err instanceof Error ? err.message : '保存数据源失败')
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <form onSubmit={submit} className="w-full max-w-2xl rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)] p-6">
        <h3 className="mb-4 text-lg font-semibold text-[var(--text-primary)]">{item ? '编辑数据源' : '新增数据源'}</h3>
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={name} onChange={(e) => setName(e.target.value)} placeholder="name" required />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={url} onChange={(e) => setURL(e.target.value)} placeholder="url" required />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={category} onChange={(e) => setCategory(e.target.value)} placeholder="category" />
        <input className="mb-3 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" value={icon} onChange={(e) => setIcon(e.target.value)} placeholder="icon" />
        <textarea className="mb-5 w-full rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" rows={3} value={description} onChange={(e) => setDescription(e.target.value)} placeholder="description" />
        <div className="flex gap-3">
          <button type="button" className="flex-1 rounded-lg border border-[var(--border-color)] py-2 text-sm" onClick={onClose}>取消</button>
          <button type="submit" className="flex-1 rounded-lg bg-[var(--accent)] py-2 text-sm text-white">保存</button>
        </div>
      </form>
    </div>
  )
}

export default function Admin() {
  const navigate = useNavigate()
  const [menu, setMenu] = useState<MenuKey>('overview')
  const [searchUser, setSearchUser] = useState('')
  const [searchFeed, setSearchFeed] = useState('')
  const [userModal, setUserModal] = useState<User | null | 'new'>(null)
  const [feedModal, setFeedModal] = useState<Feed | null | 'new'>(null)
  const [sourceModal, setSourceModal] = useState<RecommendedFeed | null | 'new'>(null)

  const { data: stats } = useAdminStats()
  const { data: settings } = useAdminSettings()
  const { data: usersData } = useAdminUsers({ q: searchUser || undefined })
  const { data: feedsData } = useAdminFeeds({ q: searchFeed || undefined, limit: 100 })
  const { data: sources = [] } = useAdminRecommendedFeeds()
  const deleteUser = useAdminDeleteUser()
  const deleteFeed = useAdminDeleteFeed()
  const deleteSource = useAdminDeleteRecommendedFeed()
  const updateSettings = useAdminUpdateSettings()
  const [fetchIntervalInput, setFetchIntervalInput] = useState('15')

  const users = usersData?.items ?? []
  const feeds = feedsData?.items ?? []

  React.useEffect(() => {
    if (settings?.fetch_interval_minutes) {
      setFetchIntervalInput(String(settings.fetch_interval_minutes))
    }
  }, [settings?.fetch_interval_minutes])

  const tabs = useMemo(
    () => [
      { key: 'overview' as const, label: '总览', icon: BarChart3 },
      { key: 'users' as const, label: '用户', icon: Users },
      { key: 'feeds' as const, label: 'SQLite 订阅', icon: Database },
      { key: 'sources' as const, label: '推荐数据源', icon: Rss },
    ],
    []
  )

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 dark:bg-[#1e2227] dark:text-slate-100">
      <div className="mx-auto flex max-w-[1500px]">
        <aside className="sticky top-0 h-screen w-64 border-r border-slate-200 bg-slate-100 p-5 dark:border-[#30363d] dark:bg-[#1a1d21]">
          <div className="mb-8 flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-slate-900 text-white ring-1 ring-slate-700/20 dark:bg-white dark:text-[#f7931e] dark:ring-white/20">
              <Shield className="h-5 w-5" />
            </div>
            <div>
              <p className="text-sm text-slate-600 dark:text-slate-300">rssreader</p>
              <p className="font-semibold text-slate-900 dark:text-slate-100">Admin Console</p>
            </div>
          </div>
          <nav className="space-y-2">
            {tabs.map((tab) => (
              <button key={tab.key} className={`flex w-full items-center gap-2 rounded-xl px-3 py-2 text-left text-sm ${menu === tab.key ? 'bg-slate-200 text-slate-900 font-medium dark:bg-[#2d333b] dark:text-slate-100 [&_svg]:!text-slate-700 dark:[&_svg]:!text-slate-100' : 'text-slate-600 hover:bg-slate-200 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-[#2d333b] dark:hover:text-slate-100 [&_svg]:!text-slate-500 dark:[&_svg]:!text-slate-300'}`} onClick={() => setMenu(tab.key)}>
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </button>
            ))}
          </nav>
        </aside>

        <main className="flex-1 p-8">
          <div className="mb-7 border-b border-slate-200 pb-5 dark:border-[#30363d]">
            <div className="flex items-center justify-between gap-3">
              <h1 className="text-4xl font-semibold text-slate-900 dark:text-slate-100">Settings</h1>
              <button
                type="button"
                onClick={() => navigate('/')}
                className="inline-flex items-center gap-1 rounded-full border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-100 dark:border-[#3b424d] dark:text-slate-200 dark:hover:bg-[#2d333b]"
              >
                <ChevronLeft className="h-4 w-4" />
                返回前台
              </button>
            </div>
            <div className="mt-4 flex flex-wrap gap-3 text-sm">
              {tabs.map((tab) => (
                <button key={tab.key} onClick={() => setMenu(tab.key)} className={`rounded-full px-4 py-1.5 ${menu === tab.key ? 'bg-slate-200 text-slate-900 font-medium dark:bg-[#2d333b] dark:text-slate-100' : 'text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-slate-100'}`}>
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {menu === 'overview' && (
            <div>
              <div className="mb-4 rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)] p-4">
                <h3 className="mb-2 text-base font-semibold">自动拉取设置</h3>
                <p className="mb-3 text-sm text-[var(--text-secondary)]">设置数据源自动轮询拉取间隔（分钟），保存后立即生效并写入 SQLite。</p>
                <div className="flex flex-wrap items-center gap-2">
                  <input
                    type="number"
                    min={1}
                    max={1440}
                    className="w-40 rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm"
                    value={fetchIntervalInput}
                    onChange={(e) => setFetchIntervalInput(e.target.value)}
                  />
                  <span className="text-sm text-[var(--text-secondary)]">分钟</span>
                  <button
                    type="button"
                    className="rounded-lg bg-[var(--accent)] px-3 py-2 text-sm text-white disabled:opacity-60"
                    disabled={updateSettings.isPending}
                    onClick={async () => {
                      const next = Number(fetchIntervalInput)
                      if (!Number.isInteger(next) || next < 1 || next > 1440) {
                        showToast('error', '轮询间隔必须是 1-1440 的整数分钟')
                        return
                      }
                      try {
                        await updateSettings.mutateAsync({ fetch_interval_minutes: next })
                        showToast('success', `自动拉取间隔已更新为 ${next} 分钟`)
                      } catch (err) {
                        showToast('error', err instanceof Error ? err.message : '保存设置失败')
                      }
                    }}
                  >
                    {updateSettings.isPending ? '保存中...' : '保存设置'}
                  </button>
                </div>
              </div>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
                <StatCard title="总用户" value={stats?.total_users || 0} icon={Users} color="bg-blue-500" />
                <StatCard title="启用用户" value={stats?.active_users || 0} icon={Check} color="bg-green-500" />
                <StatCard title="管理员" value={stats?.admin_users || 0} icon={Shield} color="bg-zinc-700" />
                <StatCard title="总订阅" value={stats?.total_feeds || 0} icon={Rss} color="bg-orange-500" />
              </div>
              <div className="mt-4 grid grid-cols-1 gap-4 md:grid-cols-3">
                <StatCard title="今日文章" value={stats?.articles_today || 0} icon={BarChart3} color="bg-teal-500" />
                <StatCard title="本周文章" value={stats?.articles_this_week || 0} icon={BarChart3} color="bg-cyan-600" />
                <StatCard title="总文章" value={stats?.total_articles || 0} icon={BarChart3} color="bg-indigo-500" />
              </div>
            </div>
          )}

          {menu === 'users' && (
            <section className="rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)]">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--border-color)] p-4">
                <h2 className="text-lg font-semibold">用户管理</h2>
                <div className="flex gap-2">
                  <input className="rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" placeholder="搜索用户名/邮箱" value={searchUser} onChange={(e) => setSearchUser(e.target.value)} />
                  <button className="flex items-center gap-2 rounded-lg bg-[var(--accent)] px-3 py-2 text-sm text-white" onClick={() => setUserModal('new')}>
                    <UserPlus className="h-4 w-4" /> 新增用户
                  </button>
                </div>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-[var(--border-color)] text-[var(--text-secondary)]">
                      <th className="px-4 py-3 text-left">用户</th>
                      <th className="px-4 py-3 text-left">角色</th>
                      <th className="px-4 py-3 text-left">状态</th>
                      <th className="px-4 py-3 text-left">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((u) => (
                      <tr key={u.id} className="border-b border-[var(--border-color)]">
                        <td className="px-4 py-3">
                          <div className="font-medium">{u.username}</div>
                          <div className="text-xs text-[var(--text-secondary)]">{u.email || '-'}</div>
                        </td>
                        <td className="px-4 py-3">{u.is_admin ? '管理员' : '普通用户'}</td>
                        <td className="px-4 py-3">{u.status === 1 ? '启用' : '禁用'}</td>
                        <td className="px-4 py-3">
                          <div className="flex gap-2">
                            <button className="rounded-md border border-[var(--border-color)] p-1.5" onClick={() => setUserModal(u)}><Edit2 className="h-4 w-4" /></button>
                            <button
                              className="rounded-md border border-red-500/40 p-1.5 text-red-500"
                              onClick={async () => {
                                if (!window.confirm(`确认删除用户 ${u.username}？`)) return
                                try {
                                  await deleteUser.mutateAsync(u.id)
                                  showToast('success', `用户 ${u.username} 已删除`)
                                } catch (err) {
                                  showToast('error', err instanceof Error ? err.message : '删除用户失败')
                                }
                              }}
                            ><Trash2 className="h-4 w-4" /></button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          )}

          {menu === 'feeds' && (
            <section className="rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)]">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--border-color)] p-4">
                <h2 className="text-lg font-semibold">SQLite 订阅表管理</h2>
                <div className="flex gap-2">
                  <input className="rounded-lg border border-[var(--border-color)] bg-[var(--bg-primary)] px-3 py-2 text-sm" placeholder="搜索 title/url/user" value={searchFeed} onChange={(e) => setSearchFeed(e.target.value)} />
                  <button className="flex items-center gap-2 rounded-lg bg-[var(--accent)] px-3 py-2 text-sm text-white" onClick={() => setFeedModal('new')}>
                    <Plus className="h-4 w-4" /> 新增 Feed
                  </button>
                </div>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-[var(--border-color)] text-[var(--text-secondary)]">
                      <th className="px-4 py-3 text-left">ID</th>
                      <th className="px-4 py-3 text-left">用户</th>
                      <th className="px-4 py-3 text-left">标题</th>
                      <th className="px-4 py-3 text-left">URL</th>
                      <th className="px-4 py-3 text-left">状态</th>
                      <th className="px-4 py-3 text-left">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {feeds.map((f) => (
                      <tr key={f.id} className="border-b border-[var(--border-color)]">
                        <td className="px-4 py-3">{f.id}</td>
                        <td className="px-4 py-3">{f.username || `#${f.user_id}`}</td>
                        <td className="px-4 py-3">{f.title || '-'}</td>
                        <td className="px-4 py-3"><span className="line-clamp-1">{f.url}</span></td>
                        <td className="px-4 py-3">{f.disabled ? 'disabled' : 'active'}</td>
                        <td className="px-4 py-3">
                          <div className="flex gap-2">
                            <button className="rounded-md border border-[var(--border-color)] p-1.5" onClick={() => setFeedModal(f)}><Edit2 className="h-4 w-4" /></button>
                            <button
                              className="rounded-md border border-red-500/40 p-1.5 text-red-500"
                              onClick={async () => {
                                if (!window.confirm(`确认删除 Feed #${f.id}？`)) return
                                try {
                                  await deleteFeed.mutateAsync(f.id)
                                  showToast('success', `Feed #${f.id} 已删除`)
                                } catch (err) {
                                  showToast('error', err instanceof Error ? err.message : '删除 Feed 失败')
                                }
                              }}
                            ><Trash2 className="h-4 w-4" /></button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          )}

          {menu === 'sources' && (
            <section className="rounded-2xl border border-[var(--border-color)] bg-[var(--bg-secondary)]">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[var(--border-color)] p-4">
                <h2 className="text-lg font-semibold">推荐数据源管理</h2>
                <button className="flex items-center gap-2 rounded-lg bg-[var(--accent)] px-3 py-2 text-sm text-white" onClick={() => setSourceModal('new')}>
                  <Plus className="h-4 w-4" /> 添加数据源
                </button>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-[var(--border-color)] text-[var(--text-secondary)]">
                      <th className="px-4 py-3 text-left">名称</th>
                      <th className="px-4 py-3 text-left">分类</th>
                      <th className="px-4 py-3 text-left">URL</th>
                      <th className="px-4 py-3 text-left">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sources.map((s) => (
                      <tr key={s.id} className="border-b border-[var(--border-color)]">
                        <td className="px-4 py-3">{s.name}</td>
                        <td className="px-4 py-3">{s.category || '-'}</td>
                        <td className="px-4 py-3"><span className="line-clamp-1">{s.url}</span></td>
                        <td className="px-4 py-3">
                          <div className="flex gap-2">
                            <button className="rounded-md border border-[var(--border-color)] p-1.5" onClick={() => setSourceModal(s)}><Edit2 className="h-4 w-4" /></button>
                            <button
                              className="rounded-md border border-red-500/40 p-1.5 text-red-500"
                              onClick={async () => {
                                if (!window.confirm(`确认删除数据源 ${s.name}？`)) return
                                try {
                                  await deleteSource.mutateAsync(s.id)
                                  showToast('success', `数据源 ${s.name} 已删除`)
                                } catch (err) {
                                  showToast('error', err instanceof Error ? err.message : '删除数据源失败')
                                }
                              }}
                            ><Trash2 className="h-4 w-4" /></button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </section>
          )}
        </main>
      </div>

      {userModal && <UserModal user={userModal === 'new' ? undefined : userModal} onClose={() => setUserModal(null)} />}
      {feedModal && <FeedModal feed={feedModal === 'new' ? undefined : feedModal} onClose={() => setFeedModal(null)} />}
      {sourceModal && <SourceModal item={sourceModal === 'new' ? undefined : sourceModal} onClose={() => setSourceModal(null)} />}
    </div>
  )
}
