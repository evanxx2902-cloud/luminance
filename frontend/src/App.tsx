import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'

// 受保护路由组件 - 已登录则显示子组件，未登录则跳转到登录页
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}

// 公开路由组件 - 已登录则跳转到首页，未登录则显示子组件
function PublicRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  return !isAuthenticated ? <>{children}</> : <Navigate to="/" replace />
}

// 首页占位组件
function HomePage() {
  const { user, logout } = useAuth()
  return (
    <main className="min-h-screen bg-surface flex items-center justify-center">
      <div
        className="bg-white rounded-clay border-[3px] border-primary p-10 max-w-md w-full text-center"
        style={{ boxShadow: 'var(--shadow-clay)' }}
      >
        <h1 className="font-heading text-4xl font-bold text-primary mb-3">
          Luminance
        </h1>
        <p className="font-body text-text text-lg mb-2">
          欢迎回来，{user?.username || '用户'}！
        </p>
        <p className="font-body text-secondary text-sm mb-6">
          教育平台 — 脚手架已就绪
        </p>
        <button
          onClick={logout}
          className="px-6 py-2 rounded-clay border-[3px] border-cta bg-cta text-white font-heading font-bold hover:bg-[#ea6c10] transition-colors cursor-pointer"
          style={{ boxShadow: 'var(--shadow-clay-cta)' }}
        >
          退出登录
        </button>
      </div>
    </main>
  )
}

function AppRoutes() {
  return (
    <Routes>
      {/* 公开路由 - 未登录才能访问 */}
      <Route
        path="/login"
        element={
          <PublicRoute>
            <LoginPage />
          </PublicRoute>
        }
      />
      <Route
        path="/register"
        element={
          <PublicRoute>
            <RegisterPage />
          </PublicRoute>
        }
      />

      {/* 受保护路由 - 需要登录 */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <HomePage />
          </ProtectedRoute>
        }
      />

      {/* 兜底路由 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthProvider>
  )
}
