'use client';

import { useState, useEffect } from 'react';
import UserLayout from '@/components/UserLayout';
import { articleAPI, categoryAPI, type Article, type Category, type PaginationResponse } from '@/lib/api';
import Link from 'next/link';
import {
  DocumentTextIcon,
  SparklesIcon,
  EyeIcon,
  CalendarIcon,
  UserIcon,
  FolderIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/outline';

export default function ArticlesPage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [pagination, setPagination] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  
  // 筛选状态
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [sortOrder, setSortOrder] = useState('desc');

  const fetchArticles = async () => {
    setLoading(true);
    try {
      const params: any = {
        page: currentPage,
        page_size: 12,
        status: 'published', // 只显示已发布的文章
        sort: sortBy,
        order: sortOrder,
      };
      
      if (selectedCategory) params.category_id = selectedCategory;
      if (searchKeyword) params.q = searchKeyword;
      
      const response: PaginationResponse<Article> = await articleAPI.search(params);
      setArticles(response.data);
      setPagination(response.pagination);
    } catch (error) {
      console.error('Failed to fetch articles:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchCategories = async () => {
    try {
      const response = await categoryAPI.getCategories();
      setCategories(response.data || []);
    } catch (error) {
      console.error('Failed to fetch categories:', error);
    }
  };

  useEffect(() => {
    fetchCategories();
  }, []);

  useEffect(() => {
    fetchArticles();
  }, [currentPage, selectedCategory, searchKeyword, sortBy, sortOrder]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    fetchArticles();
  };

  const handleFilterChange = (key: string, value: string) => {
    setCurrentPage(1);
    if (key === 'category') setSelectedCategory(value);
    if (key === 'sort') setSortBy(value);
    if (key === 'order') setSortOrder(value);
  };

  return (
    <UserLayout>
      <div className="max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        {/* 页面标题 */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-extrabold text-gray-900 sm:text-4xl">
            📚 攻略大全
          </h1>
          <p className="mt-3 max-w-2xl mx-auto text-xl text-gray-500 sm:mt-4">
            最全面的幻兽帕鲁游戏攻略内容
          </p>
        </div>

        {/* 搜索和筛选 */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <form onSubmit={handleSearch} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {/* 搜索框 */}
              <div className="md:col-span-2">
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    type="text"
                    value={searchKeyword}
                    onChange={(e) => setSearchKeyword(e.target.value)}
                    placeholder="搜索攻略标题、内容..."
                    className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-primary-500 focus:border-primary-500"
                  />
                </div>
              </div>

              {/* 分类筛选 */}
              <div>
                <select
                  value={selectedCategory}
                  onChange={(e) => handleFilterChange('category', e.target.value)}
                  className="block w-full px-3 py-2 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                >
                  <option value="">所有分类</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.id.toString()}>
                      {category.name}
                    </option>
                  ))}
                </select>
              </div>

              {/* 排序方式 */}
              <div>
                <select
                  value={`${sortBy}-${sortOrder}`}
                  onChange={(e) => {
                    const [sort, order] = e.target.value.split('-');
                    handleFilterChange('sort', sort);
                    handleFilterChange('order', order);
                  }}
                  className="block w-full px-3 py-2 border border-gray-300 bg-white rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500"
                >
                  <option value="created_at-desc">最新发布</option>
                  <option value="view_count-desc">热门阅读</option>
                  <option value="title-asc">标题A-Z</option>
                  <option value="updated_at-desc">最近更新</option>
                </select>
              </div>
            </div>

            <div className="flex justify-between items-center">
              <div className="text-sm text-gray-500">
                {pagination && (
                  <span>
                    共找到 {pagination.total} 篇攻略
                  </span>
                )}
              </div>
              <button
                type="submit"
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
              >
                搜索
              </button>
            </div>
          </form>
        </div>

        {/* 文章列表 */}
        {loading ? (
          <div className="flex items-center justify-center h-64">
            <div className="text-lg text-gray-500">加载中...</div>
          </div>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {articles.map((article) => (
              <article
                key={article.id}
                className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow overflow-hidden"
              >
                <div className="p-6">
                  {/* 分类和AI标识 */}
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center space-x-2 text-sm text-gray-500">
                      <FolderIcon className="h-4 w-4" />
                      <span>{article.category?.name || '未分类'}</span>
                    </div>
                    {article.is_ai_generated && (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                        <SparklesIcon className="h-3 w-3 mr-1" />
                        AI生成
                      </span>
                    )}
                  </div>

                  {/* 标题 */}
                  <Link href={`/articles/${article.id}`}>
                    <h3 className="text-lg font-semibold text-gray-900 hover:text-primary-600 line-clamp-2 mb-3">
                      {article.title}
                    </h3>
                  </Link>

                  {/* 摘要 */}
                  <p className="text-gray-600 text-sm line-clamp-3 mb-4">
                    {article.summary || article.content.substring(0, 120) + '...'}
                  </p>

                  {/* 元信息 */}
                  <div className="flex items-center justify-between text-xs text-gray-500">
                    <div className="flex items-center space-x-3">
                      <div className="flex items-center">
                        <UserIcon className="h-3 w-3 mr-1" />
                        {article.author?.username || '匿名'}
                      </div>
                      <div className="flex items-center">
                        <EyeIcon className="h-3 w-3 mr-1" />
                        {article.view_count || 0}
                      </div>
                    </div>
                    <div className="flex items-center">
                      <CalendarIcon className="h-3 w-3 mr-1" />
                      {new Date(article.created_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              </article>
            ))}
          </div>
        )}

        {/* 空状态 */}
        {!loading && articles.length === 0 && (
          <div className="text-center py-12">
            <DocumentTextIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">暂无攻略</h3>
            <p className="mt-1 text-sm text-gray-500">
              {searchKeyword || selectedCategory
                ? '没有找到符合条件的攻略'
                : '还没有发布任何攻略内容'}
            </p>
          </div>
        )}

        {/* 分页 */}
        {pagination && pagination.total_pages > 1 && (
          <div className="mt-8 flex items-center justify-between">
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
                  显示第 <span className="font-medium">{(currentPage - 1) * 12 + 1}</span> 到{' '}
                  <span className="font-medium">
                    {Math.min(currentPage * 12, pagination.total)}
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
                  
                  {/* 页码显示 */}
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
    </UserLayout>
  );
}