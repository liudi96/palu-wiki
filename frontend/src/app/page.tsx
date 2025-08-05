'use client';

import { useState, useEffect } from 'react';
import UserLayout from '@/components/UserLayout';
import { articleAPI, categoryAPI, type Article, type Category } from '@/lib/api';
import Link from 'next/link';
import {
  DocumentTextIcon,
  SparklesIcon,
  EyeIcon,
  CalendarIcon,
  UserIcon,
  FolderIcon,
} from '@heroicons/react/24/outline';

export default function HomePage() {
  const [featuredArticles, setFeaturedArticles] = useState<Article[]>([]);
  const [recentArticles, setRecentArticles] = useState<Article[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // 获取最新文章
        const articlesResponse = await articleAPI.getArticles({ page: 1, limit: 8 });
        const articles = articlesResponse.data || [];
        
        // 前4篇作为精选，后4篇作为最新
        setFeaturedArticles(articles.slice(0, 4));
        setRecentArticles(articles.slice(4, 8));

        // 获取分类
        const categoriesResponse = await categoryAPI.getCategories();
        setCategories(categoriesResponse.data || []);
      } catch (error) {
        console.error('Failed to fetch home page data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return (
      <UserLayout>
        <div className="flex items-center justify-center h-64">
          <div className="text-lg">加载中...</div>
        </div>
      </UserLayout>
    );
  }

  return (
    <UserLayout>
      {/* 英雄区域 */}
      <div className="bg-gradient-to-r from-primary-600 to-primary-700">
        <div className="max-w-7xl mx-auto py-16 px-4 sm:py-24 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl font-extrabold text-white sm:text-5xl md:text-6xl">
              🎮 幻兽帕鲁攻略网
            </h1>
            <p className="mt-3 max-w-md mx-auto text-base text-primary-200 sm:text-lg md:mt-5 md:text-xl md:max-w-3xl">
              最全面的帕鲁游戏攻略平台，AI智能生成，专业玩家分享，助你成为帕鲁大师！
            </p>
            <div className="mt-5 max-w-md mx-auto sm:flex sm:justify-center md:mt-8">
              <div className="rounded-md shadow">
                <Link
                  href="/articles"
                  className="w-full flex items-center justify-center px-8 py-3 border border-transparent text-base font-medium rounded-md text-primary-700 bg-white hover:bg-gray-50 md:py-4 md:text-lg md:px-10"
                >
                  📚 浏览攻略
                </Link>
              </div>
              <div className="mt-3 rounded-md shadow sm:mt-0 sm:ml-3">
                <Link
                  href="/search"
                  className="w-full flex items-center justify-center px-8 py-3 border border-transparent text-base font-medium rounded-md text-white bg-primary-500 hover:bg-primary-600 md:py-4 md:text-lg md:px-10"
                >
                  🔍 搜索内容
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 精选攻略 */}
      <div className="max-w-7xl mx-auto px-4 py-12 sm:px-6 lg:px-8">
        <div className="text-center">
          <h2 className="text-3xl font-extrabold text-gray-900 sm:text-4xl">
            🌟 精选攻略
          </h2>
          <p className="mt-3 max-w-2xl mx-auto text-xl text-gray-500 sm:mt-4">
            编辑推荐的优质攻略内容
          </p>
        </div>

        <div className="mt-12 grid gap-8 lg:grid-cols-2">
          {featuredArticles.map((article) => (
            <article
              key={article.id}
              className="flex flex-col rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow"
            >
              <div className="flex-1 bg-white p-6 flex flex-col justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 text-sm text-gray-500">
                    <FolderIcon className="h-4 w-4" />
                    <span>{article.category?.name || '未分类'}</span>
                    {article.is_ai_generated && (
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                        <SparklesIcon className="h-3 w-3 mr-1" />
                        AI生成
                      </span>
                    )}
                  </div>
                  <Link href={`/articles/${article.id}`} className="block mt-2">
                    <p className="text-xl font-semibold text-gray-900 hover:text-primary-600">
                      {article.title}
                    </p>
                    <p className="mt-3 text-base text-gray-500 line-clamp-3">
                      {article.summary || article.content.substring(0, 150) + '...'}
                    </p>
                  </Link>
                </div>
                <div className="mt-6 flex items-center">
                  <div className="flex items-center space-x-4 text-sm text-gray-500">
                    <div className="flex items-center">
                      <UserIcon className="h-4 w-4 mr-1" />
                      {article.author?.username || '匿名'}
                    </div>
                    <div className="flex items-center">
                      <EyeIcon className="h-4 w-4 mr-1" />
                      {article.view_count || 0}
                    </div>
                    <div className="flex items-center">
                      <CalendarIcon className="h-4 w-4 mr-1" />
                      {new Date(article.created_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              </div>
            </article>
          ))}
        </div>

        {featuredArticles.length === 0 && (
          <div className="text-center py-12">
            <DocumentTextIcon className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">暂无攻略</h3>
            <p className="mt-1 text-sm text-gray-500">还没有发布任何攻略内容</p>
          </div>
        )}
      </div>

      {/* 攻略分类 */}
      {categories.length > 0 && (
        <div className="bg-gray-100">
          <div className="max-w-7xl mx-auto px-4 py-12 sm:px-6 lg:px-8">
            <div className="text-center">
              <h2 className="text-3xl font-extrabold text-gray-900 sm:text-4xl">
                📂 攻略分类
              </h2>
              <p className="mt-3 max-w-2xl mx-auto text-xl text-gray-500 sm:mt-4">
                按分类浏览攻略内容
              </p>
            </div>

            <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
              {categories.map((category) => (
                <Link
                  key={category.id}
                  href={`/categories/${category.id}`}
                  className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
                >
                  <div className="flex items-center">
                    <FolderIcon className="h-8 w-8 text-primary-600" />
                    <div className="ml-4">
                      <h3 className="text-lg font-medium text-gray-900">
                        {category.name}
                      </h3>
                      <p className="text-sm text-gray-500">
                        {category.article_count || 0} 篇攻略
                      </p>
                    </div>
                  </div>
                  {category.description && (
                    <p className="mt-4 text-sm text-gray-600">
                      {category.description}
                    </p>
                  )}
                </Link>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* 最新攻略 */}
      <div className="max-w-7xl mx-auto px-4 py-12 sm:px-6 lg:px-8">
        <div className="text-center">
          <h2 className="text-3xl font-extrabold text-gray-900 sm:text-4xl">
            🆕 最新攻略
          </h2>
          <p className="mt-3 max-w-2xl mx-auto text-xl text-gray-500 sm:mt-4">
            最近发布的攻略内容
          </p>
        </div>

        <div className="mt-12 grid gap-6 lg:grid-cols-2">
          {recentArticles.map((article) => (
            <div
              key={article.id}
              className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow"
            >
              <div className="flex items-start space-x-4">
                <DocumentTextIcon className="h-6 w-6 text-gray-400 mt-1" />
                <div className="flex-1 min-w-0">
                  <Link href={`/articles/${article.id}`}>
                    <h3 className="text-lg font-medium text-gray-900 hover:text-primary-600 line-clamp-2">
                      {article.title}
                    </h3>
                  </Link>
                  <div className="mt-2 flex items-center space-x-4 text-sm text-gray-500">
                    <span>{article.category?.name || '未分类'}</span>
                    <span>{article.author?.username || '匿名'}</span>
                    <span>{new Date(article.created_at).toLocaleDateString()}</span>
                    {article.is_ai_generated && (
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800">
                        AI
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="mt-10 text-center">
          <Link
            href="/articles"
            className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700"
          >
            查看更多攻略
          </Link>
        </div>
      </div>
    </UserLayout>
  );
}