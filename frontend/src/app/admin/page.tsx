'use client';

import { useState, useEffect } from 'react';
import AdminLayout from '@/components/AdminLayout';
import { adminAPI, type DashboardStats, type Article, type User } from '@/lib/api';
import {
  UsersIcon,
  DocumentTextIcon,
  FolderIcon,
  ChatBubbleLeftRightIcon,
  SparklesIcon,
  ClockIcon,
  CheckCircleIcon,
  DocumentIcon,
} from '@heroicons/react/24/outline';

interface DashboardData {
  stats: DashboardStats;
  recent_articles: Article[];
  recent_users: User[];
}

export default function AdminDashboard() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const response = await adminAPI.getDashboard();
        setData(response);
      } catch (error: any) {
        console.error('Failed to fetch dashboard data:', error);
        setError('获取仪表板数据失败');
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  if (loading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">加载中...</div>
        </div>
      </AdminLayout>
    );
  }

  if (error) {
    return (
      <AdminLayout>
        <div className="text-red-600 text-center">{error}</div>
      </AdminLayout>
    );
  }

  const stats = data?.stats;
  const recentArticles = data?.recent_articles || [];
  const recentUsers = data?.recent_users || [];

  const statCards = [
    {
      name: '总用户数',
      value: stats?.total_users || 0,
      icon: UsersIcon,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100',
    },
    {
      name: '总文章数',
      value: stats?.total_articles || 0,
      icon: DocumentTextIcon,
      color: 'text-green-600',
      bgColor: 'bg-green-100',
    },
    {
      name: 'AI生成文章',
      value: stats?.ai_articles || 0,
      icon: SparklesIcon,
      color: 'text-purple-600',
      bgColor: 'bg-purple-100',
    },
    {
      name: '待审核文章',
      value: stats?.pending_articles || 0,
      icon: ClockIcon,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-100',
    },
    {
      name: '已发布文章',
      value: stats?.published_articles || 0,
      icon: CheckCircleIcon,
      color: 'text-green-600',
      bgColor: 'bg-green-100',
    },
    {
      name: '草稿文章',
      value: stats?.draft_articles || 0,
      icon: DocumentIcon,
      color: 'text-gray-600',
      bgColor: 'bg-gray-100',
    },
    {
      name: '分类总数',
      value: stats?.total_categories || 0,
      icon: FolderIcon,
      color: 'text-indigo-600',
      bgColor: 'bg-indigo-100',
    },
    {
      name: '评论总数',
      value: stats?.total_comments || 0,
      icon: ChatBubbleLeftRightIcon,
      color: 'text-pink-600',
      bgColor: 'bg-pink-100',
    },
  ];

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* 页面标题 */}
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">仪表板</h1>
          <p className="mt-1 text-sm text-gray-600">
            网站运营数据总览
          </p>
        </div>

        {/* 统计卡片 */}
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
          {statCards.map((item) => (
            <div
              key={item.name}
              className="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden"
            >
              <dt>
                <div className={`absolute rounded-md p-3 ${item.bgColor}`}>
                  <item.icon className={`h-6 w-6 ${item.color}`} />
                </div>
                <p className="ml-16 text-sm font-medium text-gray-500 truncate">
                  {item.name}
                </p>
              </dt>
              <dd className="ml-16 pb-6 flex items-baseline sm:pb-7">
                <p className="text-2xl font-semibold text-gray-900">
                  {item.value.toLocaleString()}
                </p>
              </dd>
            </div>
          ))}
        </div>

        {/* 最近活动 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 最近文章 */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                最新文章
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                最近创建的5篇文章
              </p>
            </div>
            <div className="border-t border-gray-200">
              <ul className="divide-y divide-gray-200">
                {recentArticles.map((article) => (
                  <li key={article.id} className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {article.title}
                          {article.is_ai_generated && (
                            <span className="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                              AI生成
                            </span>
                          )}
                        </p>
                        <p className="text-sm text-gray-500">
                          作者: {article.author?.username || '未知'} | 
                          状态: <span className={`${
                            article.status === 'published' ? 'text-green-600' :
                            article.status === 'pending' ? 'text-yellow-600' :
                            'text-gray-600'
                          }`}>
                            {article.status === 'published' ? '已发布' :
                             article.status === 'pending' ? '待审核' :
                             article.status === 'draft' ? '草稿' : article.status}
                          </span>
                        </p>
                      </div>
                      <div className="text-sm text-gray-500">
                        {new Date(article.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  </li>
                ))}
                {recentArticles.length === 0 && (
                  <li className="px-4 py-4 sm:px-6 text-center text-gray-500">
                    暂无文章数据
                  </li>
                )}
              </ul>
            </div>
          </div>

          {/* 最近用户 */}
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                新用户
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                最近注册的5个用户
              </p>
            </div>
            <div className="border-t border-gray-200">
              <ul className="divide-y divide-gray-200">
                {recentUsers.map((user) => (
                  <li key={user.id} className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {user.nickname || user.username}
                          <span className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            user.role === 'admin' ? 'bg-red-100 text-red-800' :
                            user.role === 'editor' ? 'bg-blue-100 text-blue-800' :
                            'bg-gray-100 text-gray-800'
                          }`}>
                            {user.role === 'admin' ? '管理员' :
                             user.role === 'editor' ? '编辑' : '用户'}
                          </span>
                        </p>
                        <p className="text-sm text-gray-500">
                          {user.email}
                        </p>
                      </div>
                      <div className="text-sm text-gray-500">
                        {new Date(user.created_at).toLocaleDateString()}
                      </div>
                    </div>
                  </li>
                ))}
                {recentUsers.length === 0 && (
                  <li className="px-4 py-4 sm:px-6 text-center text-gray-500">
                    暂无用户数据
                  </li>
                )}
              </ul>
            </div>
          </div>
        </div>

        {/* 快速操作 */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg leading-6 font-medium text-gray-900">
              快速操作
            </h3>
          </div>
          <div className="px-4 py-4 sm:px-6">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <button className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                <DocumentTextIcon className="h-4 w-4 mr-2" />
                查看待审核文章
              </button>
              <button className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                <UsersIcon className="h-4 w-4 mr-2" />
                用户管理
              </button>
              <button className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                <SparklesIcon className="h-4 w-4 mr-2" />
                AI内容审核
              </button>
            </div>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
}