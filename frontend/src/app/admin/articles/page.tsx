'use client';

import { useState, useEffect } from 'react';
import AdminLayout from '@/components/AdminLayout';
import { adminAPI, type Article, type PaginationResponse } from '@/lib/api';
import {
  DocumentTextIcon,
  SparklesIcon,
  CheckCircleIcon,
  ClockIcon,
  DocumentIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

export default function ArticlesManagement() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [pagination, setPagination] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  
  // 过滤和排序状态
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(10);
  const [statusFilter, setStatusFilter] = useState('');
  const [aiFilter, setAiFilter] = useState('');

  const fetchArticles = async () => {
    setLoading(true);
    try {
      const params: any = {
        page: currentPage,
        page_size: pageSize,
      };
      
      if (statusFilter) params.status = statusFilter;
      if (aiFilter) params.is_ai_generated = aiFilter;
      
      const response: PaginationResponse<Article> = await adminAPI.getArticles(params);
      setArticles(response.data);
      setPagination(response.pagination);
    } catch (error: any) {
      console.error('Failed to fetch articles:', error);
      setError('获取文章列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchArticles();
  }, [currentPage, statusFilter, aiFilter]);

  const handleStatusChange = async (articleId: number, newStatus: string) => {
    try {
      await adminAPI.updateArticleStatus(articleId, { status: newStatus });
      // 刷新文章列表
      fetchArticles();
    } catch (error: any) {
      console.error('Failed to update article status:', error);
      alert('更新文章状态失败');
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      published: { color: 'bg-green-100 text-green-800', text: '已发布', icon: CheckCircleIcon },
      pending: { color: 'bg-yellow-100 text-yellow-800', text: '待审核', icon: ClockIcon },
      draft: { color: 'bg-gray-100 text-gray-800', text: '草稿', icon: DocumentIcon },
      rejected: { color: 'bg-red-100 text-red-800', text: '已拒绝', icon: XCircleIcon },
    };
    
    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.draft;
    const Icon = config.icon;
    
    return (
      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.color}`}>
        <Icon className="h-3 w-3 mr-1" />
        {config.text}
      </span>
    );
  };

  if (loading && articles.length === 0) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">加载中...</div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="space-y-6">
        {/* 页面标题和过滤器 */}
        <div className="sm:flex sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-semibold text-gray-900">文章管理</h1>
            <p className="mt-1 text-sm text-gray-600">
              管理所有文章内容，审核和发布
            </p>
          </div>
        </div>

        {/* 过滤器 */}
        <div className="bg-white shadow rounded-lg p-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">状态筛选</label>
              <select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setCurrentPage(1);
                }}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              >
                <option value="">所有状态</option>
                <option value="published">已发布</option>
                <option value="pending">待审核</option>
                <option value="draft">草稿</option>
                <option value="rejected">已拒绝</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700">AI生成筛选</label>
              <select
                value={aiFilter}
                onChange={(e) => {
                  setAiFilter(e.target.value);
                  setCurrentPage(1);
                }}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500"
              >
                <option value="">全部</option>
                <option value="true">AI生成</option>
                <option value="false">人工创作</option>
              </select>
            </div>
            
            <div className="flex items-end">
              <button
                onClick={() => {
                  setStatusFilter('');
                  setAiFilter('');
                  setCurrentPage(1);
                }}
                className="w-full px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
              >
                重置过滤器
              </button>
            </div>
          </div>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <div className="text-red-800">{error}</div>
          </div>
        )}

        {/* 文章列表 */}
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <ul className="divide-y divide-gray-200">
            {articles.map((article) => (
              <li key={article.id} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-3">
                      <DocumentTextIcon className="h-5 w-5 text-gray-400" />
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">
                          {article.title}
                        </p>
                        <div className="flex items-center space-x-4 mt-1">
                          <p className="text-sm text-gray-500">
                            作者: {article.author?.username || '未知'}
                          </p>
                          <p className="text-sm text-gray-500">
                            分类: {article.category?.name || '未分类'}
                          </p>
                          <p className="text-sm text-gray-500">
                            浏览: {article.view_count || 0}
                          </p>
                          <p className="text-sm text-gray-500">
                            {new Date(article.created_at).toLocaleDateString()}
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-3">
                    {/* AI生成标识 */}
                    {article.is_ai_generated && (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                        <SparklesIcon className="h-3 w-3 mr-1" />
                        AI生成
                      </span>
                    )}
                    
                    {/* 状态标识 */}
                    {getStatusBadge(article.status)}
                    
                    {/* 操作按钮 */}
                    <div className="flex space-x-2">
                      {article.status === 'pending' && (
                        <>
                          <button
                            onClick={() => handleStatusChange(article.id, 'published')}
                            className="text-green-600 hover:text-green-700 text-sm font-medium"
                          >
                            通过
                          </button>
                          <button
                            onClick={() => handleStatusChange(article.id, 'rejected')}
                            className="text-red-600 hover:text-red-700 text-sm font-medium"
                          >
                            拒绝
                          </button>
                        </>
                      )}
                      
                      {article.status === 'published' && (
                        <button
                          onClick={() => handleStatusChange(article.id, 'draft')}
                          className="text-yellow-600 hover:text-yellow-700 text-sm font-medium"
                        >
                          下架
                        </button>
                      )}
                      
                      {(article.status === 'draft' || article.status === 'rejected') && (
                        <button
                          onClick={() => handleStatusChange(article.id, 'pending')}
                          className="text-blue-600 hover:text-blue-700 text-sm font-medium"
                        >
                          重新提交
                        </button>
                      )}
                    </div>
                  </div>
                </div>
                
                {/* 文章摘要 */}
                {article.summary && (
                  <div className="mt-2 ml-8">
                    <p className="text-sm text-gray-600 line-clamp-2">
                      {article.summary}
                    </p>
                  </div>
                )}
              </li>
            ))}
          </ul>
          
          {articles.length === 0 && !loading && (
            <div className="text-center py-12">
              <DocumentTextIcon className="mx-auto h-12 w-12 text-gray-400" />
              <h3 className="mt-2 text-sm font-medium text-gray-900">暂无文章</h3>
              <p className="mt-1 text-sm text-gray-500">没有找到符合条件的文章</p>
            </div>
          )}
        </div>

        {/* 分页 */}
        {pagination && pagination.total_pages > 1 && (
          <div className="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
            <div className="flex-1 flex justify-between sm:hidden">
              <button
                onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                disabled={currentPage <= 1}
                className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              <button
                onClick={() => setCurrentPage(Math.min(pagination.total_pages, currentPage + 1))}
                disabled={currentPage >= pagination.total_pages}
                className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                下一页
              </button>
            </div>
            <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  显示第 <span className="font-medium">{(currentPage - 1) * pageSize + 1}</span> 到{' '}
                  <span className="font-medium">
                    {Math.min(currentPage * pageSize, pagination.total)}
                  </span>{' '}
                  条，共 <span className="font-medium">{pagination.total}</span> 条记录
                </p>
              </div>
              <div>
                <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage <= 1}
                    className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    上一页
                  </button>
                  <span className="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700">
                    第 {currentPage} 页，共 {pagination.total_pages} 页
                  </span>
                  <button
                    onClick={() => setCurrentPage(Math.min(pagination.total_pages, currentPage + 1))}
                    disabled={currentPage >= pagination.total_pages}
                    className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    下一页
                  </button>
                </nav>
              </div>
            </div>
          </div>
        )}
      </div>
    </AdminLayout>
  );
}