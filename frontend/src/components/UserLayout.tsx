'use client';

import { ReactNode, useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import Cookies from 'js-cookie';
import { authAPI } from '@/lib/api';
import {
  HomeIcon,
  MagnifyingGlassIcon,
  DocumentTextIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  XMarkIcon,
  PlusIcon,
} from '@heroicons/react/24/outline';

interface UserLayoutProps {
  children: ReactNode;
}

interface NavItem {
  name: string;
  href: string;
  icon: any;
  requireAuth?: boolean;
}

const navigation: NavItem[] = [
  { name: '🏠 首页', href: '/', icon: HomeIcon },
  { name: '🔍 搜索', href: '/search', icon: MagnifyingGlassIcon },
  { name: '📚 攻略', href: '/articles', icon: DocumentTextIcon },
  { name: '✍️ 发布', href: '/write', icon: PlusIcon, requireAuth: true },
];

export default function UserLayout({ children }: UserLayoutProps) {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  // 跳过管理页面和登录页面的认证检查
  const isAdminPage = pathname?.startsWith('/admin');
  const isAuthPage = pathname === '/login' || pathname === '/register';

  useEffect(() => {
    const checkAuth = async () => {
      if (isAdminPage || isAuthPage) {
        setLoading(false);
        return;
      }

      const token = Cookies.get('token');
      if (token) {
        try {
          const response = await authAPI.getProfile();
          setUser(response.data);
        } catch (error) {
          console.error('Auth check failed:', error);
          Cookies.remove('token');
        }
      }
      setLoading(false);
    };

    checkAuth();
  }, [isAdminPage, isAuthPage]);

  const handleLogout = () => {
    Cookies.remove('token');
    setUser(null);
    router.push('/');
  };

  const filteredNavigation = navigation.filter(item => 
    !item.requireAuth || user
  );

  if (loading && !isAdminPage && !isAuthPage) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">加载中...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 顶部导航栏 */}
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            {/* 左侧：Logo和导航 */}
            <div className="flex items-center">
              <Link href="/" className="flex items-center space-x-2">
                <span className="text-2xl">🎮</span>
                <span className="font-bold text-xl text-gray-900">
                  幻兽帕鲁攻略网
                </span>
              </Link>

              {/* 桌面端导航 */}
              <div className="hidden md:ml-10 md:flex md:space-x-8">
                {filteredNavigation.map((item) => {
                  const current = pathname === item.href;
                  return (
                    <Link
                      key={item.name}
                      href={item.href}
                      className={`${
                        current
                          ? 'border-primary-500 text-primary-600'
                          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                      } inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium`}
                    >
                      <item.icon className="h-4 w-4 mr-1" />
                      {item.name}
                    </Link>
                  );
                })}
              </div>
            </div>

            {/* 右侧：用户菜单 */}
            <div className="flex items-center space-x-4">
              {user ? (
                <div className="flex items-center space-x-4">
                  <span className="text-sm text-gray-700">
                    欢迎，{user.nickname || user.username}
                  </span>
                  {user.role === 'admin' && (
                    <Link
                      href="/admin"
                      className="text-sm text-primary-600 hover:text-primary-700 font-medium"
                    >
                      管理后台
                    </Link>
                  )}
                  <button
                    onClick={handleLogout}
                    className="text-gray-400 hover:text-gray-500 p-1"
                  >
                    <ArrowRightOnRectangleIcon className="h-5 w-5" />
                  </button>
                </div>
              ) : (
                <div className="flex items-center space-x-4">
                  <Link
                    href="/login"
                    className="text-sm text-gray-700 hover:text-gray-900"
                  >
                    登录
                  </Link>
                  <Link
                    href="/register"
                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700"
                  >
                    注册
                  </Link>
                </div>
              )}

              {/* 移动端菜单按钮 */}
              <div className="md:hidden">
                <button
                  onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                  className="text-gray-400 hover:text-gray-500 p-2"
                >
                  {mobileMenuOpen ? (
                    <XMarkIcon className="h-6 w-6" />
                  ) : (
                    <Bars3Icon className="h-6 w-6" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* 移动端菜单 */}
        {mobileMenuOpen && (
          <div className="md:hidden">
            <div className="pt-2 pb-3 space-y-1">
              {filteredNavigation.map((item) => {
                const current = pathname === item.href;
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    className={`${
                      current
                        ? 'bg-primary-50 border-primary-500 text-primary-700'
                        : 'border-transparent text-gray-600 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800'
                    } block pl-3 pr-4 py-2 border-l-4 text-base font-medium`}
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    <item.icon className="h-5 w-5 inline mr-2" />
                    {item.name}
                  </Link>
                );
              })}
            </div>
          </div>
        )}
      </nav>

      {/* 主内容区域 */}
      <main className="flex-1">
        {children}
      </main>

      {/* 页脚 */}
      <footer className="bg-white border-t border-gray-200">
        <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div>
              <h3 className="text-sm font-semibold text-gray-900 tracking-wider uppercase">
                关于我们
              </h3>
              <p className="mt-4 text-base text-gray-500">
                专业的幻兽帕鲁游戏攻略平台，提供最新最全的游戏资讯和攻略内容。
              </p>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-900 tracking-wider uppercase">
                快速链接
              </h3>
              <ul className="mt-4 space-y-4">
                <li>
                  <Link href="/articles" className="text-base text-gray-500 hover:text-gray-900">
                    攻略大全
                  </Link>
                </li>
                <li>
                  <Link href="/search" className="text-base text-gray-500 hover:text-gray-900">
                    搜索攻略
                  </Link>
                </li>
              </ul>
            </div>
            <div>
              <h3 className="text-sm font-semibold text-gray-900 tracking-wider uppercase">
                联系我们
              </h3>
              <p className="mt-4 text-base text-gray-500">
                邮箱: contact@palu-wiki.com
              </p>
            </div>
          </div>
          <div className="mt-8 border-t border-gray-200 pt-8">
            <p className="text-base text-gray-400 xl:text-center">
              &copy; 2024 幻兽帕鲁攻略网. All rights reserved.
            </p>
          </div>
        </div>
      </footer>
    </div>
  );
}